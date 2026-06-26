package models

import (
	"strings"
	"time"
)

type PostStatus string

const (
	PostStatusActive  PostStatus = "active"
	PostStatusHidden  PostStatus = "hidden"
	PostStatusDeleted PostStatus = "deleted"
)

type Post struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	ViewsCount int        `json:"views_count"`
	Status     PostStatus `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`

	LikesCount    int `json:"likes_count" gorm:"->"`
	CommentsCount int `json:"comments_count" gorm:"->"`
	SharesCount   int `json:"shares_count" gorm:"->"`
}

func NewPost(userID, title, content string) Post {
	return Post{
		UserID:  userID,
		Title:   title,
		Content: content,
		Status:  PostStatusActive,
	}
}

func (s PostStatus) String() string {
	return string(s)
}

func ParsePostStatus(value string) PostStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(PostStatusActive):
		return PostStatusActive
	case string(PostStatusHidden):
		return PostStatusHidden
	case string(PostStatusDeleted):
		return PostStatusDeleted
	default:
		return PostStatusActive
	}
}