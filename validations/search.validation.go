package validations

import (
	errorsapp "linkup/errors"
	"strings"
)

type SearchValidation struct{}

func NewSearchValidation() *SearchValidation {
	return &SearchValidation{}
}

func (v *SearchValidation) ValidateSearch(keyword, searchType string) error {
	if len(strings.TrimSpace(keyword)) < 2 {
		return errorsapp.New(errorsapp.ErrCodeSearchKeywordTooShort)
	}

	if searchType != "" && searchType != "all" && searchType != "users" && searchType != "posts" && searchType != "hashtags" {
		return errorsapp.New(errorsapp.ErrCodeSearchTypeInvalid)
	}

	return nil
}
