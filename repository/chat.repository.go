package repository

import (
	"context"
	"errors"
	"fmt"
	"linkup/models"
	"time"

	"gorm.io/gorm"
)

var ErrChatNotFound = errors.New("chat not found")

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
		Where("p1.user_id = ? AND p2.user_id = ?", userA, userB).
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
		if err := tx.Create(chat).Error; err != nil {
			return err
		}
		if len(participants) > 0 {
			if err := tx.Create(&participants).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("create direct chat: %w", err)
	}
	return chat, nil
}

func (r *ChatRepository) FindMessageByID(ctx context.Context, messageID string) (*models.Message, error) {
	var message models.Message
	err := r.db.WithContext(ctx).Where("id = ?", messageID).First(&message).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("message not found")
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
