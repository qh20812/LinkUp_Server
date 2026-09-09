package validations

import (
	errorsapp "linkup/errors"
	"strings"
)

type FriendValidation struct{}

func NewFriendValidation() *FriendValidation {
	return &FriendValidation{}
}

func (v *FriendValidation) ValidateToggleFriendRequest(userID, targetUserID string) error {
	if strings.TrimSpace(targetUserID) == "" {
		return errorsapp.New(errorsapp.ErrCodeFriendTargetRequired)
	}
	if userID == targetUserID {
		return errorsapp.New(errorsapp.ErrCodeFriendSelfRequest)
	}
	return nil
}
