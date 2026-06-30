package repository

import (
	"context"
	"linkup/models"

	"gorm.io/gorm"
)

type MediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) Create(ctx context.Context, media *models.Media) error {
	return r.db.WithContext(ctx).Create(media).Error
}

func (r *MediaRepository) GetByID(ctx context.Context, id string) (*models.Media, error) {
	var media models.Media
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *MediaRepository) GetByUserID(ctx context.Context, userID string) ([]models.Media, error) {
	var medias []models.Media
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&medias).Error
	return medias, err
}

func (r *MediaRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Media{}, "id = ?", id).Error
}

func (r *MediaRepository) UpdateStorageUsage(ctx context.Context, userID string, addedBytes float64) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("storage_used_bytes", gorm.Expr("storage_used_bytes + ?", int64(addedBytes))).
		Error
}

func (r *MediaRepository) GetUserStorageInfo(ctx context.Context, userID string) (quota, used float64, err error) {
	var user models.User
	err = r.db.WithContext(ctx).
		Select("storage_quota_bytes", "storage_used_bytes").
		Where("id = ?", userID).
		First(&user).
		Error
	return user.StorageQuotaBytes, user.StorageUsedBytes, err
}

func (r *MediaRepository) DeleteWithStorageAdjustment(ctx context.Context, userID string, media *models.Media) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.Media{}, "id = ?", media.ID).Error; err != nil {
			return err
		}

		return tx.Model(&models.User{}).
			Where("id = ?", userID).
			Update("storage_used_bytes", gorm.Expr("GREATEST(storage_used_bytes - ?, 0)", int64(media.FileSize))).
			Error
	})
}
