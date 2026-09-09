package validations

import (
	errorsapp "linkup/errors"
	"unicode/utf8"
)

var AllowedMuteReasons = []string{"spam", "abuse", "harassment", "violation", "other"}
var AllowedMuteDurations = []int{1, 30, 60, 1440, 0}

type GroupChatValidation struct{}

func NewGroupChatValidation() *GroupChatValidation {
	return &GroupChatValidation{}
}

func (v *GroupChatValidation) ValidateCreateGroup(name string) error {
	length := utf8.RuneCountInString(name)
	if length < 3 || length > 50 {
		return errorsapp.New(errorsapp.ErrCodeGCInvalidName)
	}
	return nil
}

func (v *GroupChatValidation) ValidateMuteInput(reason string, duration int) error {
	okReason := false
	for _, r := range AllowedMuteReasons {
		if r == reason {
			okReason = true
			break
		}
	}
	if !okReason {
		return errorsapp.New(errorsapp.ErrCodeGCInvalidMuteReason)
	}
	okDur := false
	for _, d := range AllowedMuteDurations {
		if d == duration {
			okDur = true
			break
		}
	}
	if !okDur {
		return errorsapp.New(errorsapp.ErrCodeGCInvalidMuteDuration)
	}
	return nil
}
