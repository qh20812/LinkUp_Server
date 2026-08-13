package services

import (
	"context"
	"fmt"
	"io"
	"linkup/dto"
	"linkup/models"
	errorsapp "linkup/errors"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"net/http"
	"strings"
	"time"
)

const errMySQLDuplicate = "Duplicate entry"

type ChatService struct {
	chatRepo        *repository.ChatRepository
	friendRepo      *repository.FriendRepository
	inviteRepo      *repository.ChatInvitationRepository
	mediaRepo       *repository.MediaRepository
	userSettingsRepo *repository.UserSettingsRepository
	notifService    *NotificationService
	validation      *validations.ChatValidation
}

func NewChatService(chatRepo *repository.ChatRepository, friendRepo *repository.FriendRepository, inviteRepo *repository.ChatInvitationRepository, mediaRepo *repository.MediaRepository, userSettingsRepo *repository.UserSettingsRepository, notifService *NotificationService, validation *validations.ChatValidation) *ChatService {
	return &ChatService{
		chatRepo:         chatRepo,
		friendRepo:       friendRepo,
		inviteRepo:       inviteRepo,
		mediaRepo:        mediaRepo,
		userSettingsRepo: userSettingsRepo,
		notifService:     notifService,
		validation:       validation,
	}
}

func (s *ChatService) JoinChat(ctx context.Context, userID, chatID string) error {
	ok, err := s.chatRepo.IsUserParticipant(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errorsapp.New(errorsapp.ErrCodeChatNotParticipant)
	}
	return nil
}

func (s *ChatService) SendMessage(ctx context.Context, userID, chatID, content string, e2eVersion int, emojiID, mediaID, replyToMessageID *string) (*models.Message, error) {
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

	// Với tin nhắn E2E (e2eVersion == 1), nội dung đã được client mã hóa đầu cuối,
	// server chỉ lưu ciphertext mà không thể (và không được phép) đọc. Bỏ qua
	// giới hạn độ dài plaintext vì ciphertext luôn dài hơn.
	validateErr := s.validation.ValidateSendMessage(content, emojiID, mediaID)
	if validateErr != nil {
		if e2eVersion != 1 {
			return nil, validateErr
		}
		if strings.TrimSpace(content) == "" && emojiID == nil && mediaID == nil {
			return nil, validateErr
		}
	}

	_, err = s.chatRepo.FindChatByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	participant, err := s.chatRepo.IsUserParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !participant {
		return nil, errorsapp.New(errorsapp.ErrCodeChatNotParticipant)
	}

	if emojiID != nil && *emojiID != "" {
		ok, err := s.chatRepo.IsEmojiExists(ctx, *emojiID)
		if err != nil {
			return nil, fmt.Errorf("check emoji: %w", err)
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

	msg := models.NewMessage(chatID, userID, content, mediaID, emojiID)
	msg.ID = utils.GenerateUUID()
	msg.CreatedAt = time.Now().UTC()
	msg.ReplyToMessageID = replyToMessageID

	if e2eVersion == 1 {
		// E2E: lưu ciphertext client-gửi nguyên trạng, server không mã hóa lại.
		msg.Content = content
		msg.E2EVersion = 1
	} else {
		// Legacy: server mã hóa bằng khóa chat như trước đây.
		encryptionKey, err := s.chatRepo.GetEncryptionKey(ctx, chatID)
		if err != nil {
			return nil, errorsapp.Wrap(errorsapp.ErrCodeGCEncryptionKeyNotFound, err)
		}

		encryptedContent, err := utils.EncryptMessage(content, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("encrypt message: %w", err)
		}
		msg.Content = encryptedContent
		msg.E2EVersion = 0
	}

	savedMsg, err := s.chatRepo.CreateMessage(ctx, &msg)
	if err != nil {
		return nil, err
	}

	participants, err := s.chatRepo.GetParticipantIDs(ctx, chatID)
	if err == nil {
		for _, participantID := range participants {
			if participantID != userID {
				s.notifService.Create(ctx, participantID, &userID, models.NotificationTypeMessage, "đã gửi tin nhắn", nil, &userID, &chatID)
			}
		}
	}

	return savedMsg, nil
}

func (s *ChatService) GetAllMessagesDecrypted(ctx context.Context, userID, chatID string) ([]models.Message, error) {
	messages, err := s.GetAllMessages(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}

	encryptionKey, err := s.chatRepo.GetEncryptionKey(ctx, chatID)
	if err != nil {
		return nil, err
	}

	for i := range messages {
		// Chỉ giải mã tin nhắn legacy (e2e_version = 0). Tin nhắn E2E
		// (e2e_version = 1) giữ nguyên ciphertext — client tự giải mã.
		if messages[i].E2EVersion != 0 {
			continue
		}
		decrypted, err := utils.DecryptMessage(messages[i].Content, encryptionKey)
		if err != nil {
			fmt.Printf("failed to decrypt message %s: %v\n", messages[i].ID, err)
			continue
		}
		messages[i].Content = decrypted
	}

	return messages, nil
}

func (s *ChatService) GetOrCreateDirectChat(ctx context.Context, userID, targetUserID string) (*models.Chat, bool, error) {
	if err := s.validation.ValidateDirectChat(userID, targetUserID); err != nil {
		return nil, false, err
	}

	// A stranger may start a direct chat only if the target allows it.
	requiredFriendship := true
	if setting, err := s.userSettingsRepo.GetByUserID(ctx, targetUserID); err == nil && setting != nil && setting.AllowStrangerMessages {
		requiredFriendship = false
	}

	return s.ensureDirectChat(ctx, userID, targetUserID, requiredFriendship)
}

func (s *ChatService) ensureDirectChat(ctx context.Context, userID, targetUserID string, requiredFriendship bool) (*models.Chat, bool, error) {
	chat, err := s.chatRepo.FindDirectChat(ctx, userID, targetUserID)
	if err == nil {
		return chat, true, nil
	}
	if err != repository.ErrChatNotFound {
		return nil, false, err
	}

	if requiredFriendship {
		isFriend, err := s.friendRepo.IsAcceptedFriend(ctx, userID, targetUserID)
		if err != nil {
			return nil, false, err
		}
		if !isFriend {
			return nil, false, errorsapp.New(errorsapp.ErrCodeChatNotParticipant)
		}
	}

	newChat := models.Chat{
		ID:        utils.GenerateUUID(),
		Type:      models.ChatTypeDirect,
		Name:      "",
		AvatarURI: "",
		CreatedAt: time.Now().UTC(),
	}

	participants := []models.ChatParticipant{
		{ID: utils.GenerateUUID(), ChatID: newChat.ID, UserID: userID, JoinedAt: time.Now().UTC()},
		{ID: utils.GenerateUUID(), ChatID: newChat.ID, UserID: targetUserID, JoinedAt: time.Now().UTC()},
	}

	createdChat, err := s.chatRepo.CreateDirectChat(ctx, &newChat, participants)
	if err != nil {
		if strings.Contains(err.Error(), errMySQLDuplicate) {
			existing, findErr := s.chatRepo.FindDirectChat(ctx, userID, targetUserID)
			if findErr != nil {
				return nil, false, findErr
			}
			return existing, true, nil
		}
		return nil, false, err
	}
	return createdChat, false, nil
}

func (s *ChatService) RequestChatInvite(ctx context.Context, userID, targetUserID string) (*models.ChatInvite, error) {
	if err := s.validation.ValidateRequestChatInvite(userID, targetUserID); err != nil {
		return nil, err
	}

	isFriend, err := s.friendRepo.IsAcceptedFriend(ctx, userID, targetUserID)
	if err != nil {
		return nil, err
	}
	if isFriend {
		return nil, errorsapp.New(errorsapp.ErrCodeChatAlreadyFriends)
	}

	pending, err := s.inviteRepo.FindPendingBetween(ctx, userID, targetUserID)
	if err != nil {
		return nil, err
	}
	if pending != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeChatInvitePending)
	}

	existing, err := s.inviteRepo.FindActiveBetween(ctx, userID, targetUserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		switch existing.Status {
		case models.ChatInviteStatusPending:
			return nil, errorsapp.New(errorsapp.ErrCodeChatInvitePending)
		case models.ChatInviteStatusAccepted:
			return nil, errorsapp.New(errorsapp.ErrCodeChatInviteAccepted)
		}
	}

	invite := &models.ChatInvite{
		ID:          utils.GenerateUUID(),
		RequesterID: userID,
		TargetID:    targetUserID,
		Status:      models.ChatInviteStatusPending,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, err
	}

	s.notifService.Create(ctx, targetUserID, &userID, models.NotificationTypeMessage, "đã mời bạn tham gia chat", nil, &userID, nil)

	return invite, nil
}

func (s *ChatService) ResponseChatInvite(ctx context.Context, userID, inviteID string, accept bool) (*models.Chat, error) {
	invite, err := s.inviteRepo.FindPendingByID(ctx, inviteID)
	if err != nil {
		return nil, err
	}
	if err := s.validation.ValidateResponseChatInvite(userID, invite.TargetID); err != nil {
		return nil, err
	}

	if !accept {
		if err := s.inviteRepo.UpdateStatus(ctx, inviteID, models.ChatInviteStatusDeclined, nil); err != nil {
			return nil, err
		}
		return nil, nil
	}

	chat, _, err := s.ensureDirectChat(ctx, invite.RequesterID, invite.TargetID, false)
	if err != nil {
		return nil, err
	}

	if err := s.inviteRepo.UpdateStatus(ctx, inviteID, models.ChatInviteStatusAccepted, &chat.ID); err != nil {
		return nil, err
	}

	s.notifService.Create(ctx, invite.RequesterID, &userID, models.NotificationTypeMessage, "đã chấp nhận lời mời chat", nil, &userID, nil)

	return chat, nil
}

func (s *ChatService) DeleteMessage(ctx context.Context, userID, messageID, mode string) (*models.Message, error) {
	msg, err := s.chatRepo.FindMessageByID(ctx, messageID)
	if err != nil {
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

	// Delete for the requesting user only ("me") — works for both sender and receiver.
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

func (s *ChatService) GetAllMessages(ctx context.Context, userID, chatID string) ([]models.Message, error) {
	if err := s.JoinChat(ctx, userID, chatID); err != nil {
		return nil, err
	}
	return s.chatRepo.GetMessages(ctx, chatID, userID)
}

func (s *ChatService) SearchMessages(ctx context.Context, userID, chatID, keyword string) ([]models.Message, error) {
	if err := s.validation.ValidateSearchMessages(keyword); err != nil {
		return nil, err
	}
	if err := s.JoinChat(ctx, userID, chatID); err != nil {
		return nil, err
	}

	// Chat trực tiếp dùng E2E — server không đọc được nội dung nên trả về trống;
	// client tự tìm kiếm trên dữ liệu đã giải mã ở máy.
	chat, err := s.chatRepo.FindChatByID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if chat.Type == models.ChatTypeDirect {
		return []models.Message{}, nil
	}

	messages, err := s.GetAllMessagesDecrypted(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}

	needle := strings.ToLower(keyword)
	results := make([]models.Message, 0, len(messages))
	for _, msg := range messages {
		if msg.DeletedForSender || msg.DeletedForReceiver {
			continue
		}
		if strings.Contains(strings.ToLower(msg.Content), needle) {
			results = append(results, msg)
		}
	}
	return results, nil
}

func (s *ChatService) DecryptMessage(ctx context.Context, chatID, encryptedContent string) (string, error) {
	encryptionKey, err := s.chatRepo.GetEncryptionKey(ctx, chatID)
	if err != nil {
		return "", err
	}
	return utils.DecryptMessage(encryptedContent, encryptionKey)
}

func (s *ChatService) DownloadMessageMedia(ctx context.Context, userID, messageID string) (*models.Media, string, string, []byte, error) {
	message, err := s.chatRepo.FindMessageByID(ctx, messageID)
	if err != nil {
		return nil, "", "", nil, errorsapp.New(errorsapp.ErrCodeChatMessageNotFound)
	}

	isParticipant, err := s.chatRepo.IsUserParticipant(ctx, message.ChatID, userID)
	if err != nil {
		return nil, "", "", nil, err
	}
	if !isParticipant {
		return nil, "", "", nil, errorsapp.New(errorsapp.ErrCodeChatAccessDenied)
	}

	if message.MediaID == nil || *message.MediaID == "" {
		return nil, "", "", nil, errorsapp.New(errorsapp.ErrCodeGCMediaNotFound)
	}

	media, err := s.mediaRepo.GetByID(ctx, *message.MediaID)
	if err != nil {
		return nil, "", "", nil, errorsapp.New(errorsapp.ErrCodeGCMediaNotFound)
	}

	resp, err := http.Get(media.FileURI)
	if err != nil {
		return nil, "", "", nil, fmt.Errorf("download file from storage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", "", nil, fmt.Errorf("download file failed: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", "", nil, fmt.Errorf("read file content: %w", err)
	}

	contentType := media.FileType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	filename := fmt.Sprintf("%s%s", media.ID, extensionFromContentType(contentType))

	return media, contentType, filename, data, nil
}

func extensionFromContentType(contentType string) string {
	switch strings.ToLower(contentType) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "iamge/webp":
		return ".webp"
	case "video/mp4":
		return ".mp4"
	case "video/quicktime":
		return ".mov"
	case "video/x-msvideo":
		return ".avi"
	case "video/x-matroska":
		return ".mkv"
	case "video/webm":
		return ".webm"
	default:
		return ".bin"
	}
}

func (s *ChatService) ListUserChats(ctx context.Context, userID string) ([]dto.ChatConversationDTO, error) {
	chats, err := s.chatRepo.ListUserChats(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range chats {
		if chats[i].LastMessage == nil || chats[i].LastMessage.Content == "" {
			continue
		}
		// Tin nhắn E2E giữ nguyên ciphertext (client tự giải mã để hiện preview);
		// chỉ giải mã tin nhắn legacy (e2e_version = 0).
		if chats[i].LastMessage.E2EVersion != 0 {
			continue
		}
		key, keyErr := s.chatRepo.GetEncryptionKey(ctx, chats[i].ChatID)
		if keyErr != nil {
			continue
		}
		decrypted, decErr := utils.DecryptMessage(chats[i].LastMessage.Content, key)
		if decErr == nil {
			chats[i].LastMessage.Content = decrypted
		}
	}

	return chats, nil
}

func (s *ChatService) DeleteChat(ctx context.Context, userID, chatID string) error {
	if _, err := s.chatRepo.FindChatByID(ctx, chatID); err != nil {
		return err
	}

	participant, err := s.chatRepo.IsUserParticipant(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if !participant {
		return errorsapp.New(errorsapp.ErrCodeChatNotParticipant)
	}

	return s.chatRepo.DeleteChat(ctx, chatID)
}
