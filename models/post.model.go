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
	ID         string     `json:"id" db:"id"`
	UserID     string     `json:"user_id" db:"user_id"`
	Title      string     `json:"title" db:"title"`
	Content    string     `json:"content" db:"content"`
	ViewsCount int        `json:"views_count" db:"views_count"`
	Status     PostStatus `json:"status" db:"status"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty" db:"updated_at"`

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