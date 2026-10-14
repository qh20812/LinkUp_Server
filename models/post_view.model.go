package models

import "time"

// PostViewSource phân biệt nguồn lượt xem: lướt feed (impression)
// hay mở trang chi tiết bài viết.
type PostViewSource string

const (
	PostViewSourceFeed   PostViewSource = "feed"
	PostViewSourceDetail PostViewSource = "detail"
)

// PostView log từng lượt xem được tính (đã dedup 1 user/post/ngày).
// views_count trên posts là counter cache tổng hợp từ bảng này.
type PostView struct {
	ID       string         `gorm:"type:char(36);primaryKey" json:"id"`
	PostID   string         `gorm:"type:char(36);not null;index:idx_post_views_post_created" json:"post_id"`
	ViewerID string         `gorm:"type:char(36);not null;index:idx_post_views_viewer_post_created" json:"viewer_id"`
	Source   PostViewSource `gorm:"type:varchar(16);not null" json:"source"`

	CreatedAt time.Time `gorm:"index:idx_post_views_post_created;index:idx_post_views_viewer_post_created" json:"created_at"`
}

func (PostView) TableName() string {
	return "post_views"
}
