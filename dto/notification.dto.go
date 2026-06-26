package dto

import (
	"linkup/models"
	"time"
)

type NotificationResponse struct {
	ID                string                  `json:"id"`
	SenderID          *string                 `json:"sender_id,omitempty"`
	Type              models.NotificationType `json:"type"`
	RedirectPostID    *string                 `json:"redirect_post_id,omitempty"`
	RedirectUserID    *string                 `json:"redirect_user_id,omitempty"`
	RedirectCommentID *string                 `json:"redirect_comment_id,omitempty"`
	Content           string                  `json:"content"`
	IsRead            bool                    `json:"is_read"`
	CreatedAt         time.Time               `json:"created_at"`
}

func ToNotificationResponse(n *models.Notification) NotificationResponse {
	return NotificationResponse{
		ID:                n.ID,
		SenderID:          n.SenderID,
		Type:              n.Type,
		RedirectPostID:    n.RedirectPostID,
		RedirectUserID:    n.RedirectUserID,
		RedirectCommentID: n.RedirectCommentID,
		Content:           n.Content,
		IsRead:            n.IsRead,
		CreatedAt:         n.CreatedAt,
	}
}

func ToNotificationResponseList(items []models.Notification) []NotificationResponse {
	res := make([]NotificationResponse, len(items))
	for i, n := range items {
		res[i] = ToNotificationResponse(&n)
	}
	return res
}
