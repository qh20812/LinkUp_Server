package models

import "time"

type PostReaction struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	PostID    string    `json:"post_id"`
	EmojiID   string    `json:"emoji_id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewPostReaction(userID, postID, emojiID string) PostReaction {
	return PostReaction{UserID: userID, PostID: postID, EmojiID: emojiID}
}
