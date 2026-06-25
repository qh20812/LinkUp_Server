package models

import "time"

type Comment struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	PostID    string     `json:"post_id"`
	ParentID  *string    `json:"parent_id,omitempty"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
}

func NewComment(userID, postID string, parentID *string, content string) Comment {
	return Comment{UserID: userID, PostID: postID, ParentID: parentID, Content: content}
}
