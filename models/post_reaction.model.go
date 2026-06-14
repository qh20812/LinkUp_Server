package models

import "time"

type PostReaction struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	PostID    int64     `json:"post_id" db:"post_id"`
	EmojiID   int64     `json:"emoji_id" db:"emoji_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func NewPostReaction(userID, postID, emojiID int64) PostReaction {
	return PostReaction{UserID: userID, PostID: postID, EmojiID: emojiID}
}
