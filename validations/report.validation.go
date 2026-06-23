package validations

import (
	"errors"
	"strings"
)

var (
	ErrTargetTypeRequired = errors.New("target_type is required")
	ErrTargetTypeInvalid  = errors.New("target_type must be 'user', 'post', or 'comment'")
	ErrTargetIDRequired   = errors.New("target_id is required")
	ErrReportTypeRequired = errors.New("report_type is required")
	ErrReasonRequired     = errors.New("reason_detail is required")
)

type ReportValidation struct{}

func NewReportValidation() *ReportValidation {
	return &ReportValidation{}
}

func (v *ReportValidation) ValidateCreateReport(targetType, targetID, reportType, reasonDetail string) error {
	targetType = strings.TrimSpace(targetType)
	if targetType == "" {
		return ErrTargetTypeRequired
	}
	if targetType != "user" && targetType != "post" && targetType != "comment" {
		return ErrTargetTypeInvalid
	}

	if strings.TrimSpace(targetID) == "" {
		return ErrTargetIDRequired
	}

	if strings.TrimSpace(reportType) == "" {
		return ErrReportTypeRequired
	}

	if strings.TrimSpace(reasonDetail) == "" {
		return ErrReasonRequired
	}

	return nil
}
