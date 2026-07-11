package validations

import (
	"errors"
	"unicode/utf8"
)

var (
	ErrGroupNameInvalid       = errors.New("tên nhóm chat phải từ 3 đến 50 ký tự")
	ErrMuteReasonInvalid      = errors.New("lý do tắt tiếng không hợp lệ")
	ErrMuteDurationInvalid    = errors.New("thời lượng tắt tiếng không hợp lệ")
	ErrCreatorCannotLeaveGroup = errors.New("người tạo nhóm không thể rời đi, vui lòng chuyển quyền sở hữu trước")
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
		return ErrGroupNameInvalid
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
		return ErrMuteReasonInvalid
	}
	okDur := false
	for _, d := range AllowedMuteDurations {
		if d == duration {
			okDur = true
			break
		}
	}
	if !okDur {
		return ErrMuteDurationInvalid
	}
	return nil
}
