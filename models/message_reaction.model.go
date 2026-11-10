package models

import "time"

// MessageReaction lưu 1 cảm xúc (emoji) của 1 user trên 1 tin nhắn.
// 1 user được thả NHIỀU emoji khác nhau trên cùng 1 tin (mỗi emoji 1 row);
// bấm lại đúng emoji đó = gỡ row tương ứng, không ảnh hưởng emoji khác.
type MessageReaction struct {
	MessageID string    `json:"message_id" gorm:"primaryKey;column:message_id;type:varchar(36)"`
	UserID    string    `json:"user_id" gorm:"primaryKey;column:user_id;type:varchar(36)"`
	EmojiID   string    `json:"emoji_id" gorm:"primaryKey;column:emoji_id;type:varchar(36)"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (MessageReaction) TableName() string {
	return "message_reactions"
}