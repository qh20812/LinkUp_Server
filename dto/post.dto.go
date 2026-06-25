package dto

// DTO tạo bài viết
type CreatePostInput struct {
	Title   string `json:"title" binding:"required,min=5,max=150"`
	Content string `json:"content" binding:"required,max=5000"`
}

// DTO thả cảm xúc
type ReactPostInput struct {
	EmojiID string `json:"emoji_id" binding:"required"`
}

// DTO tạo bình luận/phản hồi
type CreateCommentInput struct {
	Content  string  `json:"content" binding:"required,max=1000"`
	ParentID *string `json:"parent_id,omitempty"`
}
