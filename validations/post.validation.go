package validations

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrPostTitleRequired     = errors.New("title is required")
	ErrPostTitleMinLength    = errors.New("title must be at least 5 characters")
	ErrPostTitleMaxLength    = errors.New("title must be at most 150 characters")
	ErrPostContentRequired   = errors.New("content is required")
	ErrPostContentMaxLength  = errors.New("content must be at most 5000 characters")
	ErrPostIDRequired        = errors.New("post id is required")
	ErrEmojiRequired         = errors.New("emoji_id is required")
	ErrCommentContentMaxLen  = errors.New("comment content must be at most 1000 characters")
	ErrInvalidPageSize       = errors.New("page_size must be between 1 and 100")
)

type PostValidation struct{}

func NewPostValidation() *PostValidation {
	return &PostValidation{}
}

func (v *PostValidation) ValidateCreatePost(title, content string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrPostTitleRequired
	}
	if utf8.RuneCountInString(title) < 5 {
		return ErrPostTitleMinLength
	}
	if utf8.RuneCountInString(title) > 150 {
		return ErrPostTitleMaxLength
	}

	if strings.TrimSpace(content) == "" {
		return ErrPostContentRequired
	}
	if utf8.RuneCountInString(content) > 5000 {
		return ErrPostContentMaxLength
	}

	return nil
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
