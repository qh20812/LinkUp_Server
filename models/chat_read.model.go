package models

import "time"

// ChatRead lưu vị trí đã đọc (watermark) của từng thành viên trong một chat.
// Dùng cho read receipts: một tin nhắn được xem bởi user U nếu
// LastReadAt(U) >= CreatedAt(tin nhắn). Một dòng / (chat, user) — đơn giản và
// đủ cho cả chat 1-1 lẫn group.
type ChatRead struct {
	ChatID         string    `json:"chat_id" gorm:"column:chat_id;primaryKey"`
	UserID         string    `json:"user_id" gorm:"column:user_id;primaryKey"`
	LastReadAt     time.Time `json:"last_read_at" gorm:"column:last_read_at"`
	LastMessageID  string    `json:"last_message_id" gorm:"column:last_message_id"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (ChatRead) TableName() string {
	return "chat_reads"
}