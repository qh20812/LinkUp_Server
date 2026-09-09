package services

import (
	"context"
	"fmt"
	"time"

	"linkup/dto"
	"linkup/models"
	"linkup/repository"
)

type E2EService struct {
	e2eRepo *repository.E2ERepository
	chatRepo *repository.ChatRepository
}

func NewE2EService(e2eRepo *repository.E2ERepository, chatRepo *repository.ChatRepository) *E2EService {
	return &E2EService{e2eRepo: e2eRepo, chatRepo: chatRepo}
}

func (s *E2EService) RegisterUserKey(ctx context.Context, userID, publicKey string) error {
	now := time.Now().UTC()
	key := models.UserE2EKey{
		UserID:    userID,
		PublicKey: publicKey,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.e2eRepo.UpsertUserKey(ctx, &key)
}

func (s *E2EService) GetUserKey(ctx context.Context, userID string) (*models.UserE2EKey, error) {
	return s.e2eRepo.GetUserKey(ctx, userID)
}

// StoreChatKeys stores the wrapped chat key for each participant. The caller
// must be a participant of every chat being registered.
func (s *E2EService) StoreChatKeys(ctx context.Context, callerID string, inputs []dto.ChatE2EKeyInput) error {
	if len(inputs) == 0 {
		return nil
	}

	now := time.Now().UTC()
	keys := make([]models.ChatE2EKey, 0, len(inputs))
	for _, in := range inputs {
		ok, err := s.chatRepo.IsUserParticipant(ctx, in.ChatID, callerID)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("not a participant of chat %s", in.ChatID)
		}

		keys = append(keys, models.ChatE2EKey{
			ChatID:     in.ChatID,
			UserID:     in.UserID,
			WrappedKey: in.WrappedKey,
			Nonce:      in.Nonce,
			CreatedAt:  now,
		})
	}
	return s.e2eRepo.UpsertChatKeys(ctx, keys)
}

func (s *E2EService) GetChatKey(ctx context.Context, userID, chatID string) (*dto.ChatE2EKeyResponse, error) {
	ok, err := s.chatRepo.IsUserParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("not a participant of chat %s", chatID)
	}

	key, err := s.e2eRepo.GetChatKey(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, nil
	}
	return &dto.ChatE2EKeyResponse{
		ChatID:     key.ChatID,
		WrappedKey: key.WrappedKey,
		Nonce:      key.Nonce,
	}, nil
}
