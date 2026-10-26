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
		"overlay_strength": settings.OverlayStrength,
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

// HideChat ẩn hội thoại ở phía một người dùng (xóa cục bộ). Không đụng tới
// messages/chat_participants/chats nên đối phương giữ nguyên cuộc trò chuyện và
// khóa E2E vẫn dùng lại được khi hội thoại xuất hiện lại.
//
// Upsert tối thiểu (chỉ hidden_at) để không ghi đè background_type /
// background_value / overlay_strength mà người dùng đã đặt cho chat này.
func (r *ChatUserSettingsRepository) HideChat(ctx context.Context, chatID, userID string) error {
	now := time.Now().UTC()
	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO chat_user_settings (chat_id, user_id, hidden_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE hidden_at = ?, updated_at = ?`,
		chatID, userID, now, now, now, now, now).Error
	if err != nil {
		return fmt.Errorf("hide chat: %w", err)
	}
	return nil
}

// UnhideChat bỏ mốc ẩn khi người dùng chủ động mở lại hội thoại (chọn người đó
// trong "chat mới"), để hội thoại nằm lại trong danh sách ngay cả khi chưa có
// tin nhắn mới.
func (r *ChatUserSettingsRepository) UnhideChat(ctx context.Context, chatID, userID string) error {
	err := r.db.WithContext(ctx).Model(&models.ChatUserSettings{}).
		Where("chat_id = ? AND user_id = ? AND hidden_at IS NOT NULL", chatID, userID).
		Updates(map[string]any{"hidden_at": nil, "updated_at": time.Now().UTC()}).Error
	if err != nil {
		return fmt.Errorf("unhide chat: %w", err)
	}
	return nil
}
