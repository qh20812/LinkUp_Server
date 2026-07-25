package dto

import (
	"linkup/models"
	"time"
)

type NotificationResponse struct {
	ID                string                  `json:"id"`
	SenderID          *string                 `json:"sender_id,omitempty"`
	SenderName        string                  `json:"sender_name,omitempty"`
	SenderAvatar      string                  `json:"sender_avatar,omitempty"`
	Type              models.NotificationType `json:"type"`
	RedirectPostID    *string                 `json:"redirect_post_id,omitempty"`
	RedirectUserID    *string                 `json:"redirect_user_id,omitempty"`
	RedirectCommentID *string                 `json:"redirect_comment_id,omitempty"`
	Content           string                  `json:"content"`
	IsRead            bool                    `json:"is_read"`
	CreatedAt         time.Time               `json:"created_at"`
}

type SenderProfile struct {
	DisplayName string
	AvatarURI   string
}

func ToNotificationResponse(n *models.Notification, sender *SenderProfile) NotificationResponse {
	r := NotificationResponse{
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
	if sender != nil {
		r.SenderName = sender.DisplayName
		r.SenderAvatar = sender.AvatarURI
	}
	return r
}

func ToNotificationResponseList(items []models.Notification, senderMap map[string]SenderProfile) []NotificationResponse {
	res := make([]NotificationResponse, len(items))
	for i, n := range items {
		var sender *SenderProfile
		if n.SenderID != nil {
			if s, ok := senderMap[*n.SenderID]; ok {
				sender = &s
			}
		}
		res[i] = ToNotificationResponse(&n, sender)
	}
	return res
}
