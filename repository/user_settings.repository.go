package repository

import (
	"context"
	"errors"
	"fmt"

	"linkup/models"

	"gorm.io/gorm"
)

type UserSettingsRepository struct {
	db *gorm.DB
}

func NewUserSettingsRepository(db *gorm.DB) *UserSettingsRepository {
	return &UserSettingsRepository{db: db}
}

// GetByUserID returns the user settings, or (nil, nil) when no row exists
// (caller should fall back to defaults).
func (r *UserSettingsRepository) GetByUserID(ctx context.Context, userID string) (*models.UserSetting, error) {
	var setting models.UserSetting
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user settings: %w", err)
	}
	return &setting, nil
}

func (r *UserSettingsRepository) Upsert(ctx context.Context, setting *models.UserSetting) error {
	tx := r.db.WithContext(ctx).Save(setting)
	if tx.Error != nil {
		return fmt.Errorf("upsert user settings: %w", tx.Error)
	}
	return nil
}

// UpdatePresenceSettings updates only the presence-related fields.
func (r *UserSettingsRepository) UpdatePresenceSettings(ctx context.Context, userID string, activityStatusEnabled bool, lastSeenVisibility string) error {
	tx := r.db.WithContext(ctx).
		Model(&models.UserSetting{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"activity_status_enabled": activityStatusEnabled,
			"last_seen_visibility":    lastSeenVisibility,
		})
	if tx.Error != nil {
		return fmt.Errorf("update presence settings: %w", tx.Error)
	}
	return nil
}
