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
	ChatID           string  `json:"chat_id,omitempty"`
	Content          string  `json:"content"`
	EmojiID          *string `json:"emoji_id,omitempty"`
	MediaID          *string `json:"media_id,omitempty"`
	GifURL           *string `json:"gif_url,omitempty"`
	ReplyToMessageID *string `json:"reply_to_message_id,omitempty"`
	SharedPostID     *string `json:"shared_post_id,omitempty"`
	E2EVersion       int     `json:"e2e_version,omitempty"`
}

type MessagePayload struct {
	ID               string              `json:"id"`
	ChatID           string              `json:"chat_id"`
	SenderID         string              `json:"sender_id"`
	Content          string              `json:"content"`
	EmojiID          *string             `json:"emoji_id,omitempty"`
	MediaID          *string             `json:"media_id,omitempty"`
	ReplyToMessageID *string             `json:"reply_to_message_id,omitempty"`
	ReplyTo          *ReplyPreview       `json:"reply_to,omitempty"`
	SharedPostID     *string             `json:"shared_post_id,omitempty"`
	SharedPost       *SharedPostPayload  `json:"shared_post,omitempty"`
	SenderName       string              `json:"sender_name,omitempty"`
	SenderAvatar     string              `json:"sender_avatar,omitempty"`
	Type             string              `json:"type,omitempty"`
	MessageCategory  string              `json:"message_category,omitempty"`
	IsAnonymized     bool                `json:"is_anonymized"`
	AnonymousName    *string             `json:"anonymous_name,omitempty"`
	E2EVersion       int                 `json:"e2e_version,omitempty"`
	Deleted          bool                `json:"deleted"`
	CreatedAt        time.Time           `json:"created_at"`
}

type ReplyPreview struct {
	ID           string `json:"id"`
	Content      string `json:"content"`
	SenderID     string `json:"sender_id"`
	SenderName   string `json:"sender_name"`
	SenderAvatar string `json:"sender_avatar"`
}

type SharedPostPayload struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	AvatarURI   string  `json:"avatar_uri"`
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	MediaURI    string  `json:"media_uri,omitempty"`
	MediaType   string  `json:"media_type,omitempty"`
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

type ChatPartnerDTO struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	AvatarURI   string `json:"avatar_uri"`
}

type ChatConversationDTO struct {
	ChatID      string          `json:"chat_id"`
	Partner     ChatPartnerDTO  `json:"partner"`
	LastMessage *MessagePayload `json:"last_message,omitempty"`
	IsEncrypted bool            `json:"is_encrypted"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ChatListResponse struct {
	Data []ChatConversationDTO `json:"data"`
}

type ChatInviteItemDTO struct {
	InviteID        string    `json:"invite_id"`
	RequesterID     string    `json:"requester_id"`
	RequesterName   string    `json:"requester_name,omitempty"`
	RequesterAvatar string    `json:"requester_avatar,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type ChatInviteListResponse struct {
	Data []ChatInviteItemDTO `json:"data"`
}
