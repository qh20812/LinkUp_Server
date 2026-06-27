package ws

import (
	"context"
	"linkup/models"
)

type ChatService interface {
	JoinChat(ctx context.Context, userID, chatID string) error
	SendMessage(ctx context.Context, userID, chatID, content string, emojiID, mediaID *string) (*models.Message, error)
}
