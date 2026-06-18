package models

import "time"

type Comment struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	PostID    string     `json:"post_id" db:"post_id"`
	ParentID  *string    `json:"parent_id,omitempty" db:"parent_id"`
	Content   string     `json:"content" db:"content"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

func NewComment(userID, postID string, parentID *string, content string) Comment {
	return Comment{UserID: userID, PostID: postID, ParentID: parentID, Content: content}
}
