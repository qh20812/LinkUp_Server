package repository

import (
	"context"
	"errors"
	"fmt"

	"linkup/models"

	"gorm.io/gorm"
)

type NotificationPreferenceRepository struct {
	db *gorm.DB
}

func NewNotificationPreferenceRepository(db *gorm.DB) *NotificationPreferenceRepository {
	return &NotificationPreferenceRepository{db: db}
}

func (r *NotificationPreferenceRepository) GetByUserID(ctx context.Context, userID string) (*models.NotificationPreference, error) {
	var pref models.NotificationPreference
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&pref).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get notification preference: %w", err)
	}
	return &pref, nil
}

func (r *NotificationPreferenceRepository) Upsert(ctx context.Context, pref *models.NotificationPreference) error {
	tx := r.db.WithContext(ctx).Save(pref)
	if tx.Error != nil {
		return fmt.Errorf("upsert notification preference: %w", tx.Error)
	}
	return nil
}
