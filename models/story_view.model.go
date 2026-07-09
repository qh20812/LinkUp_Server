package models

import "time"

type StoryView struct {
	ID      uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	StoryID string `gorm:"type:char(36);not null;index" json:"story_id"`
	UserID  string `gorm:"column:viewer_id;type:char(36);not null;index" json:"user_id"`

	ViewedAt time.Time `json:"viewed_at"`
}

func (StoryView) TableName() string {
	return "story_views"
}
