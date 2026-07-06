package models

import "time"

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusDeclined InvitationStatus = "declined"
)

type CommunityInvitation struct {
	ID          string           `json:"id" gorm:"primaryKey"`
	CommunityID string           `json:"community_id" gorm:"type:varchar(36);index;not null"`
	InviterID   string           `json:"inviter_id" gorm:"type:varchar(36);not null"`
	InviteeID   string           `json:"invitee_id" gorm:"type:varchar(36);index;not null"`
	Status      InvitationStatus `json:"status" gorm:"type:varchar(20);default:pending"`
	CreatedAt   time.Time        `json:"created_at"`
	RespondedAt *time.Time       `json:"responded_at,omitempty"`
}
