package models

import (
	"time"
)

type ChatInviteStatus string

const (
	ChatInviteStatusPending  ChatInviteStatus = "pending"
	ChatInviteStatusAccepted ChatInviteStatus = "accepted"
	ChatInviteStatusDeclined ChatInviteStatus = "declined"
)

type ChatInvite struct {
	ID          string           `json:"id" db:"id"`
	RequesterID string           `json:"requester_id" db:"requester_id"`
	TargetID    string           `json:"target_id" db:"target_id"`
	ChatID      *string          `json:"chat_id,omitempty" db:"chat_id"`
	Status      ChatInviteStatus `json:"status" db:"status"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}

func (ChatInvite) TableName() string {
	return "chat_invitations"
}
