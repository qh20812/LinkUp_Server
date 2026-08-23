package repository

import (
	"context"
	"errors"
	"fmt"
	"linkup/dto"
	"linkup/models"
	"linkup/utils"
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

func (r *GroupChatRepository) TransferOwnership(ctx context.Context, chatID, oldCreatorID, newCreatorID string, keepAdmin bool, transferredAt time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Chat{}).Where("id = ?", chatID).Update("creator_id", newCreatorID).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.ChatParticipant{}).
			Where("chat_id = ? AND user_id = ?", chatID, newCreatorID).
			Update("role", models.ChatRoleAdmin).Error; err != nil {
			return err
		}

		if !keepAdmin {
			if err := tx.Model(&models.ChatParticipant{}).
				Where("chat_id = ? AND user_id = ? AND role = ?", chatID, oldCreatorID, models.ChatRoleAdmin).
				Update("role", models.ChatRoleMember).Error; err != nil {
				return err
			}
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

// FindAllAdmins lấy danh sách user IDs có role CHAT_ADMIN trong group chat (trừ excludeUserID).
func (r *GroupChatRepository) FindAllAdmins(ctx context.Context, chatID, excludeUserID string) ([]string, error) {
	var userIDs []string
	err := r.db.WithContext(ctx).
		Table("chat_participants").
		Where("chat_id = ? AND role = ? AND user_id <> ?", chatID, models.ChatRoleAdmin, excludeUserID).
		Pluck("user_id", &userIDs).Error
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách admin chat thất bại: %w", err)
	}
	return userIDs, nil
}

// FindOldestMember tìm participant tham gia sớm nhất trong group chat (trừ excludeUserID).
func (r *GroupChatRepository) FindOldestMember(ctx context.Context, chatID, excludeUserID string) (*models.ChatParticipant, error) {
	var p models.ChatParticipant
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id <> ?", chatID, excludeUserID).
		Order("joined_at ASC").
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("tìm thành viên cũ nhất thất bại: %w", err)
	}
	return &p, nil
}
func (r *GroupChatRepository) GetAdminIDs(ctx context.Context, chatID string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).
		Table("chat_participants").
		Where("chat_id = ? AND role = ?", chatID, models.ChatRoleAdmin).
		Pluck("user_id", &ids).Error
	return ids, err
}

func (r *GroupChatRepository) AnonymizeMessagesBySender(ctx context.Context, chatID, senderID string) error {
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("chat_id = ? AND sender_id = ?", chatID, senderID).
		Updates(map[string]any{
			"is_anonymized":  true,
			"anonymous_name": "Thành viên ẩn danh",
		}).Error
}

func (r *GroupChatRepository) CreateMemberRequest(ctx context.Context, req *models.GroupChatMemberRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *GroupChatRepository) FindPendingMemberRequest(ctx context.Context, chatID, targetUserID string) (*models.GroupChatMemberRequest, error) {
	var req models.GroupChatMemberRequest
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND target_user_id = ? AND status = ?", chatID, targetUserID, models.GroupChatMemberRequestPending).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *GroupChatRepository) FindMemberRequestByID(ctx context.Context, requestID string) (*models.GroupChatMemberRequest, error) {
	var req models.GroupChatMemberRequest
	err := r.db.WithContext(ctx).Where("id = ?", requestID).First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *GroupChatRepository) ApproveMemberRequest(ctx context.Context, requestID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var req models.GroupChatMemberRequest
		if err := tx.Where("id = ? AND status = ?", requestID, models.GroupChatMemberRequestPending).
			First(&req).Error; err != nil {
			return err
		}

		now := time.Now().UTC()

		req.Status = models.GroupChatMemberRequestApproved
		req.RespondedAt = &now
		if err := tx.Save(&req).Error; err != nil {
			return err
		}

		participant := models.NewChatParticipant(req.ChatID, req.TargetUserID, models.ChatRoleMember)
		participant.ID = utils.GenerateUUID()
		participant.JoinedAt = now

		return tx.Create(&participant).Error
	})
}

func (r *GroupChatRepository) RejectMemberRequest(ctx context.Context, requestID string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.GroupChatMemberRequest{}).
		Where("id = ? AND status = ?", requestID, models.GroupChatMemberRequestPending).
		Updates(map[string]any{
			"status":       models.GroupChatMemberRequestRejected,
			"responded_at": now,
		}).Error
}

func (r *GroupChatRepository) GetGroupMembersWithProfiles(ctx context.Context, chatID string) ([]dto.GroupChatMemberDTO, error) {
	type row struct {
		UserID      string `gorm:"column:user_id"`
		DisplayName string `gorm:"column:display_name"`
		AvatarURI   string `gorm:"column:avatar_uri"`
		Role        string `gorm:"column:role"`
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Table("chat_participants AS cp").
		Select(`cp.user_id,
			COALESCE(p.display_name, '') AS display_name,
			COALESCE(p.avatar_uri, '') AS avatar_uri,
			cp.role`).
		Joins(`LEFT JOIN profiles AS p ON p.user_id = cp.user_id`).
		Where("cp.chat_id = ?", chatID).
		Order("cp.joined_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách thành viên nhóm thất bại: %w", err)
	}

	members := make([]dto.GroupChatMemberDTO, 0, len(rows))
	for _, r := range rows {
		members = append(members, dto.GroupChatMemberDTO{
			UserID:      r.UserID,
			DisplayName: r.DisplayName,
			AvatarURI:   r.AvatarURI,
			Role:        r.Role,
		})
	}
	return members, nil
}

func (r *GroupChatRepository) GetMemberProfiles(ctx context.Context, chatID string, userIDs []string, out interface{}) error {
	return r.db.WithContext(ctx).
		Table("chat_participants AS cp").
		Select(`cp.user_id,
			COALESCE(p.display_name, '') AS display_name,
			COALESCE(p.avatar_uri, '') AS avatar_uri`).
		Joins(`LEFT JOIN profiles AS p ON p.user_id = cp.user_id`).
		Where("cp.chat_id = ? AND cp.user_id IN ?", chatID, userIDs).
		Scan(out).Error
}

func (r *GroupChatRepository) ListUserGroupChats(ctx context.Context, userID string) ([]dto.GroupChatConversationDTO, error) {
	type row struct {
		ChatID      string     `gorm:"column:chat_id"`
		Name        string     `gorm:"column:name"`
		AvatarURI   string     `gorm:"column:avatar_uri"`
		MemberCount int        `gorm:"column:member_count"`
		LastMsgID   *string    `gorm:"column:last_message_id"`
		LastContent *string    `gorm:"column:last_content"`
		LastSender  *string    `gorm:"column:last_sender_id"`
		LastCreated *time.Time `gorm:"column:last_created_at"`
		UpdatedAt   time.Time  `gorm:"column:updated_at"`
	}

	var rows []row

	err := r.db.WithContext(ctx).
		Table("chats").
		Select(`chats.id AS chat_id,
			chats.name AS name,
			COALESCE(chats.avatar_uri, '') AS avatar_uri,
			COUNT(DISTINCT cp2.user_id) AS member_count,
			lm.id AS last_message_id,
			lm.content AS last_content,
			lm.sender_id AS last_sender_id,
			lm.created_at AS last_created_at,
			COALESCE(lm.created_at, chats.created_at) AS updated_at`).
		Joins("JOIN chat_participants AS me ON me.chat_id = chats.id AND me.user_id = ?", userID).
		Joins("JOIN chat_participants AS cp2 ON cp2.chat_id = chats.id").
		Joins(`LEFT JOIN messages AS lm ON lm.id = (
			SELECT m2.id FROM messages m2
			WHERE m2.chat_id = chats.id
				AND ((m2.sender_id = ? AND m2.deleted_for_sender = false)
					OR (m2.sender_id <> ? AND m2.deleted_for_receiver = false))
			ORDER BY m2.created_at DESC
			LIMIT 1
		)`, userID, userID).
		Where("chats.type = ?", models.ChatTypeGroup).
		Group("chats.id, chats.name, chats.avatar_uri, lm.id, lm.content, lm.sender_id, lm.created_at, chats.created_at").
		Order("COALESCE(lm.created_at, chats.created_at) DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list user group chats: %w", err)
	}

	items := make([]dto.GroupChatConversationDTO, 0, len(rows))
	for _, r := range rows {
		item := dto.GroupChatConversationDTO{
			ChatID:      r.ChatID,
			Name:        r.Name,
			AvatarURI:   r.AvatarURI,
			MemberCount: r.MemberCount,
			UpdatedAt:   r.UpdatedAt,
		}
		if r.LastMsgID != nil && r.LastContent != nil && r.LastCreated != nil {
			item.LastMessage = &dto.MessagePayload{
				ID:        *r.LastMsgID,
				ChatID:    r.ChatID,
				SenderID:  derefString(r.LastSender),
				Content:   *r.LastContent,
				CreatedAt: *r.LastCreated,
			}
		}
		items = append(items, item)
	}
	return items, nil
}
