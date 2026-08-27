package models

import "time"

type CommentReaction struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id"`
	CommentID string    `json:"comment_id"`
	EmojiID   string    `json:"emoji_id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewCommentReaction(userID, commentID, emojiID string) CommentReaction {
	return CommentReaction{UserID: userID, CommentID: commentID, EmojiID: emojiID}
}
