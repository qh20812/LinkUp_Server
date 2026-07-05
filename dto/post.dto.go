package dto

// DTO tạo bài viết
type CreatePostInput struct {
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	Status      string  `json:"status"`
	CommunityID *string `json:"community_id,omitempty"`
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
