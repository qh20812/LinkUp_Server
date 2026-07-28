package models

import (
	"strings"
	"time"
)

type CommentStatus string

const (
	CommentStatusActive  CommentStatus = "active"
	CommentStatusHidden  CommentStatus = "hidden"
	CommentStatusDeleted CommentStatus = "deleted"
)

type Comment struct {
	ID           string         `json:"id" gorm:"primaryKey"`
	UserID       string         `json:"user_id"`
	PostID       string         `json:"post_id"`
	ParentID     *string        `json:"parent_id,omitempty"`
	Content      string         `json:"content"`
	Status       CommentStatus  `json:"status"`
	ReviewReason *string        `json:"review_reason,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    *time.Time     `json:"updated_at,omitempty"`
}

func NewComment(userID, postID string, parentID *string, content string) Comment {
	return Comment{
		UserID:   userID,
		PostID:   postID,
		ParentID: parentID,
		Content:  content,
		Status:   CommentStatusActive,
	}
}

func (s CommentStatus) String() string {
	return string(s)
}

func ParseCommentStatus(value string) CommentStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(CommentStatusActive):
		return CommentStatusActive
	case string(CommentStatusHidden):
		return CommentStatusHidden
	case string(CommentStatusDeleted):
		return CommentStatusDeleted
	default:
		return CommentStatusActive
	}
}
