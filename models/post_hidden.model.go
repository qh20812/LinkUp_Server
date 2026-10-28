package models

import "time"

// PostHidden là tín hiệu âm chủ động: user bấm "Không quan tâm" một bài.
// Bài bị ẩn không xuất hiện lại trong feed của chính user đó.
type PostHidden struct {
	UserID    string    `gorm:"type:char(36);primaryKey" json:"user_id"`
	PostID    string    `gorm:"type:char(36);primaryKey" json:"post_id"`
	Reason    string    `gorm:"type:varchar(32);not null;default:not_interested" json:"reason"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (PostHidden) TableName() string {
	return "post_hidden"
}

const PostHiddenReasonNotInterested = "not_interested"
