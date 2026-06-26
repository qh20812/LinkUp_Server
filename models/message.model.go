package models

import "time"

type Message struct {
	ID        string    `json:"id"`
	ChatID    string    `json:"chat_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	MediaID   *string   `json:"media_id,omitempty"`
	EmojiID   *string   `json:"emoji_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func NewMessage(chatID, senderID string, content string, mediaID, emojiID *string) Message {
	return Message{ChatID: chatID, SenderID: senderID, Content: content, MediaID: mediaID, EmojiID: emojiID}
}
