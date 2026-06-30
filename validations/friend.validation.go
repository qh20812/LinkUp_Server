package validations

import (
	"errors"
	"strings"
)

var (
	ErrFriendTargetRequired = errors.New("người dùng mục tiêu là bắt buộc")
	ErrSelfFriendRequest    = errors.New("không thể gửi lời mời kết bạn cho chính mình")
)

type FriendValidation struct{}

func NewFriendValidation() *FriendValidation {
	return &FriendValidation{}
}

func (v *FriendValidation) ValidateToggleFriendRequest(userID, targetUserID string) error {
	if strings.TrimSpace(targetUserID) == "" {
		return ErrFriendTargetRequired
	}
	if userID == targetUserID {
		return ErrSelfFriendRequest
	}
	return nil
}
