package repository

import (
	"context"
	"linkup/models"
	"time"

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

func (r *MediaRepository) UpdateStatus(ctx context.Context, id string, status models.MediaStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.Media{}).
		Where("id = ?", id).
		Update("status", status).
		Error
}

func (r *MediaRepository) GetByStatus(ctx context.Context, status models.MediaStatus, page, pageSize int) ([]models.Media, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var items []models.Media
	var total int64

	base := r.db.WithContext(ctx).Model(&models.Media{}).Where("status = ?", status)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []models.Media{}, 0, nil
	}

	err := base.Session(&gorm.Session{}).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *MediaRepository) GetRejectedOlderThan(ctx context.Context, cutoff time.Time) ([]models.Media, error) {
	var items []models.Media
	err := r.db.WithContext(ctx).
		Model(&models.Media{}).
		Where("status = ? AND created_at < ?", models.MediaStatusRejected, cutoff).
		Find(&items).Error
	return items, err
}

func (r *MediaRepository) ClearFileURI(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.Media{}).
		Where("id = ?", id).
		Update("file_uri", "").
		Error
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
