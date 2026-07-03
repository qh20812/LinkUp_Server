package models

import "time"

type GroupChatMute struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	ChatID    string     `json:"chat_id" gorm:"index"`
	UserID    string     `json:"user_id" gorm:"index"`
	MutedBy   string     `json:"muted_by"`
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func NewGroupChatMute(chatID, userID, mutedBy, reason string, expiresAt *time.Time) GroupChatMute {
	return GroupChatMute{
		ChatID:    chatID,
		UserID:    userID,
		MutedBy:   mutedBy,
		Reason:    reason,
		ExpiresAt: expiresAt,
	}
}
