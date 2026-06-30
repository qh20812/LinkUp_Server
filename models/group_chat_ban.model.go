package models

import "time"

type GroupChatBan struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	ChatID    string    `json:"chat_id" gorm:"index"`
	UserID    string    `json:"user_id" gorm:"index"`
	BannedBy  string    `json:"banned_by"`
	CreatedAt time.Time `json:"created_at"`
}