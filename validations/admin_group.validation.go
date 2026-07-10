package validations

import "errors"

var (
	ErrAdminGroupActionInvalid      = errors.New("hành động moderation không hợp lệ")
	ErrAdminGroupReasonRequired     = errors.New("lý do là bắt buộc")
	ErrAdminGroupReasonTooShort     = errors.New("lý do phải có ít nhất 10 ký tự")
	ErrAdminGroupReasonTooLong      = errors.New("lý do không được vượt quá 1000 ký tự")
	ErrAdminGroupTransferSelf       = errors.New("không thể chuyển quyền cho chính mình")
	ErrAdminGroupTransferEmpty      = errors.New("người nhận không được để trống")
	ErrAdminGroupActionNotAllowed   = errors.New("action không được hỗ trợ cho group chat")
)

var AllowedGroupModerationActions = []string{"hide", "unhide", "archive", "warn"}

type AdminGroupValidation struct{}

func NewAdminGroupValidation() *AdminGroupValidation {
	return &AdminGroupValidation{}
}

func (v *AdminGroupValidation) ValidateModerateAction(action string) error {
	for _, a := range AllowedGroupModerationActions {
		if a == action {
			return nil
		}
	}
	return ErrAdminGroupActionInvalid
}

func (v *AdminGroupValidation) ValidateModerateReason(reason string) error {
	if reason == "" {
		return ErrAdminGroupReasonRequired
	}
	if len(reason) < 10 {
		return ErrAdminGroupReasonTooShort
	}
	if len(reason) > 1000 {
		return ErrAdminGroupReasonTooLong
	}
	return nil
}

func (v *AdminGroupValidation) ValidateTransferOwnership(targetUserID, requesterID string) error {
	if targetUserID == "" {
		return ErrAdminGroupTransferEmpty
	}
	if targetUserID == requesterID {
		return ErrAdminGroupTransferSelf
	}
	return nil
}
