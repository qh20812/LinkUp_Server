package models

import "time"

type Message struct {
	ID        int64     `json:"id" db:"id"`
	ChatID    int64     `json:"chat_id" db:"chat_id"`
	SenderID  int64     `json:"sender_id" db:"sender_id"`
	Content   string    `json:"content" db:"content"`
	MediaID   *int64    `json:"media_id,omitempty" db:"media_id"`
	EmojiID   *int64    `json:"emoji_id,omitempty" db:"emoji_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func NewMessage(chatID, senderID int64, content string, mediaID, emojiID *int64) Message {
	return Message{ChatID: chatID, SenderID: senderID, Content: content, MediaID: mediaID, EmojiID: emojiID}
}
