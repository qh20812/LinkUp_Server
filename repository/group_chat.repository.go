package repository

import (
	"context"
	"fmt"
	"linkup/models"

	"gorm.io/gorm"
)

type GroupChatRepository struct {
	db *gorm.DB
}

func NewGroupChatRepository(db *gorm.DB) *GroupChatRepository {
	return &GroupChatRepository{db: db}
}

func (r *GroupChatRepository) CreateGroup(ctx context.Context, group *models.Chat, participants []models.ChatParticipant) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// GORM tự động map struct Chat vào bảng chats
		if err := tx.Create(group).Error; err != nil {
			return fmt.Errorf("lỗi khi lưu thông tin nhóm chat: %w", err)
		}

		if len(participants) > 0 {
			if err := tx.Create(&participants).Error; err != nil {
				return fmt.Errorf("lỗi khi lưu danh sách thành viên: %w", err)
			}
		}
		return nil
	})
}

func (r *GroupChatRepository) AddMember(ctx context.Context, participant *models.ChatParticipant) error {
	return r.db.WithContext(ctx).Create(participant).Error
}

func (r *GroupChatRepository) IsUserMember(ctx context.Context, chatID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ChatParticipant{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Count(&count).Error
	return count > 0, err
}
