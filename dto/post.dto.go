package dto

// DTO tạo bài viết
type CreatePostInput struct {
	Title           string  `json:"title" form:"title"`
	Content         string  `json:"content" form:"content"`
	Status          string  `json:"status" form:"status"`
	CommunityID     *string `json:"community_id,omitempty" form:"community_id"`
	GifURL          string  `json:"gif_url,omitempty" form:"gif_url"`
	CommentsEnabled *bool   `json:"comments_enabled,omitempty" form:"comments_enabled"`
}

// DTO đổi trạng thái bình luận của bài viết (tắt/bật)
type UpdateCommentsEnabledInput struct {
	Enabled bool `json:"enabled" form:"enabled"`
}

// DTO chia sẻ bài viết có kèm text
type SharePostInput struct {
	Content string `json:"content"`
}

// DTO thả cảm xúc
type ReactPostInput struct {
	EmojiID string `json:"emoji_id" binding:"required"`
}

// DTO tạo bình luận/phản hồi
type CreateCommentInput struct {
	Content  string  `json:"content"`
	ParentID *string `json:"parent_id,omitempty"`
}

// DTO ghi nhận lượt xem bài viết (impression từ feed hoặc mở chi tiết)
type TrackPostViewInput struct {
	Source string `json:"source" binding:"omitempty,oneof=feed detail"`
}
