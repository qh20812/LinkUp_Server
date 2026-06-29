package validations

import (
	"errors"
	"strings"
)

var (
	ErrKeywordTooShort = errors.New("từ khóa tìm kiếm phải có ít nhất 2 ký tự")
	ErrSearchTypeInvalid = errors.New("loại tìm kiếm phải là 'all', 'users', 'posts' hoặc 'hashtags'")
)

type SearchValidation struct{}

func NewSearchValidation() *SearchValidation {
	return &SearchValidation{}
}

func (v *SearchValidation) ValidateSearch(keyword, searchType string) error {
	if len(strings.TrimSpace(keyword)) < 2 {
		return ErrKeywordTooShort
	}

	if searchType != "" && searchType != "all" && searchType != "users" && searchType != "posts" && searchType != "hashtags" {
		return ErrSearchTypeInvalid
	}

	return nil
}
