package models

import "time"

// FeedImpression log mỗi bài được trả ở trang 1 feed, kèm vị trí và điểm.
// Khác post_views: bảng này KHÔNG dedup theo ngày — mỗi lần tải trang là một
// dòng, dùng để đo vòng xoay top slot (AvgTopSlotVariety) và hiệu quả rank.
type FeedImpression struct {
	ID        string    `gorm:"type:char(36);primaryKey" json:"id"`
	RequestID string    `gorm:"type:char(36);not null" json:"request_id"`
	PostID    string    `gorm:"type:char(36);not null" json:"post_id"`
	ViewerID  string    `gorm:"type:char(36);not null" json:"viewer_id"`
	Position  int       `gorm:"not null" json:"position"`
	Score     float64   `gorm:"not null" json:"score"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (FeedImpression) TableName() string {
	return "feed_impressions"
}
