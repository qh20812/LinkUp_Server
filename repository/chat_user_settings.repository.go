package repository

import (
	"context"
	"errors"
	"fmt"
	"linkup/models"
	"time"

	"gorm.io/gorm"
)

type ChatUserSettingsRepository struct {
	db *gorm.DB
}

func NewChatUserSettingsRepository(db *gorm.DB) *ChatUserSettingsRepository {
	return &ChatUserSettingsRepository{db: db}
}

func (r *ChatUserSettingsRepository) GetByChatAndUser(ctx context.Context, chatID, userID string) (*models.ChatUserSettings, error) {
	var settings models.ChatUserSettings
	err := r.db.WithContext(ctx).Where("chat_id = ? AND user_id = ?", chatID, userID).First(&settings).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get chat user settings: %w", err)
	}
	return &settings, nil
}

func (r *ChatUserSettingsRepository) Upsert(ctx context.Context, settings *models.ChatUserSettings) error {
	existing, err := r.GetByChatAndUser(ctx, settings.ChatID, settings.UserID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if existing == nil {
		settings.CreatedAt = now
		settings.UpdatedAt = now
		return r.db.WithContext(ctx).Create(settings).Error
	}
	updates := map[string]any{
		"background_type":  settings.BackgroundType,
		"background_value": settings.BackgroundValue,
		"updated_at":       now,
	}
	return r.db.WithContext(ctx).Model(&models.ChatUserSettings{}).
		Where("chat_id = ? AND user_id = ?", settings.ChatID, settings.UserID).
		Updates(updates).Error
}

func (r *ChatUserSettingsRepository) Delete(ctx context.Context, chatID, userID string) error {
	return r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Delete(&models.ChatUserSettings{}).Error
}
