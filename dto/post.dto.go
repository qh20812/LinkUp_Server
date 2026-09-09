package dto

// DTO tạo bài viết
type CreatePostInput struct {
	Title       string  `json:"title" form:"title"`
	Content     string  `json:"content" form:"content"`
	Status      string  `json:"status" form:"status"`
	CommunityID *string `json:"community_id,omitempty" form:"community_id"`
	GifURL      string  `json:"gif_url,omitempty" form:"gif_url"`
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
