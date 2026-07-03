package repository

import (
	"context"
	"errors"
	"fmt"
	"linkup/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *GroupChatRepository) GetSettings(ctx context.Context, chatID string) (*models.GroupChatSettings, error) {
	var s models.GroupChatSettings
	err := r.db.WithContext(ctx).First(&s, "chat_id = ?", chatID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &models.GroupChatSettings{
				ChatID:         chatID,
				AllowMemberAdd: true,
			}, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *GroupChatRepository) UpsertSettings(ctx context.Context, settings *models.GroupChatSettings) error {
	now := time.Now().UTC()

	var existing models.GroupChatSettings
	err := r.db.WithContext(ctx).First(&existing, "chat_id = ?", settings.ChatID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			settings.CreatedAt = now
			settings.UpdatedAt = now
			return r.db.WithContext(ctx).Create(settings).Error
		}
		return err
	}

	updates := map[string]any{
		"allow_member_add":       settings.AllowMemberAdd,
		"last_admin_transfer_at": settings.LastAdminTransferAt,
		"updated_at":             now,
	}

	return r.db.WithContext(ctx).
		Model(&models.GroupChatSettings{}).
		Where("chat_id = ?", settings.ChatID).
		Updates(updates).Error
}

func (r *GroupChatRepository) SetLastAdminTransfer(ctx context.Context, chatID string, t *time.Time) error {
	if t == nil {

		return r.db.WithContext(ctx).Model(&models.GroupChatSettings{}).Where("chat_id = ?", chatID).
			Update("last_admin_transfer_at", nil).Error
	}
	return r.db.WithContext(ctx).Model(&models.GroupChatSettings{}).Where("chat_id = ?", chatID).
		Update("last_admin_transfer_at", *t).Error
}

func (r *GroupChatRepository) TransferAdmin(ctx context.Context, chatID, requesterID, targetUserID string, transferredAt time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ChatParticipant{}).
			Where("chat_id = ? AND user_id = ? AND role = ?", chatID, requesterID, models.ChatRoleAdmin).
			Update("role", models.ChatRoleMember).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.ChatParticipant{}).
			Where("chat_id = ? AND user_id = ?", chatID, targetUserID).
			Update("role", models.ChatRoleAdmin).Error; err != nil {
			return err
		}

		settings := &models.GroupChatSettings{
			ChatID:              chatID,
			LastAdminTransferAt: &transferredAt,
			UpdatedAt:           time.Now().UTC(),
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "chat_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"last_admin_transfer_at", "updated_at"}),
		}).Create(settings).Error
	})
}

func (r *GroupChatRepository) UpsertMemberSettings(ctx context.Context, settings *models.GroupChatMemberSettings) error {
	now := time.Now().UTC()
	settings.UpdatedAt = now

	existing := &models.GroupChatMemberSettings{}
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", settings.ChatID, settings.UserID).
		First(existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(settings).Error
	}
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Model(existing).Updates(map[string]any{
		"notifications_enabled": settings.NotificationsEnabled,
		"updated_at":            now,
	}).Error
}

func (r *GroupChatRepository) GetMemberSettings(ctx context.Context, chatID, userID string) (*models.GroupChatMemberSettings, error) {
	var s models.GroupChatMemberSettings
	err := r.db.WithContext(ctx).
		First(&s, "chat_id = ? AND user_id = ?", chatID, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.GroupChatMemberSettings{
				ChatID:               chatID,
				UserID:               userID,
				NotificationsEnabled: true,
			}, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *GroupChatRepository) MuteUser(ctx context.Context, mute *models.GroupChatMute) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chat_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"muted_by", "reason", "expires_at", "created_at"}),
	}).Create(mute).Error
}

func (r *GroupChatRepository) UnmuteUser(ctx context.Context, chatID, userID string) error {
	return r.db.WithContext(ctx).Where("chat_id = ? AND user_id = ?", chatID, userID).Delete(&models.GroupChatMute{}).Error
}

func (r *GroupChatRepository) GetMutesForChat(ctx context.Context, chatID string) ([]models.GroupChatMute, error) {
	var mutes []models.GroupChatMute
	err := r.db.WithContext(ctx).Where("chat_id = ?", chatID).Find(&mutes).Error
	return mutes, err
}
