package repository

import (
	"context"
	"linkup/models"

	"gorm.io/gorm"
)

type MediaRepository interface {
	Create(ctx context.Context, media *models.Media) error
	GetByID(ctx context.Context, id string) (*models.Media, error)
	GetByUserID(ctx context.Context, userID string) ([]models.Media, error)
	UpdateStorageUsage(ctx context.Context, userID string, addedBytes float64) error
	GetUserStorageInfo(ctx context.Context, userID string) (quota, used float64, err error)
}

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(ctx context.Context, media *models.Media) error {
	return r.db.WithContext(ctx).Create(media).Error
}

func (r *mediaRepository) GetByID(ctx context.Context, id string) (*models.Media, error) {
	var media models.Media
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) GetByUserID(ctx context.Context, userID string) ([]models.Media, error) {
	var medias []models.Media
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&medias).Error
	return medias, err
}

func (r *mediaRepository) UpdateStorageUsage(ctx context.Context, userID string, addedBytes float64) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("storage_used_bytes", gorm.Expr("storage_used_bytes + ?", int64(addedBytes))).
		Error
}

func (r *mediaRepository) GetUserStorageInfo(ctx context.Context, userID string) (quota, used float64, err error) {
	var user models.User
	err = r.db.WithContext(ctx).
		Select("storage_quota_bytes", "storage_used_bytes").
		Where("id = ?", userID).
		First(&user).
		Error
	return user.StorageQuotaBytes, user.StorageUsedBytes, err
}
