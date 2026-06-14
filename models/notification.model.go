package models

import "time"

type NotificationType string

type Notification struct {
	ID                int64            `json:"id" db:"id"`
	ReceiverID        int64            `json:"receiver_id" db:"receiver_id"`
	SenderID          *int64           `json:"sender_id,omitempty" db:"sender_id"`
	Type              NotificationType `json:"type" db:"type"`
	RedirectPostID    *int64           `json:"redirect_post_id,omitempty" db:"redirect_post_id"`
	RedirectUserID    *int64           `json:"redirect_user_id,omitempty" db:"redirect_user_id"`
	RedirectCommentID *int64           `json:"redirect_comment_id,omitempty" db:"redirect_comment_id"`
	Content           string           `json:"content" db:"content"`
	IsRead            bool             `json:"is_read" db:"is_read"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
}

func NewNotification(receiverID int64, senderID *int64, notifType NotificationType, content string) Notification {
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
