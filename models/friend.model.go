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
	ID         int64        `json:"id" db:"id"`
	SenderID   int64        `json:"sender_id" db:"sender_id"`
	ReceiverID int64        `json:"receiver_id" db:"receiver_id"`
	Status     FriendStatus `json:"status" db:"status"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
}

func NewFriend(senderID, receiverID int64, status FriendStatus) Friend {
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
