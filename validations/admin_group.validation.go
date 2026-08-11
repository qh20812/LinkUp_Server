package validations

import errorsapp "linkup/errors"

var (
	ErrAdminGroupActionInvalid    = errorsapp.New(errorsapp.ErrCodeAdminActionInvalid)
	ErrAdminGroupReasonRequired   = errorsapp.New(errorsapp.ErrCodeAdminReasonRequired)
	ErrAdminGroupReasonTooShort   = errorsapp.New(errorsapp.ErrCodeAdminReasonTooShort)
	ErrAdminGroupReasonTooLong    = errorsapp.New(errorsapp.ErrCodeAdminReasonTooLong)
	ErrAdminGroupTransferSelf     = errorsapp.New(errorsapp.ErrCodeAdminTransferSelf)
	ErrAdminGroupTransferEmpty    = errorsapp.New(errorsapp.ErrCodeAdminTransferEmpty)
	ErrAdminGroupActionNotAllowed = errorsapp.New(errorsapp.ErrCodeAdminActionNotAllowedGroup)
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
