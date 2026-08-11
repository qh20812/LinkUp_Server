package validations

import (
	errorsapp "linkup/errors"
	"linkup/models"
	"strings"
	"unicode/utf8"
)

var (
	ErrPostTitleRequired    = errorsapp.New(errorsapp.ErrCodePostTitleRequired)
	ErrPostTitleMinLength   = errorsapp.New(errorsapp.ErrCodePostTitleTooShort)
	ErrPostTitleMaxLength   = errorsapp.New(errorsapp.ErrCodePostTitleTooLong)
	ErrPostContentRequired  = errorsapp.New(errorsapp.ErrCodePostContentRequired)
	ErrPostContentMaxLength = errorsapp.New(errorsapp.ErrCodePostContentTooLong)
	ErrPostIDRequired       = errorsapp.New(errorsapp.ErrCodePostIDRequired)
	ErrEmojiRequired        = errorsapp.New(errorsapp.ErrCodeEmojiRequired)
	ErrCommentContentMaxLen = errorsapp.New(errorsapp.ErrCodeCommentContentTooLong)
	ErrInvalidPageSize      = errorsapp.New(errorsapp.ErrCodeInvalidPageSize)
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
		return errorsapp.New(errorsapp.ErrCodeCommentContentRequired)
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

func ValidateCreatePost(title, content, status string, hasFiles bool) error {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	status = strings.ToLower(strings.TrimSpace(status))

	if hasFiles && title == "" && content == "" {
		return nil
	}

	if title != "" && (len(title) < 5 || len(title) > 150) {
		return errorsapp.New(errorsapp.ErrCodePostTitleTooShort)
	}

	if !hasFiles && content == "" {
		return errorsapp.New(errorsapp.ErrCodePostContentRequired)
	}
	if content != "" && len(content) > 5000 {
		return errorsapp.New(errorsapp.ErrCodePostContentTooLong)
	}

	validStatuses := map[models.PostStatus]bool{
		models.PostStatusPublic:  true,
		models.PostStatusPrivate: true,
		models.PostStatusHidden:  true,
		models.PostStatusFriend:  true,
	}

	if status != "" && !validStatuses[models.PostStatus(status)] {
		return errorsapp.Newf(errorsapp.ErrCodePostInvalidStatus, map[string]any{
			"status": status,
		})
	}

	return nil
}
