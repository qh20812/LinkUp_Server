package services

import (
	"context"
	"fmt"
	"linkup/dto"
	"linkup/models"
	errorsapp "linkup/errors"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"strings"
	"time"
)

type GroupMessageService struct {
	chatRepo       *repository.ChatRepository
	groupRepo      *repository.GroupChatRepository
	mediaRepo      *repository.MediaRepository
	postRepo       *repository.PostRepository
	notifService   *NotificationService
	validation     *validations.ChatValidation
	groupCallRepo  *repository.GroupCallRepository
}

func NewGroupMessageService(
	chatRepo *repository.ChatRepository,
	groupRepo *repository.GroupChatRepository,
	mediaRepo *repository.MediaRepository,
	postRepo *repository.PostRepository,
	notifService *NotificationService,
	validation *validations.ChatValidation,
) *GroupMessageService {
	return &GroupMessageService{
		chatRepo:     chatRepo,
		groupRepo:    groupRepo,
		mediaRepo:    mediaRepo,
		postRepo:     postRepo,
		notifService: notifService,
		validation:   validation,
	}
}

func (s *GroupMessageService) ensureGroupMember(ctx context.Context, userID, chatID string) (*models.Chat, error) {
	chat, err := s.chatRepo.FindChatByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if chat.Type != models.ChatTypeGroup {
		return nil, errorsapp.New(errorsapp.ErrCodeGCNotGroupChat)
	}

	banned, err := s.groupRepo.IsUserBanned(ctx, chatID, userID)
	if err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodeGCBanned, err)
	}
	if banned {
		return nil, errorsapp.New(errorsapp.ErrCodeGCBanned)
	}

	isMember, err := s.groupRepo.IsUserMember(ctx, chatID, userID)
	if err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodeGCNotMember, err)
	}
	if !isMember {
		return nil, errorsapp.New(errorsapp.ErrCodeGCNotMember)
	}

	return chat, nil
}

// JoinRoom chỉ kiểm tra quyền tham gia, không tự động thêm member
func (s *GroupMessageService) JoinRoom(ctx context.Context, userID, chatID string) error {
	_, err := s.ensureGroupMember(ctx, userID, chatID)
	return err
}

func (s *GroupMessageService) SendMessage(
	ctx context.Context,
	userID, chatID, content string,
	emojiID, mediaID, gifURL, replyToMessageID, sharedPostID *string,
) (*models.Message, error) {
	chat, err := s.ensureGroupMember(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}

	mute, err := s.chatRepo.GetUserMute(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if mute != nil {
		var expiresStr string
		if mute.ExpiresAt != nil {
			expiresStr = mute.ExpiresAt.UTC().Format(time.RFC3339)
		} else {
			expiresStr = "vĩnh viễn"
		}
		return nil, fmt.Errorf("bạn đã bị tắt tiếng trong nhóm này (lý do: %s). Hết hạn: %s", mute.Reason, expiresStr)
	}

	if gifURL != nil && *gifURL != "" {
		gifMedia := models.Media{
			ID:        utils.GenerateUUID(),
			UserID:    userID,
			FileURI:   *gifURL,
			FileType:  "image/gif",
			FileSize:  0,
			Status:    models.MediaStatusApproved,
			CreatedAt: time.Now().UTC(),
		}
		if err := s.mediaRepo.Create(ctx, &gifMedia); err != nil {
			return nil, fmt.Errorf("create gif media: %w", err)
		}
		mediaID = &gifMedia.ID
	}

	// Validate shared post first — a shared post IS the message content.
	if sharedPostID != nil && *sharedPostID != "" {
		post, err := s.postRepo.FindByID(ctx, *sharedPostID)
		if err != nil || post.Status == models.PostStatusHidden || post.Status == models.PostStatusPrivate {
			return nil, fmt.Errorf("bài viết không tồn tại hoặc không khả dụng")
		}
	} else {
		if err := s.validation.ValidateSendMessage(content, emojiID, mediaID); err != nil {
			return nil, err
		}
	}

	if emojiID != nil && *emojiID != "" {
		ok, err := s.chatRepo.IsEmojiExists(ctx, *emojiID)
		if err != nil {
			return nil, errorsapp.Wrap(errorsapp.ErrCodeGCEmojiNotFound, err)
		}
		if !ok {
			return nil, errorsapp.New(errorsapp.ErrCodeGCEmojiNotFound)
		}
	}

	if mediaID != nil && *mediaID != "" {
		media, err := s.mediaRepo.GetByID(ctx, *mediaID)
		if err != nil {
			return nil, errorsapp.New(errorsapp.ErrCodeGCMediaNotFound)
		}
		if media.UserID != userID {
			return nil, errorsapp.New(errorsapp.ErrCodeGCMediaNotYours)
		}
	}

	if replyToMessageID != nil && *replyToMessageID != "" {
		parentMsg, err := s.chatRepo.FindMessageByID(ctx, *replyToMessageID)
		if err != nil {
			return nil, errorsapp.New(errorsapp.ErrCodeGCReplyNotFound)
		}
		if parentMsg.ChatID != chatID {
			return nil, errorsapp.New(errorsapp.ErrCodeGCReplyWrongChat)
		}
	}

	encryptionKey, err := s.chatRepo.GetEncryptionKey(ctx, chatID)
	if err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodeGCEncryptionKeyNotFound, err)
	}

	encryptedContent, err := utils.EncryptMessage(content, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("mã hóa tin nhắn thất bại: %w", err)
	}

	msg := models.NewMessage(chatID, userID, encryptedContent, mediaID, emojiID)
	msg.ID = utils.GenerateUUID()
	msg.CreatedAt = time.Now().UTC()
	msg.ReplyToMessageID = replyToMessageID
	if sharedPostID != nil && *sharedPostID != "" {
		msg.SharedPostID = sharedPostID
		msg.Type = "shared_post"
	}

	savedMsg, err := s.chatRepo.CreateMessage(ctx, &msg)
	if err != nil {
		return nil, err
	}

	participants, err := s.chatRepo.GetParticipantIDs(ctx, chatID)
	if err == nil && s.notifService != nil {
		for _, participantID := range participants {
			if participantID == userID {
				continue
			}
			memberSettings, _ := s.groupRepo.GetMemberSettings(ctx, chatID, participantID)
			if memberSettings != nil && !memberSettings.NotificationsEnabled {
				continue
			}
			_, _ = s.notifService.Create(
				ctx,
				participantID,
				&userID,
				models.NotificationTypeMessage,
				"đã gửi tin nhắn trong nhóm",
				nil,
				&userID,
				&chat.ID,
			)
		}
	}

	return savedMsg, nil
}

func (s *GroupMessageService) DeleteMessage(ctx context.Context, userID, messageID, mode string) (*models.Message, error) {
	msg, err := s.chatRepo.FindMessageByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	if _, err := s.ensureGroupMember(ctx, userID, msg.ChatID); err != nil {
		return nil, err
	}

	if err := s.validation.ValidateDeleteMessage(msg.SenderID, userID, mode); err != nil {
		return nil, err
	}
	if err := s.validation.ValidateDeleteMode(mode); err != nil {
		return nil, err
	}

	deleteForAll := strings.EqualFold(mode, "all")

	if deleteForAll {
		if msg.DeletedAt != nil {
			return nil, errorsapp.New(errorsapp.ErrCodeChatAlreadyDeleted)
		}
		deletedAt := time.Now().UTC()
		return s.chatRepo.UpdateMessageDeleteStatus(ctx, messageID, true, true, &deletedAt)
	}

	if msg.SenderID == userID {
		if msg.DeletedForSender {
			return nil, errorsapp.New(errorsapp.ErrCodeChatAlreadyDeleted)
		}
		return s.chatRepo.UpdateMessageDeleteStatus(ctx, messageID, true, false, nil)
	}

	if msg.DeletedForReceiver {
		return nil, errorsapp.New(errorsapp.ErrCodeChatAlreadyDeleted)
	}
	return s.chatRepo.UpdateMessageDeleteStatus(ctx, messageID, false, true, nil)
}

func (s *GroupMessageService) GetAllMessagesDecrypted(
	ctx context.Context,
	userID, chatID string,
) ([]models.Message, error) {
	_, err := s.ensureGroupMember(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}

	messages, err := s.chatRepo.GetMessages(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}

	encryptionKey, err := s.chatRepo.GetEncryptionKey(ctx, chatID)
	if err != nil {
		return nil, err
	}

	for i := range messages {
		decrypted, decErr := utils.DecryptMessage(messages[i].Content, encryptionKey)
		if decErr != nil {
			continue
		}
		messages[i].Content = decrypted
	}

	return messages, nil
}

func (s *GroupMessageService) GetAllMessagesRaw(ctx context.Context, userID, chatID string) ([]models.Message, error) {
	_, err := s.ensureGroupMember(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}

	return s.chatRepo.GetMessages(ctx, chatID, userID)
}

// Tìm theo plaintext bằng cách decrypt rồi lọc in-memory
// Lý do: content trong DB đang là ciphertext, search bằng LIKE trên DB sẽ không hiệu quả
func (s *GroupMessageService) SearchMessages(
	ctx context.Context,
	userID, chatID, keyword string,
) ([]models.Message, error) {
	if err := s.validation.ValidateSearchMessages(keyword); err != nil {
		return nil, err
	}

	messages, err := s.GetAllMessagesDecrypted(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}

	kw := strings.ToLower(strings.TrimSpace(keyword))
	filtered := make([]models.Message, 0)
	for _, msg := range messages {
		if strings.Contains(strings.ToLower(msg.Content), kw) {
			filtered = append(filtered, msg)
		}
	}

	return filtered, nil
}

func (s *GroupMessageService) ListGroupMemberIDs(ctx context.Context, userID, chatID string) ([]string, error) {
	_, err := s.ensureGroupMember(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}
	return s.chatRepo.GetParticipantIDs(ctx, chatID)
}

func (s *GroupMessageService) CreateSystemMessage(ctx context.Context, chatID, content string) (*models.Message, error) {
	msg := models.NewMessage(chatID, "SYSTEM", content, nil, nil)
	msg.ID = utils.GenerateUUID()
	msg.CreatedAt = time.Now().UTC()
	return s.chatRepo.CreateMessage(ctx, &msg)
}

func (s *GroupMessageService) DecryptMessageContent(ctx context.Context, msg *models.Message) error {
	if msg.Content == "" {
		return nil
	}
	key, err := s.chatRepo.GetEncryptionKey(ctx, msg.ChatID)
	if err != nil {
		return err
	}
	decrypted, err := utils.DecryptMessage(msg.Content, key)
	if err != nil {
		return err
	}
	msg.Content = decrypted
	return nil
}

// GetMemberProfiles loads display_name and avatar_uri for a list of user IDs within a group.
func (s *GroupMessageService) GetMemberProfiles(ctx context.Context, chatID string, userIDs []string) map[string]struct{ DisplayName, AvatarURI string } {
	type profile struct {
		UserID      string `gorm:"column:user_id"`
		DisplayName string `gorm:"column:display_name"`
		AvatarURI   string `gorm:"column:avatar_uri"`
	}

	if len(userIDs) == 0 {
		return nil
	}

	var profiles []profile
	err := s.groupRepo.GetMemberProfiles(ctx, chatID, userIDs, &profiles)
	if err != nil {
		return nil
	}

	result := make(map[string]struct{ DisplayName, AvatarURI string }, len(profiles))
	for _, p := range profiles {
		result[p.UserID] = struct{ DisplayName, AvatarURI string }{
			DisplayName: p.DisplayName,
			AvatarURI:   p.AvatarURI,
		}
	}
	return result
}

func (s *GroupMessageService) SetGroupCallRepository(repo *repository.GroupCallRepository) {
	s.groupCallRepo = repo
}

func (s *GroupMessageService) GetChatRepo() *repository.ChatRepository {
	return s.chatRepo
}

func (s *GroupMessageService) GetGroupCallsByChatID(ctx context.Context, userID, chatID string) ([]repository.GroupCallDocument, error) {
	if _, err := s.ensureGroupMember(ctx, userID, chatID); err != nil {
		return nil, err
	}
	if s.groupCallRepo == nil {
		return nil, nil
	}
	return s.groupCallRepo.FindByChatID(ctx, chatID)
}

func (s *GroupMessageService) LoadSharedPosts(ctx context.Context, messages []models.Message) map[string]*dto.SharedPostPayload {
	postIDs := make([]string, 0)
	seen := make(map[string]struct{})
	for _, m := range messages {
		if m.SharedPostID != nil && *m.SharedPostID != "" {
			if _, ok := seen[*m.SharedPostID]; !ok {
				postIDs = append(postIDs, *m.SharedPostID)
				seen[*m.SharedPostID] = struct{}{}
			}
		}
	}
	if len(postIDs) == 0 {
		return nil
	}
	posts, err := s.postRepo.FindByIDs(ctx, postIDs)
	if err != nil || len(posts) == 0 {
		return nil
	}
	result := make(map[string]*dto.SharedPostPayload, len(posts))
	for i := range posts {
		p := &posts[i]
		result[p.ID] = &dto.SharedPostPayload{
			ID:          p.ID,
			UserID:      p.UserID,
			Username:    p.Username,
			DisplayName: p.DisplayName,
			AvatarURI:   p.AvatarURI,
			Title:       p.Title,
			Content:     p.Content,
		}
		if len(p.Media) > 0 {
			result[p.ID].MediaURI = p.Media[0].FileURI
			result[p.ID].MediaType = p.Media[0].FileType
		}
	}
	return result
}
