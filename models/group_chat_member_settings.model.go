package models

import "time"

type GroupChatMemberSettings struct {
	ChatID               string    `json:"chat_id" gorm:"primaryKey;column:chat_id"`
	UserID               string    `json:"user_id" gorm:"primaryKey;column:user_id"`
	NotificationsEnabled bool      `json:"notifications_enabled" gorm:"default:true"`
	UpdatedAt            time.Time `json:"updated_at"`
}
