package dto

import "time"

type FriendRequestResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type FriendRequestItem struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURI   string    `json:"avatar_uri"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	Direction   string    `json:"direction"`
}

type FriendRequestListResponse struct {
	Sent     []FriendRequestItem `json:"sent"`
	Received []FriendRequestItem `json:"received"`
}

type FriendActionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
