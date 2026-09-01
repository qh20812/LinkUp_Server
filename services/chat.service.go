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
	chatRepo         *repository.ChatRepository
	friendRepo       *repository.FriendRepository
	inviteRepo       *repository.ChatInvitationRepository
	mediaRepo        *repository.MediaRepository
	postRepo         *repository.PostRepository
	userSettingsRepo *repository.UserSettingsRepository
	profileRepo      *repository.ProfileRepository
	notifService     *NotificationService
	validation       *validations.ChatValidation
}

func NewChatService(chatRepo *repository.ChatRepository, friendRepo *repository.FriendRepository, inviteRepo *repository.ChatInvitationRepository, mediaRepo *repository.MediaRepository, postRepo *repository.PostRepository, userSettingsRepo *repository.UserSettingsRepository, profileRepo *repository.ProfileRepository, notifService *NotificationService, validation *validations.ChatValidation) *ChatService {
	return &ChatService{
		chatRepo:         chatRepo,
		friendRepo:       friendRepo,
		inviteRepo:       inviteRepo,
		mediaRepo:        mediaRepo,
		postRepo:         postRepo,
		userSettingsRepo: userSettingsRepo,
		profileRepo:      profileRepo,
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

func (s *ChatService) SendMessage(ctx context.Context, userID, chatID, content string, e2eVersion int, emojiID, mediaID, gifURL, replyToMessageID, sharedPostID, mediaGroupID *string) (*models.Message, error) {
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

	// GIF ngoài (GIPHY/Tenor): tạo bản ghi media như bài viết, rồi gắn media_id
	// để tái sử dụng toàn bộ pipeline media sẵn có của tin nhắn (render, preview,
	// download). Không upload lên Cloudinary, chỉ lưu URL gốc.
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
	// When sharedPostID is provided, skip content/emoji/media validation.
	if sharedPostID != nil && *sharedPostID != "" {
		post, err := s.postRepo.FindByID(ctx, *sharedPostID)
		if err != nil || post.Status == models.PostStatusHidden || post.Status == models.PostStatusPrivate {
			return nil, fmt.Errorf("bài viết không tồn tại hoặc không khả dụng")
		}
	} else {
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
	msg.MediaGroupID = mediaGroupID
	if sharedPostID != nil && *sharedPostID != "" {
		msg.SharedPostID = sharedPostID
		msg.Type = "shared_post"
	}

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

func (s *ChatService) GetReplyPreviews(ctx context.Context, messageIDs []string) map[string]*dto.ReplyPreview {
	return s.chatRepo.GetReplyPreviews(ctx, messageIDs)
}

func (s *ChatService) GetMediaFileTypes(ctx context.Context, mediaIDs []string) map[string]string {
	ids := make([]string, 0, len(mediaIDs))
	seen := make(map[string]struct{}, len(mediaIDs))
	for _, id := range mediaIDs {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	result, err := s.mediaRepo.GetFileTypesByIDs(ctx, ids)
	if err != nil {
		return map[string]string{}
	}
	return result
}

func (s *ChatService) GetMessagesHistory(ctx context.Context, userID, chatID string, cursor *dto.HistoryCursor, limit int) ([]models.Message, error) {
	if err := s.JoinChat(ctx, userID, chatID); err != nil {
		return nil, err
	}

	var beforeCreatedAt *time.Time
	var beforeID string
	if cursor != nil {
		t := cursor.CreatedAt
		beforeCreatedAt = &t
		beforeID = cursor.ID
	}

	messages, err := s.chatRepo.GetMessagesPaged(ctx, chatID, beforeCreatedAt, beforeID, limit)
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

func (s *ChatService) GetEncryptionKey(ctx context.Context, chatID string) (string, error) {
	return s.chatRepo.GetEncryptionKey(ctx, chatID)
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

func (s *ChatService) ListReceivedInvites(ctx context.Context, userID string) ([]dto.ChatInviteItemDTO, error) {
	invites, err := s.inviteRepo.FindPendingByTarget(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(invites) == 0 {
		return []dto.ChatInviteItemDTO{}, nil
	}

	requesterIDs := make([]string, 0, len(invites))
	for _, inv := range invites {
		requesterIDs = append(requesterIDs, inv.RequesterID)
	}

	profileMap := make(map[string]dto.SenderProfile)
	if profiles, err := s.profileRepo.FindByIDs(ctx, requesterIDs); err == nil {
		for _, p := range profiles {
			name := p.DisplayName
			if name == "" {
				name = "User"
			}
			profileMap[p.UserID] = dto.SenderProfile{
				DisplayName: name,
				AvatarURI:   p.AvatarURI,
			}
		}
	}

	items := make([]dto.ChatInviteItemDTO, 0, len(invites))
	for _, inv := range invites {
		item := dto.ChatInviteItemDTO{
			InviteID:    inv.ID,
			RequesterID: inv.RequesterID,
			CreatedAt:   inv.CreatedAt,
		}
		if sender, ok := profileMap[inv.RequesterID]; ok {
			item.RequesterName = sender.DisplayName
			item.RequesterAvatar = sender.AvatarURI
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *ChatService) ResponseChatInvite(ctx context.Context, userID, inviteID string, accept bool) (*models.Chat, error) {	invite, err := s.inviteRepo.FindPendingByID(ctx, inviteID)
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

	// Chat legacy (e2e_version = 0) được server mã hóa bằng khóa chat nên có thể
	// giải mã và tìm kiếm. Chat E2E (e2e_version = 1) server không đọc được nội
	// dung — client tự tìm kiếm trên dữ liệu đã giải mã ở máy.
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
		// E2E messages là ciphertext không thể khớp — bỏ qua.
		if msg.E2EVersion != 0 {
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

// ── Pin Message ────────────────────────────────────────────────────────────

func (s *ChatService) PinMessage(ctx context.Context, userID, chatID, messageID string) (*dto.PinnedMessageDTO, error) {
	// Validate chat exists
	chat, err := s.chatRepo.FindChatByID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if chat.Type != models.ChatTypeDirect {
		return nil, fmt.Errorf("không phải chat trực tiếp")
	}

	// Validate user is participant
	ok, err := s.chatRepo.IsUserParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errorsapp.New(errorsapp.ErrCodeChatNotParticipant)
	}

	// Validate message exists in this chat
	msg, err := s.chatRepo.FindMessageByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("không tìm thấy tin nhắn")
	}
	if msg.ChatID != chatID {
		return nil, fmt.Errorf("tin nhắn không thuộc chat này")
	}
	if msg.DeletedAt != nil {
		return nil, fmt.Errorf("không thể ghim tin nhắn đã xóa")
	}

	// Check if already pinned
	pinned, err := s.chatRepo.IsMessagePinned(ctx, chatID, messageID)
	if err != nil {
		return nil, err
	}
	if pinned {
		return nil, fmt.Errorf("tin nhắn đã được ghim")
	}

	// Auto-unpin oldest if at max
	count, err := s.chatRepo.CountPinnedMessages(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if count >= 2 {
		if err := s.chatRepo.AutoUnpinOldest(ctx, chatID); err != nil {
			return nil, err
		}
	}

	pm, err := s.chatRepo.PinMessage(ctx, chatID, messageID, userID)
	if err != nil {
		return nil, err
	}

	// Decrypt content for the DTO
	content := msg.Content
	if msg.E2EVersion == 0 {
		decrypted, dErr := s.chatRepo.GetEncryptionKey(ctx, chatID)
		if dErr == nil && decrypted != "" {
			if dec, e := utils.DecryptMessage(content, decrypted); e == nil {
				content = dec
			}
		}
	}

	senderName := s.chatRepo.GetDisplayName(ctx, msg.SenderID)

	return &dto.PinnedMessageDTO{
		ID:         pm.ID,
		MessageID:  pm.MessageID,
		PinnedBy:   pm.PinnedBy,
		PinnedAt:   pm.PinnedAt,
		Content:    content,
		SenderID:   msg.SenderID,
		SenderName: senderName,
	}, nil
}

func (s *ChatService) UnpinMessage(ctx context.Context, userID, chatID, messageID string) error {
	// Validate chat exists
	if _, err := s.chatRepo.FindChatByID(ctx, chatID); err != nil {
		return err
	}

	// Validate user is participant
	ok, err := s.chatRepo.IsUserParticipant(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errorsapp.New(errorsapp.ErrCodeChatNotParticipant)
	}

	return s.chatRepo.UnpinMessage(ctx, chatID, messageID)
}

func (s *ChatService) GetPinnedMessages(ctx context.Context, userID, chatID string) ([]dto.PinnedMessageDTO, error) {
	// Validate chat exists
	if _, err := s.chatRepo.FindChatByID(ctx, chatID); err != nil {
		return nil, err
	}

	// Validate user is participant
	ok, err := s.chatRepo.IsUserParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errorsapp.New(errorsapp.ErrCodeChatNotParticipant)
	}

	pins, err := s.chatRepo.GetPinnedMessages(ctx, chatID)
	if err != nil {
		return nil, err
	}

	// Decrypt content for each pinned message
	encKey, _ := s.chatRepo.GetEncryptionKey(ctx, chatID)
	for i := range pins {
		if encKey != "" {
			if decrypted, e := utils.DecryptMessage(pins[i].Content, encKey); e == nil {
				pins[i].Content = decrypted
			}
		}
	}

	return pins, nil
}
