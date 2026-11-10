package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
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
	GetActionCounts(adID string) (map[string]int64, error)
	GetUniqueReach(adID string) (int64, error)
	CheckAdOwnership(adID, partnerID string) (bool, error)
	UpdateAdFields(ad *models.Ad, fields map[string]interface{}) error
	DeleteAdMedia(adID string) error
	// ListAdMedia/ReplaceAdMedia phục vụ thay toàn bộ media của ad (C2).
	ListAdMedia(adID string) ([]models.AdMedia, error)
	ReplaceAdMedia(adID string, mediaList []models.AdMedia) error
	TrackActionAndDeduct(analytics *models.AdAnalytics) (bool, error)
	ListAds(ctx context.Context, keyword, status string, limit, offset int) ([]dto.AdminAdListItem, error)
	CountAds(ctx context.Context, keyword, status string) (int64, error)
	ListByPartnerID(partnerID, keyword, status string, limit, offset int) ([]dto.PartnerAdListItem, error)
	CountByPartnerID(partnerID, keyword, status string) (int64, error)
	FindRankedAdsForUser(now time.Time, gender string, age int, location string, limit int) ([]models.Ad, error)
	GetUserAdViewCount(adID, userID string, since time.Time) (int, error)
	FilterByFrequencyCap(ads []models.Ad, userID string, limit int) ([]models.Ad, error)
	GetOverviewByPartner(partnerID string) (*dto.AdOverviewResponse, error)
	// GetAdMetricsRange trả số liệu trong khoảng [from,to) kèm chuỗi theo ngày.
	GetAdMetricsRange(adID string, from, to time.Time) (*AdRangeMetrics, error)
	// MarkFinishedAdsCompleted chuyển các ad active đã hết hạn / hết budget /
	// đạt max_impressions sang completed. Trả số dòng đã cập nhật.
	MarkFinishedAdsCompleted(now time.Time) (int64, error)
	// SyncAdBudgetReservation giữ chỗ/trả lại ví để khớp với ngân sách còn lại
	// của ad (target = budget - total_spent). Trả delta reserved (dương = giữ
	// thêm, âm = trả lại). Ad do admin tạo (PackageID NULL) không dùng ví.
	SyncAdBudgetReservation(ctx context.Context, adID string) (int64, error)
	// ReleaseAdBudget trả toàn bộ tiền đang giữ chỗ của ad về ví (dùng khi ad
	// bị từ chối/xoá/kết thúc). Trả số tiền đã trả lại.
	ReleaseAdBudget(ctx context.Context, adID string) (int64, error)
	// ListAdsWithOutstandingReserve liệt kê ad đã kết thúc nhưng còn giữ chỗ,
	// phục vụ worker dọn tiền về ví.
	ListAdsWithOutstandingReserve(ctx context.Context, limit int) ([]models.Ad, error)
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

	// Gắn toàn bộ media để trang edit hiển thị/thay thế đúng cả carousel.
	if len(items) > 0 {
		ids := make([]string, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.ID)
		}
		var media []models.AdMedia
		if err := r.db.Where("ad_id IN ?", ids).
			Order("sort_order ASC, created_at ASC").
			Find(&media).Error; err != nil {
			return nil, fmt.Errorf("load partner ad media: %w", err)
		}
		byAd := make(map[string][]dto.FeedAdMedia, len(ids))
		for _, item := range media {
			byAd[item.AdID] = append(byAd[item.AdID], dto.FeedAdMedia{
				URL:       item.URL,
				MediaType: item.MediaType,
			})
		}
		for i := range items {
			items[i].MediaList = byAd[items[i].ID]
		}
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
		Where("ad_id = ? AND action_type = ? AND user_id IS NOT NULL", adID, string(models.ActionImpression)).
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

// ListAdMedia trả media hiện tại theo đúng thứ tự hiển thị.
func (r *adRepositoryImpl) ListAdMedia(adID string) ([]models.AdMedia, error) {
	var media []models.AdMedia
	err := r.db.Where("ad_id = ?", adID).
		Order("sort_order ASC, created_at ASC").
		Find(&media).Error
	if err != nil {
		return nil, fmt.Errorf("list ad media: %w", err)
	}
	return media, nil
}

// ReplaceAdMedia thay toàn bộ ad_media trong 1 transaction: xoá cũ rồi chèn mới.
func (r *adRepositoryImpl) ReplaceAdMedia(adID string, mediaList []models.AdMedia) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ad_id = ?", adID).Delete(&models.AdMedia{}).Error; err != nil {
			return fmt.Errorf("delete old ad media: %w", err)
		}
		if len(mediaList) == 0 {
			return nil
		}
		if err := tx.Create(&mediaList).Error; err != nil {
			return fmt.Errorf("insert new ad media: %w", err)
		}
		return nil
	})
}

// Cửa sổ chống trùng theo loại hành động: impression 30 phút (tránh đếm lặp
// khi feed re-render/cuộn qua lại), click 5 phút (chống double-click nhưng vẫn
// tính CPC thật). Các action còn lại do client gửi 1 lần/session nên không dedupe.
const (
	impressionDedupeWindow = 30 * time.Minute
	clickDedupeWindow      = 5 * time.Minute
)

func actionDedupeWindow(actionType string) time.Duration {
	switch actionType {
	case string(models.ActionImpression):
		return impressionDedupeWindow
	case string(models.ActionClick):
		return clickDedupeWindow
	default:
		return 0
	}
}

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

		// 2. Chống trùng theo action/ad/user (impression 30', click 5').
		if window := actionDedupeWindow(analytics.ActionType); window > 0 &&
			analytics.UserID != nil && *analytics.UserID != "" {
			var recent int64
			since := time.Now().Add(-window)
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
			if err := releaseAdSlotTx(tx, &ad); err != nil {
				return err
			}
			if err := releaseAdReserveTx(tx, &ad); err != nil {
				return err
			}
			return tx.Model(&models.Ad{}).Where("id = ?", ad.ID).
				Updates(map[string]interface{}{
					"status":          ad.Status,
					"wallet_reserved": ad.WalletReserved,
				}).Error
		}

		cost := actionCost(&ad, analytics.ActionType)
		costInt := int64(0)
		if cost > 0 {
			// Ví dùng số nguyên VND; làm tròn nhưng không để phí về 0.
			costInt = int64(math.Round(cost))
			if costInt < 1 {
				costInt = 1
			}
			charged := float64(costInt)

			// 4. Budget còn lại: chặn trước khi ghi để không vượt ngân sách.
			if ad.Budget > 0 && ad.TotalSpent+charged > ad.Budget {
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
				if daily+charged > ad.DailyBudget {
					return nil
				}
			}
		}

		// 5b. Ad có ngân sách giữ từ ví: khoá ví và kiểm tra còn đủ tiền giữ
		// chỗ. Không đủ thì dừng ad ngay và trả phần còn lại, không ghi nhận
		// tương tác miễn phí.
		var spendWallet *models.Wallet
		if costInt > 0 && ad.WalletReserved > 0 {
			wallet, walletErr := lockWalletTx(tx, ad.PartnerID)
			if walletErr != nil {
				return walletErr
			}
			if wallet.Reserved < costInt || wallet.Balance < costInt {
				ad.Status = models.AdStatusCompleted
				if err := releaseAdSlotTx(tx, &ad); err != nil {
					return err
				}
				if err := releaseAdReserveTx(tx, &ad); err != nil {
					return err
				}
				return tx.Model(&models.Ad{}).Where("id = ?", ad.ID).
					Updates(map[string]interface{}{
						"status":          ad.Status,
						"wallet_reserved": ad.WalletReserved,
					}).Error
			}
			spendWallet = wallet
		}

		// 6. Ghi log tương tác + cost của chính hành động này.
		analytics.Cost = float64(costInt)
		if err := tx.Create(analytics).Error; err != nil {
			return err
		}

		// 6b. Trừ tiền thật khỏi ví (balance và reserved cùng giảm) trong cùng
		// transaction với analytics + ad.
		if spendWallet != nil {
			if err := applyWalletDeltaTx(
				tx, spendWallet,
				-costInt, -costInt, -costInt,
				models.WalletTxSpend, "ad_analytics", analytics.ID,
				"Trừ tiền hiển thị/click quảng cáo",
			); err != nil {
				return err
			}
			ad.WalletReserved -= costInt
			if ad.WalletReserved < 0 {
				ad.WalletReserved = 0
			}
		}

		// 7. Cập nhật chi tiêu; tự hoàn thành khi chạm budget hoặc max impressions.
		ad.TotalSpent += float64(costInt)
		reachedMax := ad.MaxImpressions > 0 &&
			analytics.ActionType == string(models.ActionImpression) &&
			impressions+1 >= int64(ad.MaxImpressions)
		if (ad.Budget > 0 && ad.TotalSpent >= ad.Budget) || reachedMax {
			ad.Status = models.AdStatusCompleted
			// Ad kết thúc: trả slot + phần ngân sách chưa tiêu về ví.
			if err := releaseAdSlotTx(tx, &ad); err != nil {
				return err
			}
			if err := releaseAdReserveTx(tx, &ad); err != nil {
				return err
			}
		}
		if err := tx.Model(&models.Ad{}).Where("id = ?", ad.ID).
			Updates(map[string]interface{}{
				"total_spent":     ad.TotalSpent,
				"status":          ad.Status,
				"wallet_reserved": ad.WalletReserved,
			}).Error; err != nil {
			return err
		}

		recorded = true
		return nil
	})

	return recorded, err
}

// releaseAdSlotTx trả 1 slot cho subscription đang hoạt động của partner khi ad
// chuyển sang trạng thái không chiếm slot (completed). No-op với ad do admin
// tạo hộ (không gắn gói) hoặc partner không còn gói active.
func releaseAdSlotTx(tx *gorm.DB, ad *models.Ad) error {
	if ad.PackageID == nil {
		return nil
	}
	return tx.Exec(`
		UPDATE partner_subscriptions
		SET slots_used = GREATEST(slots_used - 1, 0), updated_at = ?
		WHERE user_id = ? AND status = ?`,
		time.Now(), ad.PartnerID, models.SubscriptionStatusActive).Error
}

// releaseAdReserveTx trả toàn bộ tiền đang giữ chỗ của ad về ví trong transaction
// hiện tại. No-op nếu ad không giữ chỗ đồng nào (ad cũ hoặc ad admin tạo).
func releaseAdReserveTx(tx *gorm.DB, ad *models.Ad) error {
	if ad.WalletReserved <= 0 {
		return nil
	}
	wallet, err := lockWalletTx(tx, ad.PartnerID)
	if err != nil {
		return err
	}
	amount := ad.WalletReserved
	if amount > wallet.Reserved {
		amount = wallet.Reserved
	}
	if amount > 0 {
		if err := applyWalletDeltaTx(
			tx, wallet,
			0, -amount, -amount,
			models.WalletTxRelease, "ad", ad.ID,
			"Trả lại ngân sách chưa tiêu của quảng cáo",
		); err != nil {
			return err
		}
	}
	ad.WalletReserved = 0
	return nil
}

// SyncAdBudgetReservation giữ chỗ thêm hoặc trả lại phần ngân sách chưa tiêu để
// ví khớp với ad. Ad do admin tạo hộ (PackageID NULL) không dùng ví của partner.
func (r *adRepositoryImpl) SyncAdBudgetReservation(ctx context.Context, adID string) (int64, error) {
	var applied int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ad models.Ad
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&ad, "id = ?", adID).Error; err != nil {
			return err
		}
		if ad.PackageID == nil {
			return nil
		}

		target := int64(math.Round(ad.Budget - ad.TotalSpent))
		if target < 0 {
			target = 0
		}
		delta := target - ad.WalletReserved
		if delta == 0 {
			return nil
		}

		wallet, err := lockWalletTx(tx, ad.PartnerID)
		if err != nil {
			return err
		}

		if delta > 0 {
			if wallet.Available() < delta {
				return errorsapp.Newf(errorsapp.ErrCodeWalletInsufficient, map[string]any{
					"required":  delta,
					"available": wallet.Available(),
				})
			}
			if err := applyWalletDeltaTx(
				tx, wallet,
				0, delta, delta,
				models.WalletTxReserve, "ad", ad.ID,
				"Giữ chỗ ngân sách quảng cáo",
			); err != nil {
				return err
			}
		} else {
			amount := -delta
			if amount > ad.WalletReserved {
				amount = ad.WalletReserved
			}
			if amount > wallet.Reserved {
				amount = wallet.Reserved
			}
			if amount > 0 {
				if err := applyWalletDeltaTx(
					tx, wallet,
					0, -amount, -amount,
					models.WalletTxRelease, "ad", ad.ID,
					"Điều chỉnh giảm ngân sách giữ chỗ",
				); err != nil {
					return err
				}
				delta = -amount
			} else {
				delta = 0
			}
		}

		if err := tx.Model(&models.Ad{}).Where("id = ?", ad.ID).
			Update("wallet_reserved", target).Error; err != nil {
			return err
		}
		applied = delta
		return nil
	})

	return applied, err
}

// ReleaseAdBudget trả toàn bộ tiền đang giữ chỗ của ad về ví. Dùng khi ad bị
// từ chối, bị xoá hoặc đã kết thúc (worker/CompleteFinishedAds).
func (r *adRepositoryImpl) ReleaseAdBudget(ctx context.Context, adID string) (int64, error) {
	var released int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ad models.Ad
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&ad, "id = ?", adID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		before := ad.WalletReserved
		if err := releaseAdReserveTx(tx, &ad); err != nil {
			return err
		}
		if before <= 0 {
			return nil
		}
		released = before
		return tx.Model(&models.Ad{}).Where("id = ?", ad.ID).
			Update("wallet_reserved", 0).Error
	})

	return released, err
}

func (r *adRepositoryImpl) ListAdsWithOutstandingReserve(ctx context.Context, limit int) ([]models.Ad, error) {
	if limit <= 0 {
		limit = 200
	}
	var ads []models.Ad
	err := r.db.WithContext(ctx).
		Where("wallet_reserved > 0 AND status IN ?", []models.AdStatus{
			models.AdStatusCompleted, models.AdStatusRejected,
		}).
		Limit(limit).
		Find(&ads).Error
	return ads, err
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
	// Đầu ngày theo timezone server (trước đây Truncate(24h) tính theo UTC nên
	// mốc ngày bị lệch 7 tiếng với giờ Việt Nam).
	dayStart := startOfDay(time.Now())
	var filtered []models.Ad
	for _, ad := range ads {
		if len(filtered) >= limit {
			break
		}
		count, err := r.GetUserAdViewCount(ad.ID, userID, dayStart)
		if err != nil {
			continue
		}
		if count < MaxImpressionsPerUserPerDay {
			filtered = append(filtered, ad)
		}
	}
	return filtered, nil
}

// MarkFinishedAdsCompleted chuyển các ad active đã hết hạn / hết budget /
// đạt max_impressions sang completed VÀ trả slot cho gói trong cùng transaction.
// Chỉ ad đang active mới được xử lý nên chạy lại không trả slot lần hai.
func (r *adRepositoryImpl) MarkFinishedAdsCompleted(now time.Time) (int64, error) {
	var count int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var ads []models.Ad
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("status = ?", models.AdStatusActive).
			Where(`(expires_at IS NOT NULL AND expires_at < ?)
			    OR (budget > 0 AND total_spent >= budget)
			    OR (max_impressions > 0 AND (
			        SELECT COUNT(*) FROM ad_analytics x
			        WHERE x.ad_id = ads.id AND x.action_type = 'impression') >= max_impressions)`, now).
			Find(&ads).Error; err != nil {
			return err
		}

		for i := range ads {
			ad := &ads[i]
			if err := tx.Model(&models.Ad{}).
				Where("id = ? AND status = ?", ad.ID, models.AdStatusActive).
				Update("status", models.AdStatusCompleted).Error; err != nil {
				return err
			}
			if err := releaseAdSlotTx(tx, ad); err != nil {
				return err
			}
			count++
		}
		return nil
	})
	return count, err
}

// AdRangeMetrics là số liệu tổng hợp + chuỗi ngày trong một khoảng thời gian.
type AdRangeMetrics struct {
	Counts      map[string]int64
	Cost        float64
	UniqueReach int64
	Series      []dto.AdDailyMetric
}

// GetAdMetricsRange tổng hợp ad_analytics trong [from, to): đếm theo action,
// tổng cost, unique reach (chỉ tính impression) và chuỗi theo ngày cho biểu đồ.
func (r *adRepositoryImpl) GetAdMetricsRange(adID string, from, to time.Time) (*AdRangeMetrics, error) {
	out := &AdRangeMetrics{Counts: map[string]int64{}}

	type actionRow struct {
		ActionType string  `gorm:"column:action_type"`
		Count      int64   `gorm:"column:count"`
		Cost       float64 `gorm:"column:cost"`
	}
	var rows []actionRow
	if err := r.db.Raw(`
		SELECT LOWER(action_type) AS action_type, COUNT(*) AS count, COALESCE(SUM(cost), 0) AS cost
		FROM ad_analytics
		WHERE ad_id = ? AND created_at >= ? AND created_at < ?
		GROUP BY LOWER(action_type)`, adID, from, to).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("ad metrics range: %w", err)
	}
	for _, row := range rows {
		out.Counts[row.ActionType] = row.Count
		out.Cost += row.Cost
	}

	// Unique reach chỉ tính người đã thực sự thấy quảng cáo (impression).
	if err := r.db.Model(&models.AdAnalytics{}).
		Where("ad_id = ? AND action_type = ? AND user_id IS NOT NULL AND created_at >= ? AND created_at < ?",
			adID, string(models.ActionImpression), from, to).
		Distinct("user_id").
		Count(&out.UniqueReach).Error; err != nil {
		return nil, fmt.Errorf("ad unique reach range: %w", err)
	}

	type dayRow struct {
		Date             string  `gorm:"column:date"`
		Impressions      int64   `gorm:"column:impressions"`
		Clicks           int64   `gorm:"column:clicks"`
		Views            int64   `gorm:"column:views"`
		Swipes           int64   `gorm:"column:swipes"`
		VideoStarts      int64   `gorm:"column:video_starts"`
		VideoCompletions int64   `gorm:"column:video_completions"`
		Cost             float64 `gorm:"column:cost"`
	}
	var days []dayRow
	if err := r.db.Raw(`
		SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS date,
		       SUM(action_type = 'impression') AS impressions,
		       SUM(action_type = 'click') AS clicks,
		       SUM(action_type = 'view') AS views,
		       SUM(action_type = 'swipe') AS swipes,
		       SUM(action_type = 'video_start') AS video_starts,
		       SUM(action_type = 'video_end') AS video_completions,
		       COALESCE(SUM(cost), 0) AS cost
		FROM ad_analytics
		WHERE ad_id = ? AND created_at >= ? AND created_at < ?
		GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')
		ORDER BY date`, adID, from, to).Scan(&days).Error; err != nil {
		return nil, fmt.Errorf("ad daily series: %w", err)
	}
	out.Series = make([]dto.AdDailyMetric, 0, len(days))
	for _, d := range days {
		out.Series = append(out.Series, dto.AdDailyMetric{
			Date:             d.Date,
			Impressions:      d.Impressions,
			Clicks:           d.Clicks,
			Views:            d.Views,
			Swipes:           d.Swipes,
			VideoStarts:      d.VideoStarts,
			VideoCompletions: d.VideoCompletions,
			Cost:             d.Cost,
		})
	}

	return out, nil
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
			COALESCE((
				SELECT COUNT(*) FROM ads a
				WHERE a.partner_id = ps.user_id AND a.status IN (`+slotConsumingStatusSQL+`)
			), 0) AS slots_used,
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
