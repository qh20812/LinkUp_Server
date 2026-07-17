package validations

import (
	"errors"
	"fmt"
	"linkup/models"
	"strings"
	"unicode/utf8"
)

var (
	ErrPostTitleRequired    = errors.New("tiêu đề bài viết là bắt buộc")
	ErrPostTitleMinLength   = errors.New("tiêu đề bài viết phải có ít nhất 5 ký tự")
	ErrPostTitleMaxLength   = errors.New("tiêu đề bài viết không được vượt quá 150 ký tự")
	ErrPostContentRequired  = errors.New("nội dung bài viết là bắt buộc")
	ErrPostContentMaxLength = errors.New("nội dung bài viết không được vượt quá 5000 ký tự")
	ErrPostIDRequired       = errors.New("mã bài viết là bắt buộc")
	ErrEmojiRequired        = errors.New("emoji_id là bắt buộc")
	ErrCommentContentMaxLen = errors.New("nội dung bình luận không được vượt quá 1000 ký tự")
	ErrInvalidPageSize      = errors.New("page_size phải từ 1 đến 100")
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
func ValidateCreatePost(title, content, status string, hasFiles bool) error {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	status = strings.ToLower(strings.TrimSpace(status))

	if hasFiles && title == "" && content == "" {
		return nil
	}

	// Kiểm tra tiêu đề (Từ 5 đến 150 ký tự, không bắt buộc nếu có file)
	if title != "" && (len(title) < 5 || len(title) > 150) {
		return errors.New("tiêu đề bài viết phải từ 5 đến 150 ký tự")
	}

	// Kiểm tra nội dung (Tối đa 5000 ký tự, không bắt buộc nếu có file)
	if !hasFiles && content == "" {
		return errors.New("nội dung bài viết không được bỏ trống")
	}
	if content != "" && len(content) > 5000 {
		return errors.New("nội dung bài viết không được vượt quá 5000 ký tự")
	}

	// Kiểm tra trạng thái bài viết hợp lệ
	validStatuses := map[models.PostStatus]bool{
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
