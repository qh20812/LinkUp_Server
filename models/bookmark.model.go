package models

import "time"

type Bookmark struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	PostID    int64     `json:"post_id" db:"post_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func NewBookmark(userID, postID int64) Bookmark {
	return Bookmark{UserID: userID, PostID: postID}
}
