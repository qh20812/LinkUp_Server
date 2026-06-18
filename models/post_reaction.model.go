package models

import "time"

type PostReaction struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	PostID    string    `json:"post_id" db:"post_id"`
	EmojiID   string    `json:"emoji_id" db:"emoji_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func NewPostReaction(userID, postID, emojiID string) PostReaction {
	return PostReaction{UserID: userID, PostID: postID, EmojiID: emojiID}
}
