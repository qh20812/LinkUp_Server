package validations

import (
	"errors"
	"strings"
)

var (
	ErrFriendTargetRequired = errors.New("target user is required")
	ErrSelfFriendRequest    = errors.New("cannot send friend request to yourself")
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
