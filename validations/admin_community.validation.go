package validations

import "errors"

var (
	ErrAdminCommunityActionInvalid      = errors.New("hành động moderation không hợp lệ")
	ErrAdminCommunityReasonRequired     = errors.New("lý do là bắt buộc")
	ErrAdminCommunityReasonTooShort     = errors.New("lý do phải có ít nhất 10 ký tự")
	ErrAdminCommunityReasonTooLong      = errors.New("lý do không được vượt quá 1000 ký tự")
	ErrAdminCommunityTransferSelf       = errors.New("không thể chuyển quyền cho chính mình")
	ErrAdminCommunityTransferEmpty      = errors.New("người nhận không được để trống")
	ErrAdminCommunityActionNotAllowed   = errors.New("action không được hỗ trợ cho cộng đồng")
)

var AllowedCommunityModerationActions = []string{"hide", "unhide", "archive", "warn"}

type AdminCommunityValidation struct{}

func NewAdminCommunityValidation() *AdminCommunityValidation {
	return &AdminCommunityValidation{}
}

func (v *AdminCommunityValidation) ValidateModerateAction(action string) error {
	for _, a := range AllowedCommunityModerationActions {
		if a == action {
			return nil
		}
	}
	return ErrAdminCommunityActionInvalid
}

func (v *AdminCommunityValidation) ValidateModerateReason(reason string) error {
	if reason == "" {
		return ErrAdminCommunityReasonRequired
	}
	if len(reason) < 10 {
		return ErrAdminCommunityReasonTooShort
	}
	if len(reason) > 1000 {
		return ErrAdminCommunityReasonTooLong
	}
	return nil
}

func (v *AdminCommunityValidation) ValidateTransferOwnership(targetUserID, requesterID string) error {
	if targetUserID == "" {
		return ErrAdminCommunityTransferEmpty
	}
	if targetUserID == requesterID {
		return ErrAdminCommunityTransferSelf
	}
	return nil
}
