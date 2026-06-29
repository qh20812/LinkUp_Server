package validations

import (
	"errors"
	"fmt"
	"linkup/models"
	"strings"
	"unicode/utf8"
)

var (
	ErrPostTitleRequired    = errors.New("title is required")
	ErrPostTitleMinLength   = errors.New("title must be at least 5 characters")
	ErrPostTitleMaxLength   = errors.New("title must be at most 150 characters")
	ErrPostContentRequired  = errors.New("content is required")
	ErrPostContentMaxLength = errors.New("content must be at most 5000 characters")
	ErrPostIDRequired       = errors.New("post id is required")
	ErrEmojiRequired        = errors.New("emoji_id is required")
	ErrCommentContentMaxLen = errors.New("comment content must be at most 1000 characters")
	ErrInvalidPageSize      = errors.New("page_size must be between 1 and 100")
)

type PostValidation struct{}

func NewPostValidation() *PostValidation {
	return &PostValidation{}
}

func (v *PostValidation) ValidateReactPost(emojiID string) error {
	if strings.TrimSpace(emojiID) == "" {
		return ErrEmojiRequired
	}
	return nil
}

func (v *PostValidation) ValidateCreateComment(content string) error {
	if strings.TrimSpace(content) == "" {
		return ErrPostContentRequired
	}
	if utf8.RuneCountInString(content) > 1000 {
		return ErrCommentContentMaxLen
	}
	return nil
}

func (v *PostValidation) ValidatePostID(postID string) error {
	if strings.TrimSpace(postID) == "" {
		return ErrPostIDRequired
	}
	return nil
}

func (v *PostValidation) NormalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return page, pageSize
}

// ValidateCreatePost kiểm tra dữ liệu đầu vào khi tạo bài viết
func ValidateCreatePost(title, content, status string) error {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	status = strings.ToLower(strings.TrimSpace(status))

	// Kiểm tra tiêu đề (Từ 5 đến 150 ký tự)
	if len(title) < 5 || len(title) > 150 {
		return errors.New("tiêu đề bài viết phải từ 5 đến 150 ký tự")
	}

	// Kiểm tra nội dung (Tối đa 5000 ký tự)
	if content == "" {
		return errors.New("nội dung bài viết không được bỏ trống")
	}
	if len(content) > 5000 {
		return errors.New("nội dung bài viết không được vượt quá 5000 ký tự")
	}

	// Kiểm tra trạng thái bài viết hợp lệ
	validStatuses := map[models.PostStatus]bool{
		models.PostStatusActive:  true,
		models.PostStatusPublic:  true,
		models.PostStatusPrivate: true,
		models.PostStatusHidden:  true,
		models.PostStatusFriend:  true,
	}

	if status != "" && !validStatuses[models.PostStatus(status)] {
		return fmt.Errorf("trạng thái '%s' không hợp lệ. Chỉ chấp nhận: public, private, friend, hidden", status)
	}

	return nil
}
