package models

import (
	"strings"
	"time"
)

type NotificationType string

const (
	NotificationTypeLike          NotificationType = "like"
	NotificationTypeComment       NotificationType = "comment"
	NotificationTypeFollow        NotificationType = "follow"
	NotificationTypeMessage       NotificationType = "message"
	NotificationTypeFriendRequest NotificationType = "friend_request"
)

type Notification struct {
	ID                string           `json:"id" db:"id"`
	ReceiverID        string           `json:"receiver_id" db:"receiver_id"`
	SenderID          *string          `json:"sender_id,omitempty" db:"sender_id"`
	Type              NotificationType `json:"type" db:"type"`
	RedirectPostID    *string          `json:"redirect_post_id,omitempty" db:"redirect_post_id"`
	RedirectUserID    *string          `json:"redirect_user_id,omitempty" db:"redirect_user_id"`
	RedirectCommentID *string          `json:"redirect_comment_id,omitempty" db:"redirect_comment_id"`
	Content           string           `json:"content" db:"content"`
	IsRead            bool             `json:"is_read" db:"is_read"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
}

func NewNotification(receiverID string, senderID *string, notifType NotificationType, content string) Notification {
	return Notification{
		ReceiverID: receiverID,
		SenderID:   senderID,
		Type:       notifType,
		Content:    content,
	}
}

func (n NotificationType) String() string {
	return string(n)
}

func ParseNotificationType(value string) NotificationType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(NotificationTypeLike):
		return NotificationTypeLike
	case string(NotificationTypeComment):
		return NotificationTypeComment
	case string(NotificationTypeFollow):
		return NotificationTypeFollow
	case string(NotificationTypeMessage):
		return NotificationTypeMessage
	case string(NotificationTypeFriendRequest):
		return NotificationTypeFriendRequest
	default:
		return NotificationTypeLike
	}
}
