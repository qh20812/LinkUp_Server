package models

import "time"

// StoryMute ghi nhận việc một user ẩn story của một user khác trên story bar
type StoryMute struct {
	ID          string    `gorm:"primaryKey;type:char(36)" json:"id"`
	UserID      string    `gorm:"type:char(36);not null;index" json:"user_id"`
	MutedUserID string    `gorm:"type:char(36);not null;index" json:"muted_user_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func (StoryMute) TableName() string {
	return "story_mutes"
}