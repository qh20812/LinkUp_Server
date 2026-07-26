package models

import "time"

type StoryInteract struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	StoryID    string    `gorm:"type:char(36);not null;index" json:"story_id"`
	UserID     string    `gorm:"type:char(36);not null;index" json:"user_id"`
	Type       string    `gorm:"type:varchar(20);not null" json:"type"` // react, reply, share
	EmojiID    *string   `gorm:"type:varchar(36)" json:"emoji_id,omitempty"`
	ClickCount int       `gorm:"type:int;default:1" json:"click_count"` // Đếm số click react (Tối đa 5)
	Content    string    `gorm:"type:text" json:"content,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func (StoryInteract) TableName() string {
	return "story_interacts"
}
