package models

import "time"

// FeedHeroState lưu vòng xoay slot #1 của từng user.
// Mỗi lần tải trang 1 của feed, hero_round tăng 1; rank dùng
// (hero_round-1) % số slot để chọn bài đứng đầu. Nhờ vậy một bài không
// thể đứng #1 vĩnh viễn dù điểm cao hơn phần còn lại.
type FeedHeroState struct {
	UserID    string    `gorm:"type:char(36);primaryKey" json:"user_id"`
	HeroRound int       `gorm:"not null;default:0" json:"hero_round"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

func (FeedHeroState) TableName() string {
	return "feed_hero_state"
}
