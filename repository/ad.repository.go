package repository

import (
	"errors"
	"time"

	"linkup/models"

	"gorm.io/gorm"
)

type AdRepository interface {
	Create(ad *models.Ad) error
	FindByID(id string) (*models.Ad, error)
	Update(ad *models.Ad) error
	FindActiveAds(now time.Time) ([]models.Ad, error)
	GetAll() ([]models.Ad, error)
	LogAnalytics(analytics *models.AdAnalytics) error
	GetCountsByAction(adID string) (impressions, clicks, interactions int64, err error)
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

func (r *adRepositoryImpl) FindByID(id string) (*models.Ad, error) {
	var ad models.Ad
	err := r.db.Where("id = ?", id).First(&ad).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("advertisement not found")
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
		return errors.New("advertisement not found to update")
	}
	return nil
}

func (r *adRepositoryImpl) GetAll() ([]models.Ad, error) {
	var list []models.Ad
	err := r.db.Find(&list).Error
	return list, err
}

func (r *adRepositoryImpl) FindActiveAds(now time.Time) ([]models.Ad, error) {
	var activeAds []models.Ad
	err := r.db.Where(
		"status = ? AND (started_at IS NULL OR started_at <= ?) AND (expires_at IS NULL OR expires_at >= ?)",
		models.AdStatusActive, now, now,
	).Find(&activeAds).Error

	return activeAds, err
}

func (r *adRepositoryImpl) LogAnalytics(analytics *models.AdAnalytics) error {
	return r.db.Create(analytics).Error
}

func (r *adRepositoryImpl) GetCountsByAction(adID string) (impressions, clicks, interactions int64, err error) {
	// Đếm số lượng lượt hiển thị (impression)
	if err = r.db.Model(&models.AdAnalytics{}).Where("ad_id = ? AND LOWER(action_type) = ?", adID, "impression").Count(&impressions).Error; err != nil {
		return 0, 0, 0, err
	}
	// Đếm số lượng lượt click
	if err = r.db.Model(&models.AdAnalytics{}).Where("ad_id = ? AND LOWER(action_type) = ?", adID, "click").Count(&clicks).Error; err != nil {
		return 0, 0, 0, err
	}
	// Đếm số lượng lượt tương tác (interact)
	if err = r.db.Model(&models.AdAnalytics{}).Where("ad_id = ? AND LOWER(action_type) = ?", adID, "interact").Count(&interactions).Error; err != nil {
		return 0, 0, 0, err
	}
	return impressions, clicks, interactions, nil
}
