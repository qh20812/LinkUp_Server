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
	ID         int64      `json:"id" db:"id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	Title      string     `json:"title" db:"title"`
	Content    string     `json:"content" db:"content"`
	ViewsCount int        `json:"views_count" db:"views_count"`
	Status     PostStatus `json:"status" db:"status"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

func NewPost(userID int64, title, content string) Post {
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
