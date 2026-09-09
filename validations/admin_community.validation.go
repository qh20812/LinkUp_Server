package validations

import errorsapp "linkup/errors"

var (
	ErrAdminCommunityActionInvalid    = errorsapp.New(errorsapp.ErrCodeAdminActionInvalid)
	ErrAdminCommunityReasonRequired   = errorsapp.New(errorsapp.ErrCodeAdminReasonRequired)
	ErrAdminCommunityReasonTooShort   = errorsapp.New(errorsapp.ErrCodeAdminReasonTooShort)
	ErrAdminCommunityReasonTooLong    = errorsapp.New(errorsapp.ErrCodeAdminReasonTooLong)
	ErrAdminCommunityTransferSelf     = errorsapp.New(errorsapp.ErrCodeAdminTransferSelf)
	ErrAdminCommunityTransferEmpty    = errorsapp.New(errorsapp.ErrCodeAdminTransferEmpty)
	ErrAdminCommunityActionNotAllowed = errorsapp.New(errorsapp.ErrCodeAdminActionNotAllowedComm)
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
