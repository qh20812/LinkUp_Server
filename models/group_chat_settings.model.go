package models

import "time"

type GroupChatSettings struct {
	ChatID               string     `json:"chat_id" gorm:"primaryKey;column:chat_id"`
	AllowMemberAdd       bool       `json:"allow_member_add" gorm:"default:true"`
	LastAdminTransferAt  *time.Time `json:"last_admin_transfer_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
