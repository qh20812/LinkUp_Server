package models

import (
	"strings"
	"time"
)

type FriendStatus string

const (
	FriendStatusPending  FriendStatus = "pending"
	FriendStatusAccepted FriendStatus = "accepted"
	FriendStatusRejected FriendStatus = "rejected"
)

type Friend struct {
	ID         string        `json:"id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Status     FriendStatus `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
}

func NewFriend(senderID, receiverID string, status FriendStatus) Friend {
	if status == "" {
		status = FriendStatusPending
	}
	return Friend{SenderID: senderID, ReceiverID: receiverID, Status: status}
}

func (s FriendStatus) String() string {
	return string(s)
}

func ParseFriendStatus(value string) FriendStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(FriendStatusPending):
		return FriendStatusPending
	case string(FriendStatusAccepted):
		return FriendStatusAccepted
	case string(FriendStatusRejected):
		return FriendStatusRejected
	default:
		return FriendStatusPending
	}
}
