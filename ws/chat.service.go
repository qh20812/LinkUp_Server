package ws

import (
	"context"
	"time"

	"linkup/dto"
	"linkup/models"
)

type ChatService interface {
	JoinChat(ctx context.Context, userID, chatID string) error
	SendMessage(ctx context.Context, userID, chatID, content string, e2eVersion int, emojiID, mediaID, gifURL, replyToMessageID, sharedPostID, mediaGroupID, forwardedFrom *string) (*models.Message, error)
	GetAllMessages(ctx context.Context, userID, chatID string) ([]models.Message, error)
	GetMessagesHistory(ctx context.Context, userID, chatID string, cursor *dto.HistoryCursor, limit int) ([]models.Message, error)
	DeleteMessage(ctx context.Context, userID, messageID, mode string) (*models.Message, error)
	SearchMessages(ctx context.Context, userID, chatID, keyword string) ([]models.Message, error)
	GetAllMessagesDecrypted(ctx context.Context, userID, chatID string) ([]models.Message, error)
	DecryptMessage(ctx context.Context, chatID, encryptedContent string) (string, error)
	GetEncryptionKey(ctx context.Context, chatID string) (string, error)
	GetReplyPreviews(ctx context.Context, messageIDs []string) map[string]*dto.ReplyPreview
	GetMediaFileTypes(ctx context.Context, mediaIDs []string) map[string]string
	GetMediaDurations(ctx context.Context, mediaIDs []string) map[string]int
	PinMessage(ctx context.Context, userID, chatID, messageID string) (*dto.PinnedMessageDTO, error)
	UnpinMessage(ctx context.Context, userID, chatID, messageID string) error
	GetPinnedMessages(ctx context.Context, userID, chatID string) ([]dto.PinnedMessageDTO, error)
	MarkMessageRead(ctx context.Context, userID, chatID, messageID string) (*dto.MessageReadStatePayload, error)
	GetChatReadWatermarks(ctx context.Context, chatID string) (map[string]time.Time, error)
	ReactToMessage(ctx context.Context, userID, chatID, messageID, emojiID string) (action string, reactions []dto.MessageReactionPayload, err error)
	GetMessageReactions(ctx context.Context, messageIDs []string) map[string][]dto.MessageReactionPayload
}
