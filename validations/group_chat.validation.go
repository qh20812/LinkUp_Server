package validations

import (
	"errors"
	"unicode/utf8"
)

var (
	ErrGroupNameInvalid = errors.New("tên nhóm chat phải từ 3 đến 50 ký tự")
)

type GroupChatValidation struct{}

func NewGroupChatValidation() *GroupChatValidation {
	return &GroupChatValidation{}
}

func (v *GroupChatValidation) ValidateCreateGroup(name string) error {
	length := utf8.RuneCountInString(name)
	if length < 3 || length > 50 {
		return ErrGroupNameInvalid
	}
	return nil
}
