package models

import "time"

type Message struct {
	ID        string    `json:"id" db:"id"`
	ChatID    string    `json:"chat_id" db:"chat_id"`
	SenderID  string    `json:"sender_id" db:"sender_id"`
	Content   string    `json:"content" db:"content"`
	MediaID   *string   `json:"media_id,omitempty" db:"media_id"`
	EmojiID   *string   `json:"emoji_id,omitempty" db:"emoji_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func NewMessage(chatID, senderID string, content string, mediaID, emojiID *string) Message {
	return Message{ChatID: chatID, SenderID: senderID, Content: content, MediaID: mediaID, EmojiID: emojiID}
}
