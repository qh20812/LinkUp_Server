package dto

import "time"

// CreateStoryResponse phản hồi sau khi đăng story thành công
type CreateStoryResponse struct {
	ID        string    `json:"id"`
	MediaURI  string    `json:"media_uri"`
	MediaType string    `json:"media_type"`
	Caption   string    `json:"caption"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// StoryResponse hiển thị thông tin story trên New Feed
type StoryResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	MediaURI  string    `json:"media_uri"`
	MediaType string    `json:"media_type"`
	Caption   string    `json:"caption"`
	CreatedAt time.Time `json:"created_at"`
}

// InteractStoryRequest nhận tương tác (React/Reply/Share)
type InteractStoryRequest struct {
	Type    string `json:"type" binding:"required,oneof=react reply share"`
	EmojiID string `json:"emoji_id"`
	Content string `json:"content"`
}

// StoryAnalyticsResponse thống kê chi tiết cho chủ sở hữu
type StoryAnalyticsResponse struct {
	StoryID      string `json:"story_id"`
	TotalViews   int64  `json:"total_views"`
	TotalReacts  int64  `json:"total_reacts"`
	TotalReplies int64  `json:"total_replies"`
	TotalShares  int64  `json:"total_shares"`
}
