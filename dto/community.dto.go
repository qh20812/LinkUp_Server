package dto

import "time"

type CreateCommunityInput struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
	AvatarURI   string `json:"avatar_uri" binding:"omitempty,url"`
}

type JoinRequestItem struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURI   string    `json:"avatar_uri"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type JoinRequestListResponse struct {
	Requests []JoinRequestItem `json:"requests"`
}
