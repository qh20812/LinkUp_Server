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
			return fmt.Errorf("delete chat: %w", err)
		}

		return nil
	})

	return err
}

func (r *ChatRepository) UpdateChat(ctx context.Context, chat *models.Chat) error {
	return r.db.WithContext(ctx).Save(chat).Error
}

// FindGroupChatsByCreator lấy danh sách group chat do người dùng tạo.
func (r *ChatRepository) FindGroupChatsByCreator(ctx context.Context, creatorID string) ([]models.Chat, error) {
	var chats []models.Chat
	err := r.db.WithContext(ctx).
		Where("creator_id = ? AND type = ?", creatorID, models.ChatTypeGroup).
		Find(&chats).Error
	if err != nil {
		return nil, fmt.Errorf("tìm group chat theo người tạo thất bại: %w", err)
	}
	return chats, nil
}

// FindOldestParticipant tìm participant tham gia sớm nhất trong group chat (trừ excludeUserID).
func (r *ChatRepository) FindOldestParticipant(ctx context.Context, chatID, excludeUserID string) (*models.ChatParticipant, error) {
	var p models.ChatParticipant
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id <> ?", chatID, excludeUserID).
		Order("joined_at ASC").
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("tìm participant cũ nhất thất bại: %w", err)
	}
	return &p, nil
}

// ── Admin: Group management ─────────────────────────────────────────────────

func (r *ChatRepository) ListGroups(ctx context.Context, keyword, status string, pageSize, offset int) ([]dto.AdminGroupListItem, error) {
	query := r.db.WithContext(ctx).
		Table("chats").
		Select(`chats.id, chats.name, chats.creator_id, chats.status, chats.created_at,
			COALESCE((SELECT COUNT(*) FROM chat_participants WHERE chat_id = chats.id), 0) AS member_count,
			COALESCE((SELECT display_name FROM profiles WHERE user_id = chats.creator_id), '') AS creator_name`).
		Where("chats.type = ?", models.ChatTypeGroup)

	if keyword != "" {
		query = query.Where("chats.name LIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("chats.status = ?", status)
	}

	var results []struct {
		ID          string `gorm:"column:id"`
		Name        string `gorm:"column:name"`
		CreatorID   *string `gorm:"column:creator_id"`
		CreatorName string `gorm:"column:creator_name"`
		MemberCount int    `gorm:"column:member_count"`
		Status      string `gorm:"column:status"`
		CreatedAt   time.Time `gorm:"column:created_at"`
	}
	if err := query.Order("chats.created_at DESC").Offset(offset).Limit(pageSize).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}

	items := make([]dto.AdminGroupListItem, 0, len(results))
	for _, r := range results {
		items = append(items, dto.AdminGroupListItem{
			ID:          r.ID,
			Name:        r.Name,
			CreatorID:   r.CreatorID,
			CreatorName: r.CreatorName,
			MemberCount: r.MemberCount,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt,
		})
	}
	return items, nil
}

func (r *ChatRepository) CountGroups(ctx context.Context, keyword, status string) (int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Chat{}).Where("type = ?", models.ChatTypeGroup)
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count groups: %w", err)
	}
	return total, nil
}

func (r *ChatRepository) UpdateChatStatus(ctx context.Context, chatID string, status models.ChatStatus) error {
	return r.db.WithContext(ctx).Model(&models.Chat{}).Where("id = ?", chatID).Update("status", status).Error
}

func (r *ChatRepository) GetGroupMembers(ctx context.Context, chatID string) ([]dto.AdminGroupMember, error) {
	type memberRow struct {
		UserID      string `gorm:"column:user_id"`
		Role        string `gorm:"column:role"`
		DisplayName string `gorm:"column:display_name"`
		AvatarURI   string `gorm:"column:avatar_uri"`
	}
	var rows []memberRow
	err := r.db.WithContext(ctx).
		Table("chat_participants").
		Select(`chat_participants.user_id, chat_participants.role,
			COALESCE(profiles.display_name, '') AS display_name,
			COALESCE(profiles.avatar_uri, '') AS avatar_uri`).
		Joins("LEFT JOIN profiles ON profiles.user_id = chat_participants.user_id").
		Where("chat_participants.chat_id = ?", chatID).
		Order("chat_participants.joined_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("get group members: %w", err)
	}

	items := make([]dto.AdminGroupMember, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.AdminGroupMember{
			UserID:      row.UserID,
			DisplayName: row.DisplayName,
			AvatarURI:   row.AvatarURI,
			Role:        row.Role,
		})
	}
	return items, nil
}

// ListUserChats trả về danh sách chat trực tiếp của người dùng (cùng đối phương
// và tin nhắn cuối). Tin nhắn cuối trả về ở dạng mã hóa — service sẽ giải mã.
func (r *ChatRepository) ListUserChats(ctx context.Context, userID string) ([]dto.ChatConversationDTO, error) {
	rows := []struct {
		ChatID             string     `gorm:"column:chat_id"`
		PartnerUserID      string     `gorm:"column:partner_user_id"`
		PartnerDisplayName string     `gorm:"column:partner_display_name"`
		PartnerAvatarURI   string     `gorm:"column:partner_avatar_uri"`
		LastMessageID      *string    `gorm:"column:last_message_id"`
		LastContent        *string    `gorm:"column:last_content"`
		LastSenderID       *string    `gorm:"column:last_sender_id"`
		LastE2EVersion     *int       `gorm:"column:last_e2e_version"`
		LastCreatedAt      *time.Time `gorm:"column:last_created_at"`
		UpdatedAt          time.Time  `gorm:"column:updated_at"`
	}{}

	err := r.db.WithContext(ctx).
		Table("chats").
		Select(`chats.id AS chat_id,
			partner.user_id AS partner_user_id,
			COALESCE(profiles.display_name, '') AS partner_display_name,
			COALESCE(profiles.avatar_uri, '') AS partner_avatar_uri,
			lm.id AS last_message_id,
			lm.content AS last_content,
			lm.sender_id AS last_sender_id,
			lm.e2e_version AS last_e2e_version,
			lm.created_at AS last_created_at,
			COALESCE(lm.created_at, chats.created_at) AS updated_at`).
		Joins("JOIN chat_participants AS me ON me.chat_id = chats.id AND me.user_id = ?", userID).
		Joins("JOIN chat_participants AS partner ON partner.chat_id = chats.id AND partner.user_id <> ?", userID).
		Joins("LEFT JOIN profiles ON profiles.user_id = partner.user_id").
		Joins(`LEFT JOIN messages AS lm ON lm.id = (
			SELECT m2.id FROM messages m2
			WHERE m2.chat_id = chats.id
				AND ((m2.sender_id = ? AND m2.deleted_for_sender = false)
					OR (m2.sender_id <> ? AND m2.deleted_for_receiver = false))
			ORDER BY m2.created_at DESC
			LIMIT 1
		)`, userID, userID).
		Where("chats.type = ?", models.ChatTypeDirect).
		Order("COALESCE(lm.created_at, chats.created_at) DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list user chats: %w", err)
	}

	items := make([]dto.ChatConversationDTO, 0, len(rows))
	for _, row := range rows {
		conv := dto.ChatConversationDTO{
			ChatID: row.ChatID,
			Partner: dto.ChatPartnerDTO{
				UserID:      row.PartnerUserID,
				DisplayName: row.PartnerDisplayName,
				AvatarURI:   row.PartnerAvatarURI,
			},
			UpdatedAt: row.UpdatedAt,
		}

		if row.LastMessageID != nil && row.LastContent != nil && row.LastCreatedAt != nil {
			e2eVersion := 0
			if row.LastE2EVersion != nil {
				e2eVersion = *row.LastE2EVersion
			}
			conv.IsEncrypted = e2eVersion == 1
			conv.LastMessage = &dto.MessagePayload{
				ID:         *row.LastMessageID,
				ChatID:     row.ChatID,
				SenderID:   derefString(row.LastSenderID),
				Content:    *row.LastContent,
				E2EVersion: e2eVersion,
				CreatedAt:  *row.LastCreatedAt,
			}
		}

		items = append(items, conv)
	}
	return items, nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (r *ChatRepository) GetUserMute(ctx context.Context, chatID, userID string) (*models.GroupChatMute, error) {
    var mute models.GroupChatMute
    err := r.db.WithContext(ctx).Where("chat_id = ? AND user_id = ?", chatID, userID).First(&mute).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("get user mute: %w", err)
    }

    now := time.Now().UTC()
    if mute.ExpiresAt != nil && mute.ExpiresAt.Before(now) {
        if err := r.db.WithContext(ctx).Where("chat_id = ? AND user_id = ?", chatID, userID).Delete(&models.GroupChatMute{}).Error; err != nil {
            return nil, fmt.Errorf("cleanup expired mute: %w", err)
        }
        return nil, nil
    }
    return &mute, nil
}

func (r *ChatRepository) GetDisplayName(ctx context.Context, userID string) string {
	var result struct {
		DisplayName string `gorm:"column:display_name"`
	}
	err := r.db.WithContext(ctx).
		Table("profiles").
		Select("COALESCE(display_name, '') AS display_name").
		Where("user_id = ?", userID).
		First(&result).Error
	if err != nil || result.DisplayName == "" {
		return userID
	}
	return result.DisplayName
}
