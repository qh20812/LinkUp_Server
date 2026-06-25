package models

import "time"

type PostShare struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// NewPostShare khởi tạo một đối tượng Share mới
func NewPostShare(userID, postID string) PostShare {
	return PostShare{
		UserID: userID,
		PostID: postID,
	}
}
