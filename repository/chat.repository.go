package repository

import (
	"context"
	"errors"
	"fmt"
	"linkup/models"
	"linkup/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrChatNotFound = errors.New("không tìm thấy chat")

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) FindChatByID(ctx context.Context, chatID string) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.WithContext(ctx).Where("id = ?", chatID).First(&chat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrChatNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find chat by id: %w", err)
	}
	return &chat, nil
}

func (r *ChatRepository) IsUserParticipant(ctx context.Context, chatID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("chat_participants").
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check chat participant: %w", err)
	}
	return count > 0, nil
}

func (r *ChatRepository) FindDirectChatPartner(ctx context.Context, chatID, userID string) (string, error) {
	var participant models.ChatParticipant
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id <> ?", chatID, userID).
		First(&participant).Error
	if err != nil {
		return "", fmt.Errorf("find direct chat partner: %w", err)
	}
	return participant.UserID, nil
}

func (r *ChatRepository) CreateMessage(ctx context.Context, message *models.Message) (*models.Message, error) {
	tx := r.db.WithContext(ctx).Create(message)
	if tx.Error != nil {
		return nil, fmt.Errorf("create message: %w", tx.Error)
	}
	return message, nil
}

func (r *ChatRepository) FindDirectChat(ctx context.Context, userA, userB string) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.WithContext(ctx).
		Table("chats").
		Joins("JOIN chat_participants p1 ON p1.chat_id = chats.id").
		Joins("JOIN chat_participants p2 ON p2.chat_id = chats.id").
		Where("chats.type = ?", models.ChatTypeDirect).
		Where("(p1.user_id = ? AND p2.user_id = ?) OR (p1.user_id = ? AND p2.user_id = ?)", userA, userB, userB, userA).
		First(&chat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrChatNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find direct chat: %w", err)
	}
	return &chat, nil
}

func (r *ChatRepository) CreateDirectChat(ctx context.Context, chat *models.Chat, participants []models.ChatParticipant) (*models.Chat, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		encKey, err := utils.GenerateEncryptionKey()
		if err != nil {
			return fmt.Errorf("generate encryption key: %w", err)
		}
		chat.EncryptionKey = encKey

		if err := tx.Create(chat).Error; err != nil {
			return err
		}

		return tx.CreateInBatches(participants, 100).Error
	})
	return chat, err
}

func (r *ChatRepository) GetEncryptionKey(ctx context.Context, chatID string) (string, error) {
	var chat models.Chat
	err := r.db.WithContext(ctx).Select("encryption_key").Where("id = ?", chatID).First(&chat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("chat not found")
	}
	if err != nil {
		return "", fmt.Errorf("get encryption key: %w", err)
	}
	return chat.EncryptionKey, nil
}

func (r *ChatRepository) FindMessageByID(ctx context.Context, messageID string) (*models.Message, error) {
	var message models.Message
	err := r.db.WithContext(ctx).Where("id = ?", messageID).First(&message).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("không tìm thấy tin nhắn")
	}
	if err != nil {
		return nil, fmt.Errorf("find message by id: %w", err)
	}
	return &message, nil
}

func (r *ChatRepository) UpdateMessageDeleteStatus(ctx context.Context, messageID string, deletedForSender, deletedForReceiver bool, deletedAt *time.Time) (*models.Message, error) {
	updates := map[string]any{
		"deleted_for_sender":   deletedForSender,
		"deleted_for_receiver": deletedForReceiver,
	}
	if deletedAt != nil {
		updates["deleted_at"] = deletedAt
	}

	tx := r.db.WithContext(ctx).Model(&models.Message{}).Where("id = ?", messageID).Updates(updates)
	if tx.Error != nil {
		return nil, fmt.Errorf("update message delete status: %w", tx.Error)
	}

	return r.FindMessageByID(ctx, messageID)
}

func (r *ChatRepository) GetMessages(ctx context.Context, chatID, userID string) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Where("(sender_id = ? AND deleted_for_sender = false) OR (sender_id <> ? AND deleted_for_receiver = false)", userID, userID).
		Order("created_at DESC").
		Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	return messages, nil
}

func (r *ChatRepository) GetParticipantIDs(ctx context.Context, chatID string) ([]string, error) {
	var userIDs []string
	err := r.db.WithContext(ctx).
		Table("chat_participants").
		Where("chat_id = ?", chatID).
		Pluck("user_id", &userIDs).Error
	if err != nil {
		return nil, fmt.Errorf("get participants: %w", err)
	}
	return userIDs, nil
}

func (r *ChatRepository) SearchMessages(ctx context.Context, chatID, userID, keyword string) ([]models.Message, error) {
	var messages []models.Message
	pattern := "%" + strings.ToLower(keyword) + "%"
	err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Where("(sender_id = ? AND deleted_for_sender = false) OR (sender_id <> ? AND deleted_for_receiver = false)", userID, userID).
		Where("LOWER(content) LIKE ?", pattern).
		Order("created_at DESC").
		Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	return messages, nil
}

func (r *ChatRepository) IsEmojiExists(ctx context.Context, emojiID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("emojis").
		Where("id = ?", emojiID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check emoji exists: %w", err)
	}
	return count > 0, nil
}

func (r *ChatRepository) DeleteChat(ctx context.Context, chatID string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var chat models.Chat
		if err := tx.Where("id = ?", chatID).First(&chat).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrChatNotFound
			}
			return fmt.Errorf("find chat for deletion: %w", err)
		}

		if err := tx.Where("chat_id = ?", chatID).Delete(&models.Message{}).Error; err != nil {
			return fmt.Errorf("delete chat messages: %w", err)
		}
		if err := tx.Where("chat_id = ?", chatID).Delete(&models.ChatParticipant{}).Error; err != nil {
			return fmt.Errorf("delete chat participants: %w", err)
		}
		if err := tx.Where("id = ?", chatID).Delete(&models.Chat{}).Error; err != nil {
			return fmt.Errorf("delete chat: %w", &err)
		}

		return nil
	})

	return err
}

func (r *ChatRepository) UpdateChat(ctx context.Context, chat *models.Chat) error {
    return r.db.WithContext(ctx).Save(chat).Error
}