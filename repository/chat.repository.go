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

func (r *ChatRepository) DB() *gorm.DB {
	return r.db
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

func (r *ChatRepository) GetMessagesPaged(ctx context.Context, chatID string, beforeCreatedAt *time.Time, beforeID string, limit int) ([]models.Message, error) {
	if limit <= 0 {
		limit = 30
	}
	q := r.db.WithContext(ctx).Where("chat_id = ?", chatID)
	if beforeCreatedAt != nil {
		q = q.Where("(created_at < ? OR (created_at = ? AND id < ?))", *beforeCreatedAt, *beforeCreatedAt, beforeID)
	}
	var messages []models.Message
	err := q.Order("created_at DESC, id DESC").Limit(limit + 1).Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	return messages, nil
}

func (r *ChatRepository) GetReplyPreviews(ctx context.Context, messageIDs []string) map[string]*dto.ReplyPreview {
	if len(messageIDs) == 0 {
		return nil
	}
	type row struct {
		ID          string `gorm:"column:id"`
		Content     string `gorm:"column:content"`
		SenderID    string `gorm:"column:sender_id"`
		DisplayName string `gorm:"column:display_name"`
		AvatarURI   string `gorm:"column:avatar_uri"`
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("messages AS m").
		Select(`m.id,
			m.content,
			m.sender_id,
			COALESCE(p.display_name, '') AS display_name,
			COALESCE(p.avatar_uri, '') AS avatar_uri`).
		Joins(`LEFT JOIN profiles AS p ON p.user_id = m.sender_id`).
		Where("m.id IN ?", messageIDs).
		Scan(&rows).Error
	if err != nil {
		return nil
	}
	result := make(map[string]*dto.ReplyPreview, len(rows))
	for _, r := range rows {
		result[r.ID] = &dto.ReplyPreview{
			ID:           r.ID,
			Content:      r.Content,
			SenderID:     r.SenderID,
			SenderName:   r.DisplayName,
			SenderAvatar: r.AvatarURI,
		}
	}
	return result
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
LastMessageID        *string     `gorm:"column:last_message_id"`
		LastContent          *string     `gorm:"column:last_content"`
		LastSenderID         *string     `gorm:"column:last_sender_id"`
		LastE2EVersion       *int        `gorm:"column:last_e2e_version"`
		LastMediaType        string      `gorm:"column:last_media_type"`
		LastMediaDuration    int         `gorm:"column:last_media_duration"`
		LastForwardedFrom    *string     `gorm:"column:last_forwarded_from"`
		LastCreatedAt        *time.Time  `gorm:"column:last_created_at"`
		UpdatedAt            time.Time   `gorm:"column:updated_at"`
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
			COALESCE(lmm.file_type, '') AS last_media_type,
			COALESCE(lmm.duration_seconds, 0) AS last_media_duration,
			lm.forwarded_from AS last_forwarded_from,
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
		Joins("LEFT JOIN media AS lmm ON lmm.id = lm.media_id").
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
				ID:            *row.LastMessageID,
				ChatID:        row.ChatID,
				SenderID:      derefString(row.LastSenderID),
				Content:       *row.LastContent,
				MediaType:     row.LastMediaType,
				DurationSeconds: row.LastMediaDuration,
				ForwardedFrom: row.LastForwardedFrom,
				E2EVersion:    e2eVersion,
				CreatedAt:     *row.LastCreatedAt,
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

// ── Pinned Messages ────────────────────────────────────────────────────────

const maxPinnedPerChat = 2

// PinMessage ghim tin nhắn vào chat. Trả về lỗi nếu chat đã đủ 2 tin ghim.
// Khi đầy, phải gọi AutoUnpinOldest trước.
func (r *ChatRepository) PinMessage(ctx context.Context, chatID, messageID, pinnedBy string) (*models.PinnedMessage, error) {
	// Đếm số tin đã ghim trong chat
	var count int64
	if err := r.db.WithContext(ctx).
		Table("pinned_messages").
		Where("chat_id = ?", chatID).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("count pinned messages: %w", err)
	}
	if count >= maxPinnedPerChat {
		return nil, fmt.Errorf("chat đã đạt tối đa %d tin nhắn ghim", maxPinnedPerChat)
	}

	pm := &models.PinnedMessage{
		ID:        utils.GenerateUUID(),
		ChatID:    chatID,
		MessageID: messageID,
		PinnedBy:  pinnedBy,
		PinnedAt:  time.Now().UTC(),
	}
	if err := r.db.WithContext(ctx).Create(pm).Error; err != nil {
		return nil, fmt.Errorf("pin message: %w", err)
	}
	return pm, nil
}

// AutoUnpinOldest bỏ ghim tin nhắn cũ nhất trong chat (dùng khi muốn ghim thêm
// mà đã đạt max).
func (r *ChatRepository) AutoUnpinOldest(ctx context.Context, chatID string) error {
	var oldest models.PinnedMessage
	err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Order("pinned_at ASC").
		First(&oldest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("find oldest pin: %w", err)
	}
	return r.db.WithContext(ctx).Where("id = ?", oldest.ID).Delete(&models.PinnedMessage{}).Error
}

// UnpinMessage bỏ ghim một tin nhắn trong chat.
func (r *ChatRepository) UnpinMessage(ctx context.Context, chatID, messageID string) error {
	tx := r.db.WithContext(ctx).
		Where("chat_id = ? AND message_id = ?", chatID, messageID).
		Delete(&models.PinnedMessage{})
	if tx.Error != nil {
		return fmt.Errorf("unpin message: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("tin nhắn chưa được ghim")
	}
	return nil
}

// GetPinnedMessages lấy danh sách tin nhắn đã ghim trong chat, kèm nội dung đã
// giải mã và tên người gửi.
func (r *ChatRepository) GetPinnedMessages(ctx context.Context, chatID string) ([]dto.PinnedMessageDTO, error) {
	type pinRow struct {
		ID           string `gorm:"column:id"`
		MessageID    string `gorm:"column:message_id"`
		PinnedBy     string `gorm:"column:pinned_by"`
		PinnedAt     time.Time `gorm:"column:pinned_at"`
		Content      string `gorm:"column:content"`
		SenderID     string `gorm:"column:sender_id"`
		DisplayName  string `gorm:"column:display_name"`
	}
	var rows []pinRow
	err := r.db.WithContext(ctx).
		Table("pinned_messages AS pm").
		Select(`pm.id, pm.message_id, pm.pinned_by, pm.pinned_at,
			m.content, m.sender_id,
			COALESCE(p.display_name, '') AS display_name`).
		Joins("JOIN messages AS m ON m.id = pm.message_id").
		Joins("LEFT JOIN profiles AS p ON p.user_id = m.sender_id").
		Where("pm.chat_id = ?", chatID).
		Order("pm.pinned_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("get pinned messages: %w", err)
	}

	result := make([]dto.PinnedMessageDTO, 0, len(rows))
	for _, row := range rows {
		result = append(result, dto.PinnedMessageDTO{
			ID:         row.ID,
			MessageID:  row.MessageID,
			PinnedBy:   row.PinnedBy,
			PinnedAt:   row.PinnedAt,
			Content:    row.Content,
			SenderID:   row.SenderID,
			SenderName: row.DisplayName,
		})
	}
	return result, nil
}

// CountPinnedMessages đếm số tin đã ghim trong chat.
func (r *ChatRepository) CountPinnedMessages(ctx context.Context, chatID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("pinned_messages").
		Where("chat_id = ?", chatID).
		Count(&count).Error
	return count, err
}

// IsMessagePinned kiểm tra tin nhắn đã được ghim chưa.
func (r *ChatRepository) IsMessagePinned(ctx context.Context, chatID, messageID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("pinned_messages").
		Where("chat_id = ? AND message_id = ?", chatID, messageID).
		Count(&count).Error
	return count > 0, err
}

// UpsertChatRead nâng watermark đọc của user trong chat. Chỉ cập nhật khi
// tin mới (theo created_at) trễ hơn watermark hiện tại — bảo toàn tính
// đơn điệu, tránh "đọc lùi" khi client gửi tin cũ hơn.
func (r *ChatRepository) UpsertChatRead(ctx context.Context, chatID, userID, messageID string, messageCreatedAt time.Time) (bool, error) {
	var existing models.ChatRead
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		First(&existing).Error
	if err == nil {
		if !messageCreatedAt.After(existing.LastReadAt) {
			// Watermark đã cao hơn — không nâng lùi.
			return false, nil
		}
		err = r.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
			"last_read_at":    messageCreatedAt,
			"last_message_id": messageID,
			"updated_at":      time.Now().UTC(),
		}).Error
		if err != nil {
			return false, fmt.Errorf("update chat read: %w", err)
		}
		return true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, fmt.Errorf("get chat read: %w", err)
	}

	now := time.Now().UTC()
	cr := models.ChatRead{
		ChatID:        chatID,
		UserID:        userID,
		LastReadAt:    messageCreatedAt,
		LastMessageID: messageID,
		UpdatedAt:     now,
	}
	err = r.db.WithContext(ctx).Create(&cr).Error
	if err != nil {
		return false, fmt.Errorf("create chat read: %w", err)
	}
	return true, nil
}

// GetChatReadWatermarks trả về watermark đọc (theo user) của toàn chat.
// Map user_id → last_read_at. Dùng để tính seen_by cho từng tin nhắn.
func (r *ChatRepository) GetChatReadWatermarks(ctx context.Context, chatID string) (map[string]time.Time, error) {
	var rows []models.ChatRead
	err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("get chat read watermarks: %w", err)
	}
	result := make(map[string]time.Time, len(rows))
	for _, row := range rows {
		result[row.UserID] = row.LastReadAt
	}
	return result, nil
}

// ── Message reactions (Phase 4) ─────────────────────────────────────────────

// ToggleMessageReaction bật/tắt reaction của user trên tin nhắn.
// - Chưa có → tạo mới, trả "added".
// - Có và cùng emoji → xóa, trả "removed".
// - Có nhưng khác emoji → đổi emoji, trả "updated".
func (r *ChatRepository) ToggleMessageReaction(ctx context.Context, messageID, userID, emojiID string) (string, error) {
	var existing models.MessageReaction
	err := r.db.WithContext(ctx).
		Where("message_id = ? AND user_id = ?", messageID, userID).
		First(&existing).Error
	if err == nil {
		if existing.EmojiID == emojiID {
			err = r.db.WithContext(ctx).
				Where("message_id = ? AND user_id = ?", messageID, userID).
				Delete(&models.MessageReaction{}).Error
			if err != nil {
				return "", fmt.Errorf("delete message reaction: %w", err)
			}
			return "removed", nil
		}
		err = r.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
			"emoji_id":   emojiID,
			"updated_at": time.Now().UTC(),
		}).Error
		if err != nil {
			return "", fmt.Errorf("update message reaction: %w", err)
		}
		return "updated", nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("get message reaction: %w", err)
	}

	reaction := models.MessageReaction{
		MessageID: messageID,
		UserID:    userID,
		EmojiID:   emojiID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err = r.db.WithContext(ctx).Create(&reaction).Error
	if err != nil {
		return "", fmt.Errorf("create message reaction: %w", err)
	}
	return "added", nil
}

// GetMessageReactions trả về reaction (theo message_id) cho danh sách tin nhắn.
func (r *ChatRepository) GetMessageReactions(ctx context.Context, messageIDs []string) map[string][]models.MessageReaction {
	if len(messageIDs) == 0 {
		return nil
	}
	var rows []models.MessageReaction
	err := r.db.WithContext(ctx).
		Where("message_id IN ?", messageIDs).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil
	}
	result := make(map[string][]models.MessageReaction, len(messageIDs))
	for _, row := range rows {
		result[row.MessageID] = append(result[row.MessageID], row)
	}
	return result
}

// ── Message forwarding (Phase 4) ────────────────────────────────────────────

// IncrementForwardsCount tăng số lần chuyển tiếp của tin nhắn gốc lên 1.
func (r *ChatRepository) IncrementForwardsCount(ctx context.Context, messageID string) error {
	err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id = ?", messageID).
		UpdateColumn("forwards_count", gorm.Expr("forwards_count + 1")).Error
	if err != nil {
		return fmt.Errorf("increment forwards count: %w", err)
	}
	return nil
}

// CanReadMessage kiểm tra xem user có quyền đọc tin nhắn (thuộc hội thoại gốc)
// hay không — dùng khi chuyển tiếp để tránh lộ nội dung hội thoại khác.
func (r *ChatRepository) CanReadMessage(ctx context.Context, messageID, userID string) (bool, error) {
	msg, err := r.FindMessageByID(ctx, messageID)
	if err != nil {
		return false, err
	}
	return r.IsUserParticipant(ctx, msg.ChatID, userID)
}
