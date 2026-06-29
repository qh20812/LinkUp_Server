package services

import (
	"context"
	"errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"strings"
	"time"
)

const errMySQLDuplicate = "Duplicate entry"

type ChatService struct {
	chatRepo     *repository.ChatRepository
	friendRepo   *repository.FriendRepository
	inviteRepo   *repository.ChatInvitationRepository
	notifService *NotificationService
	validation   *validations.ChatValidation
}

func NewChatService(chatRepo *repository.ChatRepository, friendRepo *repository.FriendRepository, inviteRepo *repository.ChatInvitationRepository, notifService *NotificationService, validation *validations.ChatValidation) *ChatService {
	return &ChatService{
		chatRepo:     chatRepo,
		friendRepo:   friendRepo,
		inviteRepo:   inviteRepo,
		notifService: notifService,
		validation:   validation,
	}
}

func (s *ChatService) JoinChat(ctx context.Context, userID, chatID string) error {
	ok, err := s.chatRepo.IsUserParticipant(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("bạn không có quyền tham gia chat này")
	}
	return nil
}

func (s *ChatService) SendMessage(ctx context.Context, userID, chatID, content string, emojiID, mediaID *string) (*models.Message, error) {
	if err := s.validation.ValidateSendMessage(content, emojiID, mediaID); err != nil {
		return nil, err
	}

	_, err := s.chatRepo.FindChatByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	participant, err := s.chatRepo.IsUserParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !participant {
		return nil, errors.New("bạn không tham gia chat này")
	}

	if mediaID != nil {
		owned, err := s.chatRepo.IsMediaOwnedByUser(ctx, *mediaID, userID)
		if err != nil {
			return nil, err
		}
		if !owned {
			return nil, errors.New("media không thuộc về bạn")
		}
	}

	msg := models.NewMessage(chatID, userID, content, mediaID, emojiID)
	msg.ID = utils.GenerateUUID()
	msg.CreatedAt = time.Now().UTC()

	savedMsg, err := s.chatRepo.CreateMessage(ctx, &msg)
	if err != nil {
		return nil, err
	}

	participants, err := s.chatRepo.GetParticipantIDs(ctx, chatID)
	if err == nil {
		preview := content
		if len(preview) > 100 {
			preview = preview[:100]
		}
		for _, pid := range participants {
			if pid != userID {
				s.notifService.Create(ctx, pid, &userID, models.NotificationTypeMessage, preview, nil, nil, nil)
			}
		}
	}

	return savedMsg, nil
}

func (s *ChatService) GetOrCreateDirectChat(ctx context.Context, userID, targetUserID string) (*models.Chat, error) {
	if err := s.validation.ValidateDirectChat(userID, targetUserID); err != nil {
		return nil, err
	}

	chat, err := s.chatRepo.FindDirectChat(ctx, userID, targetUserID)
	if err == nil {
		return chat, nil, true
	}
	if err != repository.ErrChatNotFound {
		return nil, err
	}

	isFriend, err := s.friendRepo.IsAcceptedFriend(ctx, userID, targetUserID)
	if err != nil {
		return nil, false, err
	}
	if !isFriend {
		return nil, errors.New("chưa là bạn bè, vui lòng gửi yêu cầu chat")
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
			return s.chatRepo.FindDirectChat(ctx, userID, targetUserID)
		}
		return nil, err
	}
	return createdChat, nil
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
		return nil, errors.New("đã là bạn, vui lòng mở chat trực tiếp")
	}

	pending, err := s.inviteRepo.FindPendingBetween(ctx, userID, targetUserID)
	if err != nil {
		return nil, err
	}
	if pending != nil {
		return pending, nil
	}

	existing, err := s.inviteRepo.FindActiveBetween(ctx, userID, targetUserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		switch existing.Status {
		case models.ChatInviteStatusPending:
			return nil, errors.New("lời mời đang chờ phản hổi")
		case models.ChatInviteStatusAccepted:
			return nil, errors.New("đã có lời mời được chấp nhận, không thể gửi")
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

	chat, _, err := s.GetOrCreateDirectChat(ctx, invite.RequesterID, invite.TargetID)
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

	if err := s.validation.ValidateDeleteMessage(msg.SenderID, userID, msg.DeletedForSender, msg.DeletedForReceiver); err != nil {
		return nil, err
	}
	if err := s.validation.ValidateDeleteMode(mode); err != nil {
		return nil, err
	}

	deleteForAll := strings.EqualFold(mode, "all")

	if deleteForAll {
		deletedAt := time.Now().UTC()
		return s.chatRepo.UpdateMessageDeleteStatus(ctx, messageID, true, true, &deletedAt)
	}

	return s.chatRepo.UpdateMessageDeleteStatus(ctx, messageID, true, false, nil)
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
	return s.chatRepo.SearchMessages(ctx, chatID, userID, keyword)
}
