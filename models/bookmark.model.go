package models

import "time"

type Bookmark struct {
	ID        string     `json:"id" db:"id"`
	UserID    string `json:"user_id" db:"user_id"`
	PostID    string `json:"post_id" db:"post_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func NewBookmark(userID, postID string) Bookmark {
	return Bookmark{UserID: userID, PostID: postID}
}
