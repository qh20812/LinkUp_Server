package validations

import (
	"errors"
	"strings"
)

var (
	ErrKeywordTooShort = errors.New("keyword must be at least 2 characters")
	ErrSearchTypeInvalid = errors.New("search type must be 'all', 'users', 'posts', or 'hashtags'")
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
