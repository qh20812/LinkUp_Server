package dto

import "time"

type AdminUserFilterInput struct {
	Keyword  string `json:"keyword" form:"keyword"`
	Status   string `json:"status" form:"status"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
}

type AdminUserListItem struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Status      string     `json:"status"`
	DisplayName string     `json:"display_name"`
	AvatarURI   string     `json:"avatar_uri"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type AdminUserListResponse struct {
	Users    []AdminUserListItem `json:"users"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Message  string              `json:"message,omitempty"`
}

type AdminUserUpdateStatusInput struct {
	Status string `json:"status" binding:"required"`
}
