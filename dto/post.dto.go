package dto

// DTO tạo bài viết
type CreatePostInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"status"` // Cho phép điền: public, private, friend
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
