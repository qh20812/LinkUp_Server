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

type FriendItem struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	AvatarURI   string `json:"avatar_uri"`
	Status      string `json:"status"`
}

type FriendListResponse struct {
	Data     []FriendItem `json:"data"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Total    int64        `json:"total"`
	HasMore  bool         `json:"has_more"`
}

type FriendSuggestionItem struct {
	UserID      string   `json:"user_id"`
	DisplayName string   `json:"display_name"`
	AvatarURI   string   `json:"avatar_uri"`
	MutualCount int      `json:"mutual_count"`
	MutualNames []string `json:"mutual_names,omitempty"`
}

type FriendSuggestionsResponse struct {
	Data     []FriendSuggestionItem `json:"data"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Total    int64                  `json:"total"`
	HasMore  bool                   `json:"has_more"`
}

type FriendStatusResponse struct {
	Status    string  `json:"status"`
	RequestID *string `json:"request_id,omitempty"`
}
