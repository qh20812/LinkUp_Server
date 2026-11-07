package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ActionCount struct {
	ActionType string `gorm:"column:action_type"`
	Count      int64  `gorm:"column:count"`
}

type AdRepository interface {
	Create(ad *models.Ad) error
	CreateMediaBatch(mediaList []models.AdMedia) error
	FindByID(id string) (*models.Ad, error)
	Update(ad *models.Ad) error
	Delete(ctx context.Context, id string) error
	// DeleteWithMedia xoá media + ad trong 1 transaction, tránh row mồ côi ở
	// ad_media (bảng này không có FK nên DB không tự cascade).
	DeleteWithMedia(ctx context.Context, id string) error
	FindActiveAds(now time.Time) ([]models.Ad, error)
	GetAll() ([]models.Ad, error)
	LogAnalytics(analytics *models.AdAnalytics) error
	GetCountsByAction(adID string) (impressions, clicks, interactions int64, err error)
	GetActionCounts(adID string) (map[string]int64, error)
	GetUniqueReach(adID string) (int64, error)
	CheckAdOwnership(adID, partnerID string) (bool, error)
	UpdateAdFields(ad *models.Ad, fields map[string]interface{}) error
	DeleteAdMedia(adID string) error
	TrackActionAndDeduct(analytics *models.AdAnalytics) (bool, error)
	ListAds(ctx context.Context, keyword, status string, limit, offset int) ([]dto.AdminAdListItem, error)
	CountAds(ctx context.Context, keyword, status string) (int64, error)
	ListByPartnerID(partnerID, keyword, status string, limit, offset int) ([]dto.PartnerAdListItem, error)
	CountByPartnerID(partnerID, keyword, status string) (int64, error)
	FindRankedAdsForUser(now time.Time, gender string, age int, location string, limit int) ([]models.Ad, error)
	GetUserAdViewCount(adID, userID string, since time.Time) (int, error)
	FilterByFrequencyCap(ads []models.Ad, userID string, limit int) ([]models.Ad, error)
	GetDailySpend(adID string, date time.Time) (float64, error)
	HasRecentAction(adID, userID, actionType string, within time.Duration) (bool, error)
	GetOverviewByPartner(partnerID string) (*dto.AdOverviewResponse, error)
	// MarkFinishedAdsCompleted chuyển các ad active đã hết hạn / hết budget /
	// đạt max_impressions sang completed. Trả số dòng đã cập nhật.
	MarkFinishedAdsCompleted(now time.Time) (int64, error)
}

type adRepositoryImpl struct {
	db *gorm.DB
}

func NewAdRepository(db *gorm.DB) AdRepository {
	return &adRepositoryImpl{
		db: db,
	}
}

func (r *adRepositoryImpl) Create(ad *models.Ad) error {
	return r.db.Create(ad).Error
}

func (r *adRepositoryImpl) CreateMediaBatch(mediaList []models.AdMedia) error {
	if len(mediaList) == 0 {
		return nil
	}
	return r.db.Create(&mediaList).Error
}

func (r *adRepositoryImpl) FindByID(id string) (*models.Ad, error) {
	var ad models.Ad
	err := r.db.Preload("MediaList").Where("id = ?", id).First(&ad).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorsapp.New(errorsapp.ErrCodeAdNotFound)
		}
		return nil, err
	}
	return &ad, nil
}

func (r *adRepositoryImpl) Update(ad *models.Ad) error {
	result := r.db.Save(ad)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errorsapp.New(errorsapp.ErrCodeAdNotUpdated)
	}
	return nil
}

func (r *adRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Ad{})
	if result.Error != nil {
		return fmt.Errorf("delete ad: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errorsapp.New(errorsapp.ErrCodeAdNotDeleted)
	}
	return nil
}

// DeleteWithMedia xoá toàn bộ media của ad rồi xoá ad trong cùng transaction.
func (r *adRepositoryImpl) DeleteWithMedia(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ad_id = ?", id).Delete(&models.AdMedia{}).Error; err != nil {
			return fmt.Errorf("delete ad media: %w", err)
		}
		result := tx.Where("id = ?", id).Delete(&models.Ad{})
		if result.Error != nil {
			return fmt.Errorf("delete ad: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return errorsapp.New(errorsapp.ErrCodeAdNotDeleted)
		}
		return nil
	})
}

func (r *adRepositoryImpl) GetAll() ([]models.Ad, error) {
	var list []models.Ad
	err := r.db.Preload("MediaList").Find(&list).Error
	return list, err
}

// ListByPartnerID trả về danh sách ads của riêng partner với metrics, có lọc
// keyword/status và phân trang (mục 6.4).
func (r *adRepositoryImpl) ListByPartnerID(partnerID, keyword, status string, limit, offset int) ([]dto.PartnerAdListItem, error) {
	var items []dto.PartnerAdListItem

	query := `
		SELECT a.id, a.title, a.content, a.format, a.target_url, a.status,
		       a.budget, a.daily_budget, a.cpm_price, a.cpc_price, a.max_impressions,
		       a.total_spent, a.target_gender, a.target_age_min, a.target_age_max,
		       COALESCE(NULLIF(a.target_locations, ''), JSON_ARRAY()) AS target_locations,
		       a.started_at, a.expires_at, a.created_at, a.rejection_reason, a.suspended_reason,
		       COALESCE((SELECT am.url FROM ad_media am WHERE am.ad_id = a.id ORDER BY am.sort_order, am.created_at LIMIT 1), '') AS media_uri,
		       (SELECT COUNT(*) FROM ad_analytics aa WHERE aa.ad_id = a.id AND aa.action_type = 'impression') AS impressions,
		       (SELECT COUNT(*) FROM ad_analytics aa WHERE aa.ad_id = a.id AND aa.action_type = 'click') AS clicks,
		       ROUND(
		           IFNULL((SELECT COUNT(*) FROM ad_analytics aa WHERE aa.ad_id = a.id AND aa.action_type = 'click') * 100.0 /
		           NULLIF((SELECT COUNT(*) FROM ad_analytics aa WHERE aa.ad_id = a.id AND aa.action_type = 'impression'), 0), 0), 2) AS ctr
		FROM ads a
		WHERE a.partner_id = ?`

	args := []interface{}{partnerID}
	query, args = appendAdFilters(query, args, keyword, status)
	query += ` ORDER BY a.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	if err := r.db.Raw(query, args...).Scan(&items).Error; err != nil {
		return nil, fmt.Errorf("list partner ads: %w", err)
	}
	return items, nil
}

// CountByPartnerID đếm tổng ads của riêng partner, cùng bộ lọc với danh sách.
func (r *adRepositoryImpl) CountByPartnerID(partnerID, keyword, status string) (int64, error) {
	var total int64
	query := `SELECT COUNT(*) FROM ads a WHERE a.partner_id = ?`
	args := []interface{}{partnerID}
	query, args = appendAdFilters(query, args, keyword, status)

	if err := r.db.Raw(query, args...).Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("count partner ads: %w", err)
	}
	return total, nil
}

// appendAdFilters thêm điều kiện keyword/status dùng chung cho list + count.
func appendAdFilters(query string, args []interface{}, keyword, status string) (string, []interface{}) {
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query += ` AND (a.title LIKE ? OR a.content LIKE ?)`
		args = append(args, like, like)
	}
	if status = strings.TrimSpace(strings.ToLower(status)); status != "" {
		query += ` AND a.status = ?`
		args = append(args, status)
	}
	return query, args
}

func (r *adRepositoryImpl) FindActiveAds(now time.Time) ([]models.Ad, error) {
	var activeAds []models.Ad
	err := r.db.Preload("MediaList").Where(
		"status = ? AND total_spent < budget AND (started_at IS NULL OR started_at <= ?) AND (expires_at IS NULL OR expires_at >= ?)",
		models.AdStatusActive, now, now,
	).Find(&activeAds).Error

	return activeAds, err
}

func (r *adRepositoryImpl) LogAnalytics(analytics *models.AdAnalytics) error {
	return r.db.Create(analytics).Error
}

func (r *adRepositoryImpl) GetCountsByAction(adID string) (impressions, clicks, interactions int64, err error) {
	if err = r.db.Model(&models.AdAnalytics{}).Where("ad_id = ? AND LOWER(action_type) = ?", adID, "impression").Count(&impressions).Error; err != nil {
		return 0, 0, 0, err
	}
	if err = r.db.Model(&models.AdAnalytics{}).Where("ad_id = ? AND LOWER(action_type) = ?", adID, "click").Count(&clicks).Error; err != nil {
		return 0, 0, 0, err
	}
	if err = r.db.Model(&models.AdAnalytics{}).Where("ad_id = ? AND LOWER(action_type) = ?", adID, "interact").Count(&interactions).Error; err != nil {
		return 0, 0, 0, err
	}
	return impressions, clicks, interactions, nil
}

func (r *adRepositoryImpl) GetActionCounts(adID string) (map[string]int64, error) {
	var results []ActionCount
	err := r.db.Model(&models.AdAnalytics{}).
		Select("LOWER(action_type) as action_type, COUNT(*) as count").
		Where("ad_id = ?", adID).
		Group("LOWER(action_type)").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, res := range results {
		counts[res.ActionType] = res.Count
	}
	return counts, nil
}

func (r *adRepositoryImpl) GetUniqueReach(adID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.AdAnalytics{}).
		Where("ad_id = ? AND user_id IS NOT NULL", adID).
		Distinct("user_id").
		Count(&count).Error

	return count, err
}

// CheckAdOwnership kiểm tra Ad có thuộc về Partner hay không (mục 2.2)
func (r *adRepositoryImpl) CheckAdOwnership(adID, partnerID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Ad{}).Where("id = ? AND partner_id = ?", adID, partnerID).Count(&count).Error
	return count > 0, err
}

// UpdateAdFields cập nhật các field cụ thể của ad (partial update)
func (r *adRepositoryImpl) UpdateAdFields(ad *models.Ad, fields map[string]interface{}) error {
	return r.db.Model(ad).Updates(fields).Error
}

// DeleteAdMedia xóa tất cả media của ad
func (r *adRepositoryImpl) DeleteAdMedia(adID string) error {
	return r.db.Where("ad_id = ?", adID).Delete(&models.AdMedia{}).Error
}

// ActionDedupeWindow là cửa sổ chống trùng một action/ad/user. P1-3 sẽ tách
// riêng theo action type (impression/click); hiện giữ 24h như hành vi cũ.
const ActionDedupeWindow = 24 * time.Hour

// TrackActionAndDeduct ghi nhận tương tác và trừ ngân sách trong MỘT transaction
// có khoá dòng ad (SELECT ... FOR UPDATE). Mọi điều kiện chặn (chống trùng,
// daily budget, budget còn lại, max_impressions) đều kiểm tra bên trong
// transaction để nhiều request đồng thời không vượt ngân sách.
// Trả recorded=false khi bị chặn (không phải lỗi).
func (r *adRepositoryImpl) TrackActionAndDeduct(analytics *models.AdAnalytics) (bool, error) {
	recorded := false

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Khoá dòng ad: GORM v1 dùng Set("gorm:query_option") đã bị bỏ ở v2
		//    nên trước đây không hề khoá.
		var ad models.Ad
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&ad, "id = ?", analytics.AdID).Error; err != nil {
			return err
		}
		if ad.Status != models.AdStatusActive {
			return nil
		}

		// 2. Chống trùng theo action/ad/user.
		if analytics.UserID != nil && *analytics.UserID != "" {
			var recent int64
			since := time.Now().Add(-ActionDedupeWindow)
			if err := tx.Model(&models.AdAnalytics{}).
				Where("ad_id = ? AND user_id = ? AND action_type = ? AND created_at > ?",
					analytics.AdID, *analytics.UserID, analytics.ActionType, since).
				Count(&recent).Error; err != nil {
				return err
			}
			if recent > 0 {
				return nil
			}
		}

		// 3. Đếm impression hiện có (phục vụ max_impressions).
		var impressions int64
		if err := tx.Model(&models.AdAnalytics{}).
			Where("ad_id = ? AND action_type = ?", analytics.AdID, string(models.ActionImpression)).
			Count(&impressions).Error; err != nil {
			return err
		}
		if ad.MaxImpressions > 0 && impressions >= int64(ad.MaxImpressions) {
			ad.Status = models.AdStatusCompleted
			return tx.Model(&models.Ad{}).Where("id = ?", ad.ID).Update("status", ad.Status).Error
		}

		cost := actionCost(&ad, analytics.ActionType)
		if cost > 0 {
			// 4. Budget còn lại: chặn trước khi ghi để không vượt ngân sách.
			if ad.Budget > 0 && ad.TotalSpent+cost > ad.Budget {
				return nil
			}
			// 5. Daily budget.
			if ad.DailyBudget > 0 {
				start := startOfDay(time.Now())
				var daily float64
				if err := tx.Model(&models.AdAnalytics{}).
					Where("ad_id = ? AND created_at >= ? AND created_at < ?", ad.ID, start, start.Add(24*time.Hour)).
					Select("COALESCE(SUM(cost), 0)").Scan(&daily).Error; err != nil {
					return err
				}
				if daily+cost > ad.DailyBudget {
					return nil
				}
			}
		}

		// 6. Ghi log tương tác + cost của chính hành động này.
		if err := tx.Create(analytics).Error; err != nil {
			return err
		}
		if cost > 0 {
			if err := tx.Model(&models.AdAnalytics{}).Where("id = ?", analytics.ID).
				Update("cost", cost).Error; err != nil {
				return err
			}
		}

		// 7. Cập nhật chi tiêu; tự hoàn thành khi chạm budget hoặc max impressions.
		ad.TotalSpent += cost
		reachedMax := ad.MaxImpressions > 0 &&
			analytics.ActionType == string(models.ActionImpression) &&
			impressions+1 >= int64(ad.MaxImpressions)
		if (ad.Budget > 0 && ad.TotalSpent >= ad.Budget) || reachedMax {
			ad.Status = models.AdStatusCompleted
		}
		if err := tx.Model(&models.Ad{}).Where("id = ?", ad.ID).
			Updates(map[string]interface{}{
				"total_spent": ad.TotalSpent,
				"status":      ad.Status,
			}).Error; err != nil {
			return err
		}

		recorded = true
		return nil
	})

	return recorded, err
}

func actionCost(ad *models.Ad, actionType string) float64 {
	switch actionType {
	case string(models.ActionImpression), string(models.ActionView):
		return ad.CPMPrice / 1000.0
	case string(models.ActionClick):
		return ad.CPCPrice
	}
	return 0
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func (r *adRepositoryImpl) ListAds(ctx context.Context, keyword, status string, limit, offset int) ([]dto.AdminAdListItem, error) {
	var ads []dto.AdminAdListItem

	query := `
		SELECT a.id, a.title, a.content, a.partner_id, a.target_url, a.status, a.budget,
		       a.started_at, a.expires_at, a.created_at, a.rejection_reason, a.suspended_reason,
		       u.username AS partner_name,
		       COALESCE(p.display_name, u.username) AS partner_display_name,
		       COALESCE((SELECT am.url FROM ad_media am WHERE am.ad_id = a.id ORDER BY am.sort_order, am.created_at LIMIT 1), '') AS media_uri,
		       (SELECT COUNT(*) FROM ad_analytics aa WHERE aa.ad_id = a.id AND aa.action_type = 'impression') AS impressions,
		       (SELECT COUNT(*) FROM ad_analytics aa WHERE aa.ad_id = a.id AND aa.action_type = 'click') AS clicks,
		       ROUND(
		           IFNULL((SELECT COUNT(*) FROM ad_analytics aa WHERE aa.ad_id = a.id AND aa.action_type = 'click') * 100.0 /
		           NULLIF((SELECT COUNT(*) FROM ad_analytics aa WHERE aa.ad_id = a.id AND aa.action_type = 'impression'), 0), 0), 2) AS ctr
		FROM ads a
		JOIN users u ON u.id = a.partner_id
		LEFT JOIN profiles p ON p.user_id = a.partner_id
		WHERE 1=1`

	var args []interface{}

	if keyword != "" {
		query += ` AND (a.title LIKE ? OR u.username LIKE ? OR p.display_name LIKE ?)`
		like := "%" + keyword + "%"
		args = append(args, like, like, like)
	}
	if status != "" {
		query += ` AND a.status = ?`
		args = append(args, status)
	}

	query += ` ORDER BY a.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&ads).Error; err != nil {
		return nil, fmt.Errorf("list ads: %w", err)
	}
	return ads, nil
}

func (r *adRepositoryImpl) CountAds(ctx context.Context, keyword, status string) (int64, error) {
	var total int64
	query := `SELECT COUNT(*) FROM ads a
		JOIN users u ON u.id = a.partner_id
		LEFT JOIN profiles p ON p.user_id = a.partner_id
		WHERE 1=1`
	var args []interface{}

	if keyword != "" {
		query += ` AND (a.title LIKE ? OR u.username LIKE ? OR p.display_name LIKE ?)`
		like := "%" + keyword + "%"
		args = append(args, like, like, like)
	}
	if status != "" {
		query += ` AND a.status = ?`
		args = append(args, status)
	}

	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("count ads: %w", err)
	}
	return total, nil
}

// FindRankedAdsForUser trả về ads active được sắp xếp theo điểm ranking (mục 4.2)
func (r *adRepositoryImpl) FindRankedAdsForUser(now time.Time, gender string, age int, location string, limit int) ([]models.Ad, error) {
	query := `
		SELECT a.*,
			(
				CASE
					WHEN a.target_gender = 'all' OR a.target_gender = ? THEN 0.4
					ELSE 0.0
				END
				+
				CASE
					WHEN a.target_age_min = 0 AND a.target_age_max = 100 THEN 0.0
					WHEN ? BETWEEN a.target_age_min AND a.target_age_max THEN 0.4
					ELSE 0.0
				END
				+
				(0.2 / (1.0 + DATEDIFF(NOW(), a.created_at)))
				+
				(0.15 * (a.budget - a.total_spent) / GREATEST(a.budget, 1))
				+
				(0.15 * LEAST(
					(SELECT IFNULL(COUNT(CASE WHEN aa.action_type='click' THEN 1 END), 0) * 100.0 /
					 NULLIF(COUNT(CASE WHEN aa.action_type='impression' THEN 1 END), 0)
					 FROM ad_analytics aa WHERE aa.ad_id = a.id) / 10.0,
					1.0
				))
				+
				(0.10 * RAND())
			) AS rank_score
		FROM ads a
		WHERE a.status = 'active'
		  AND a.total_spent < a.budget
		  AND (a.started_at IS NULL OR a.started_at <= ?)
		  AND (a.expires_at IS NULL OR a.expires_at >= ?)
		  AND (a.max_impressions = 0 OR a.max_impressions > (
		      SELECT COUNT(*) FROM ad_analytics x
		      WHERE x.ad_id = a.id AND x.action_type = 'impression'))
		  AND (a.package_id IS NULL OR EXISTS (
		      SELECT 1 FROM partner_subscriptions ps
		      WHERE ps.user_id = a.partner_id
		        AND ps.status = 'active'
		        AND ps.expires_at > ?))
		  AND (a.target_gender = 'all' OR a.target_gender = ?)
		  AND (a.target_age_min = 0 OR a.target_age_min <= ?)
		  AND (a.target_age_max = 100 OR a.target_age_max >= ?)
		  AND (JSON_LENGTH(a.target_locations) = 0
		       OR JSON_CONTAINS(a.target_locations, JSON_QUOTE(?), '$'))
		ORDER BY rank_score DESC
		LIMIT ?`

	var ads []models.Ad
	err := r.db.Raw(query, gender, age, now, now, now, gender, age, age, location, limit).Scan(&ads).Error
	if err != nil {
		return nil, err
	}
	if len(ads) == 0 {
		return ads, nil
	}

	// Gắn media để quảng cáo trên feed có ảnh. Query gốc dùng SELECT a.* nên
	// media_list luôn rỗng — trước đây card quảng cáo trong feed không có ảnh.
	adIDs := make([]string, 0, len(ads))
	for _, ad := range ads {
		adIDs = append(adIDs, ad.ID)
	}

	var media []models.AdMedia
	if err := r.db.Where("ad_id IN ?", adIDs).
		Order("sort_order ASC, created_at ASC").
		Find(&media).Error; err != nil {
		return nil, err
	}

	mediaByAd := make(map[string][]models.AdMedia, len(ads))
	for _, m := range media {
		mediaByAd[m.AdID] = append(mediaByAd[m.AdID], m)
	}
	for i := range ads {
		ads[i].MediaList = mediaByAd[ads[i].ID]
	}

	return ads, nil
}

// GetUserAdViewCount đếm số lần user đã thấy ad trong khoảng thời gian (mục 4.3)
func (r *adRepositoryImpl) GetUserAdViewCount(adID, userID string, since time.Time) (int, error) {
	var count int64
	err := r.db.Model(&models.AdAnalytics{}).
		Where("ad_id = ? AND user_id = ? AND action_type = 'impression' AND created_at > ?",
			adID, userID, since).
		Count(&count).Error
	return int(count), err
}

// FilterByFrequencyCap lọc ads theo giới hạn hiển thị mỗi user mỗi ngày (mục 4.3)
const MaxImpressionsPerUserPerDay = 3

func (r *adRepositoryImpl) FilterByFrequencyCap(ads []models.Ad, userID string, limit int) ([]models.Ad, error) {
	if userID == "" {
		if len(ads) > limit {
			return ads[:limit], nil
		}
		return ads, nil
	}
	startOfDay := time.Now().Truncate(24 * time.Hour)
	var filtered []models.Ad
	for _, ad := range ads {
		if len(filtered) >= limit {
			break
		}
		count, err := r.GetUserAdViewCount(ad.ID, userID, startOfDay)
		if err != nil {
			continue
		}
		if count < MaxImpressionsPerUserPerDay {
			filtered = append(filtered, ad)
		}
	}
	return filtered, nil
}

// GetDailySpend lấy tổng chi phí trong ngày (mục 2.4)
func (r *adRepositoryImpl) GetDailySpend(adID string, date time.Time) (float64, error) {
	var total float64
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)
	err := r.db.Model(&models.AdAnalytics{}).
		Where("ad_id = ? AND created_at >= ? AND created_at < ?", adID, start, end).
		Select("COALESCE(SUM(cost), 0)").Scan(&total).Error
	return total, err
}

// HasRecentAction kiểm tra user có đã thực hiện action này gần đây không (mục 2.2)
func (r *adRepositoryImpl) HasRecentAction(adID, userID, actionType string, within time.Duration) (bool, error) {
	var count int64
	since := time.Now().Add(-within)
	err := r.db.Model(&models.AdAnalytics{}).
		Where("ad_id = ? AND user_id = ? AND action_type = ? AND created_at > ?",
			adID, userID, actionType, since).
		Count(&count).Error
	return count > 0, err
}

func (r *adRepositoryImpl) MarkFinishedAdsCompleted(now time.Time) (int64, error) {
	res := r.db.Model(&models.Ad{}).
		Where("status = ?", models.AdStatusActive).
		Where(`(expires_at IS NOT NULL AND expires_at < ?)
		    OR (budget > 0 AND total_spent >= budget)
		    OR (max_impressions > 0 AND (
		        SELECT COUNT(*) FROM ad_analytics x
		        WHERE x.ad_id = ads.id AND x.action_type = 'impression') >= max_impressions)`, now).
		Update("status", models.AdStatusCompleted)
	return res.RowsAffected, res.Error
}

// GetOverviewByPartner lấy tổng quan quảng cáo của partner (mục 6.1)
func (r *adRepositoryImpl) GetOverviewByPartner(partnerID string) (*dto.AdOverviewResponse, error) {
	overview := &dto.AdOverviewResponse{}

	// Tổng budget, spent từ ads
	type budgetSummary struct {
		TotalBudget float64
		TotalSpent  float64
		ActiveAds   int64
	}
	var bs budgetSummary
	err := r.db.Raw(`
		SELECT
			COALESCE(SUM(budget), 0) AS total_budget,
			COALESCE(SUM(total_spent), 0) AS total_spent,
			SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) AS active_ads
		FROM ads WHERE partner_id = ?`, partnerID).Scan(&bs).Error
	if err != nil {
		return nil, err
	}
	overview.TotalBudget = bs.TotalBudget
	overview.TotalSpent = bs.TotalSpent
	overview.RemainingBudget = bs.TotalBudget - bs.TotalSpent
	if overview.RemainingBudget < 0 {
		overview.RemainingBudget = 0
	}
	overview.ActiveAds = int(bs.ActiveAds)

	// Impressions + Clicks
	type metricsSummary struct {
		Impressions int64
		Clicks      int64
	}
	var ms metricsSummary
	err = r.db.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN aa.action_type = 'impression' THEN 1 ELSE 0 END), 0) AS impressions,
			COALESCE(SUM(CASE WHEN aa.action_type = 'click' THEN 1 ELSE 0 END), 0) AS clicks
		FROM ad_analytics aa
		JOIN ads a ON a.id = aa.ad_id
		WHERE a.partner_id = ?`, partnerID).Scan(&ms).Error
	if err != nil {
		return nil, err
	}
	overview.Impressions = ms.Impressions
	overview.Clicks = ms.Clicks
	if overview.Impressions > 0 {
		overview.CTR = (float64(overview.Clicks) / float64(overview.Impressions)) * 100.0
	}

	// Subscription info
	type subInfo struct {
		SlotsUsed    int
		MaxSlots     int
		PackageName  string
		ExpiresAt    *time.Time
	}
	var si subInfo
	err = r.db.Raw(`
		SELECT
			COALESCE(ps.slots_used, 0) AS slots_used,
			COALESCE(p.max_slots, 0) AS max_slots,
			COALESCE(p.name, '') AS package_name,
			ps.expires_at
		FROM partner_subscriptions ps
		JOIN ad_packages p ON p.id = ps.package_id
		WHERE ps.user_id = ? AND ps.status = 'active'
		ORDER BY ps.expires_at DESC LIMIT 1`, partnerID).Scan(&si).Error
	if err == nil {
		overview.SlotsUsed = si.SlotsUsed
		overview.MaxSlots = si.MaxSlots
		overview.SubscriptionName = si.PackageName
		overview.ExpiresAt = si.ExpiresAt
	}

	return overview, nil
}
