package models

import "time"

type PostShare struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func NewPostShare(userID, postID, content string) PostShare {
	return PostShare{
		UserID:  userID,
		PostID:  postID,
		Content: content,
	}
}
