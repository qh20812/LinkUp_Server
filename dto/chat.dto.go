package dto

import (
	"encoding/json"
	"time"
)

type WsEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type ChatJoinPayload struct {
	ChatID string `json:"chat_id"`
}

type SendMessagePayload struct {
	ChatID  string  `json:"chat_id,omitempty"`
	Content string  `json:"content"`
	EmojiID *string `json:"emoji_id,omitempty"`
	MediaID *string `json:"media_id,omitempty"`
}

type MessagePayload struct {
	ID            string    `json:"id"`
	ChatID        string    `json:"chat_id"`
	SenderID      string    `json:"sender_id"`
	Content       string    `json:"content"`
	EmojiID       *string   `json:"emoji_id,omitempty"`
	MediaID       *string   `json:"media_id,omitempty"`
	IsAnonymized  bool      `json:"is_anonymized"`
	AnonymousName *string   `json:"anonymous_name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type TypingPayload struct {
	ChatID   string `json:"chat_id"`
	UserID   string `json:"user_id"`
	IsTyping bool   `json:"is_typing"`
}

type DirectChatRequest struct {
	TargetUserID string `json:"target_user_id" binding:"required"`
}

type DirectChatResponse struct {
	ChatID  string `json:"chat_id"`
	Message string `json:"message,omitempty"`
}

type ChatInviteRequest struct {
	TargetUserID string `json:"target_user_id" binding:"required"`
}

type ChatInviteResponseRequest struct {
	InviteID string `json:"invite_id" binding:"required"`
	Accept   bool   `json:"accept"`
}

type ChatInviteResponse struct {
	InviteID string  `json:"invite_id"`
	ChatID   *string `json:"chat_id,omitempty"`
	Message  string  `json:"message"`
}

type DeleteMessagePayload struct {
	ChatID    string `json:"chat_id"`
	MessageID string `json:"message_id"`
	Mode      string `json:"mode"`
}

type MessageDeletedPayload struct {
	ChatID    string `json:"chat_id"`
	MessageID string `json:"message_id"`
	DeletedBy string `json:"deleted_by"`
	Mode      string `json:"mode"`
}

type SearchMessagePayload struct {
	ChatID  string `json:"chat_id"`
	Keyword string `json:"keyword"`
}

type SearchMessageResultPayload struct {
	ChatID   string           `json:"chat_id"`
	Keyword  string           `json:"keyword"`
	Messages []MessagePayload `json:"messages"`
}
