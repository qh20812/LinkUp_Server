package services

import (
	"context"
	"errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"strings"
	"time"
)

type ChatService struct {
	chatRepo   *repository.ChatRepository
	friendRepo *repository.FriendRepository
	inviteRepo *repository.ChatInvitationRepository
}

func NewChatService(chatRepo *repository.ChatRepository, friendRepo *repository.FriendRepository, inviteRepo *repository.ChatInvitationRepository) *ChatService {
	return &ChatService{
		chatRepo:   chatRepo,
		friendRepo: friendRepo,
		inviteRepo: inviteRepo,
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
	if strings.TrimSpace(content) == "" && emojiID == nil && mediaID == nil {
		return nil, errors.New("nội dung tin nhắn không được để trống")
	}
	if len(content) > 2000 {
		return nil, errors.New("nội dung tin nhắn quá dài")
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

	msg := models.NewMessage(chatID, userID, content, mediaID, emojiID)
	msg.ID = utils.GenerateUUID()
	msg.CreatedAt = time.Now().UTC()

	return s.chatRepo.CreateMessage(ctx, &msg)
}

func (s *ChatService) GetOrCreateDirectChat(ctx context.Context, userID, targetUserID string) (*models.Chat, error) {
	chat, err := s.chatRepo.FindDirectChat(ctx, userID, targetUserID)
	if err == nil {
		return chat, nil
	}
	if err != repository.ErrChatNotFound {
		return nil, err
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

	return s.chatRepo.CreateDirectChat(ctx, &newChat, participants)
}

func (s *ChatService) RequestChatInvite(ctx context.Context, userID, targetUserID string) (*models.ChatInvite, error) {
	if userID == targetUserID {
		return nil, errors.New("không thể tự mời chính mình")
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
	return invite, nil
}

func (s *ChatService) ResponseChatInvite(ctx context.Context, userID, inviteID string, accept bool) (*models.Chat, error) {
	invite, err := s.inviteRepo.FindPendingByID(ctx, inviteID)
	if err != nil {
		return nil, err
	}
	if invite.TargetID != userID {
		return nil, errors.New("bạn không có quyền phản hồi lời mời này")
	}

	if !accept {
		if err := s.inviteRepo.UpdateStatus(ctx, inviteID, models.ChatInviteStatusDeclined, nil); err != nil {
			return nil, err
		}
		return nil, nil
	}

	chat, err := s.GetOrCreateDirectChat(ctx, invite.RequesterID, invite.TargetID)
	if err != nil {
		return nil, err
	}

	if err := s.inviteRepo.UpdateStatus(ctx, inviteID, models.ChatInviteStatusAccepted, &chat.ID); err != nil {
		return nil, err
	}

	return chat, nil
}
