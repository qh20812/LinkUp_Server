package models

import (
	"strings"
	"time"
)

type JoinRequestStatus string

const (
	JoinRequestStatusPending  JoinRequestStatus = "pending"
	JoinRequestStatusApproved JoinRequestStatus = "approved"
	JoinRequestStatusRejected JoinRequestStatus = "rejected"
)

type CommunityJoinRequest struct {
	ID          string            `json:"id"`
	CommunityID string            `json:"community_id"`
	UserID      string            `json:"user_id"`
	Status      JoinRequestStatus `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	RespondedAt *time.Time        `json:"responded_at,omitempty"`
}

func NewCommunityJoinRequest(communityID, userID string) CommunityJoinRequest {
	return CommunityJoinRequest{
		CommunityID: communityID,
		UserID:      userID,
		Status:      JoinRequestStatusPending,
	}
}

func ParseJoinRequestStatus(value string) JoinRequestStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(JoinRequestStatusApproved):
		return JoinRequestStatusApproved
	case string(JoinRequestStatusRejected):
		return JoinRequestStatusRejected
	default:
		return JoinRequestStatusPending
	}
}
