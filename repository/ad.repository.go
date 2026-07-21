package repository

import (
	"errors"
	"time"

	"linkup/models"

	"gorm.io/gorm"
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
	FindActiveAds(now time.Time) ([]models.Ad, error)
	GetAll() ([]models.Ad, error)
	LogAnalytics(analytics *models.AdAnalytics) error
	GetActionCounts(adID string) (map[string]int64, error)
	GetUniqueReach(adID string) (int64, error)
	CheckAdOwnership(adID, partnerID string) (bool, error)
	TrackActionAndDeduct(analytics *models.AdAnalytics, cpmPrice, cpcPrice float64) error
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
			return nil, errors.New("không tìm thấy quảng cáo")
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
		return errors.New("không tìm thấy quảng cáo để cập nhật")
	}
	return nil
}

func (r *adRepositoryImpl) GetAll() ([]models.Ad, error) {
	var list []models.Ad
	err := r.db.Preload("MediaList").Find(&list).Error
	return list, err
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

// TrackActionAndDeduct lưu log analytics và tính toán trừ ngân sách thực tế (mục 2.4 & 2.5)
func (r *adRepositoryImpl) TrackActionAndDeduct(analytics *models.AdAnalytics, cpmPrice, cpcPrice float64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Lưu log analytics
		if err := tx.Create(analytics).Error; err != nil {
			return err
		}

		// 2. Khóa dòng Ad để tính chi phí
		var ad models.Ad
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&ad, "id = ?", analytics.AdID).Error; err != nil {
			return err
		}

		var cost float64 = 0
		switch analytics.ActionType {
		case string(models.ActionImpression), string(models.ActionView):
			if cpmPrice > 0 {
				cost = cpmPrice / 1000.0
			} else if ad.CPMPrice > 0 {
				cost = ad.CPMPrice / 1000.0
			}
		case string(models.ActionClick):
			if cpcPrice > 0 {
				cost = cpcPrice
			} else if ad.CPCPrice > 0 {
				cost = ad.CPCPrice
			}
		}

		ad.TotalSpent += cost

		// 3. Tự ngưng quảng cáo khi chạm Budget (mục 2.5)[cite: 2]
		if ad.Budget > 0 && ad.TotalSpent >= ad.Budget {
			ad.Status = models.AdStatusCompleted
		}

		return tx.Save(&ad).Error
	})
}
