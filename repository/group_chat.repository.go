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

// Rời khỏi nhóm chat (Xóa bản ghi tham gia)
func (r *GroupChatRepository) LeaveGroup(ctx context.Context, chatID, userID string) error {
	return r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Delete(&models.ChatParticipant{}).Error
}

// Chặn thành viên (Sử dụng Transaction: Vừa xóa khỏi nhóm vừa đưa vào danh sách Ban)
func (r *GroupChatRepository) BanUser(ctx context.Context, ban *models.GroupChatBan) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Xóa khỏi bảng thành viên nếu user này đang ở trong nhóm
		if err := tx.Where("chat_id = ? AND user_id = ?", ban.ChatID, ban.UserID).
			Delete(&models.ChatParticipant{}).Error; err != nil {
			return err
		}
		// 2. Thêm vào bảng danh sách chặn (Ban)
		if err := tx.Create(ban).Error; err != nil {
			return err
		}
		return nil
	})
}

// Kiểm tra xem user có phải Admin của nhóm hay không
// Trong repository/group_chat_repository.go
func (r *GroupChatRepository) IsUserAdmin(ctx context.Context, chatID, userID string) (bool, error) {
	var count int64
	// 🌟 THÊM .Debug() VÀO ĐÂY:
	err := r.db.Debug().WithContext(ctx).
		Model(&models.ChatParticipant{}).
		Where("chat_id = ? AND user_id = ? AND role = ?", chatID, userID, "CHAT_ADMIN").
		Count(&count).Error
	return count > 0, err
}

// Kiểm tra xem User này có nằm trong danh sách đen (bị Ban) của nhóm không
func (r *GroupChatRepository) IsUserBanned(ctx context.Context, chatID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("group_chat_bans").
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Count(&count).Error
	return count > 0, err
}
