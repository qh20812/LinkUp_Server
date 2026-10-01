package models

import "time"

// UserInterest là interest profile của user theo từng hashtag.
// Score tích lũy từ các tương tác (view/react/comment/share/save) và
// được decay lười (lazy) lúc đọc — không cần cron job.
type UserInterest struct {
	UserID string  `gorm:"type:char(36);primaryKey" json:"user_id"`
	Tag    string  `gorm:"type:varchar(255);primaryKey" json:"tag"`
	Score  float64 `gorm:"not null;default:0" json:"score"`

	UpdatedAt time.Time `json:"updated_at"`
}

func (UserInterest) TableName() string {
	return "user_interests"
}
