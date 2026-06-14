package models

import "time"

type Comment struct {
	ID        int64      `json:"id" db:"id"`
	UserID    int64      `json:"user_id" db:"user_id"`
	PostID    int64      `json:"post_id" db:"post_id"`
	ParentID  *int64     `json:"parent_id,omitempty" db:"parent_id"`
	Content   string     `json:"content" db:"content"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

func NewComment(userID, postID int64, parentID *int64, content string) Comment {
	return Comment{UserID: userID, PostID: postID, ParentID: parentID, Content: content}
}
