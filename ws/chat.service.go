package ws

import (
	"context"
	"linkup/models"
)

type ChatService interface {
	JoinChat(ctx context.Context, userID, chatID string) error
	SendMessage(ctx context.Context, userID, chatID, content string, emojiID, mediaID *string) (*models.Message, error)
	GetAllMessages(ctx context.Context, userID, chatID string) ([]models.Message, error)
	DeleteMessage(ctx context.Context, userID, messageID, mode string) (*models.Message, error)
	SearchMessages(ctx context.Context, userID, chatID, keyword string) ([]models.Message, error)
	GetAllMessagesDecrypted(ctx context.Context, userID, chatID string) ([]models.Message, error)
}
