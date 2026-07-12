package validations

import (
	"errors"
	"strings"
)

var (
	ErrTargetTypeRequired = errors.New("target_type là bắt buộc")
	ErrTargetTypeInvalid  = errors.New("target_type phải là 'user', 'post' hoặc 'comment'")
	ErrTargetIDRequired   = errors.New("target_id là bắt buộc")
	ErrReportTypeRequired = errors.New("report_type là bắt buộc")
	ErrReasonRequired     = errors.New("reason_detail là bắt buộc")
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

func (v *ReportValidation) ValidateUpdateReport(reportType, reasonDetail string) error {
	if strings.TrimSpace(reportType) == "" {
		return ErrReportTypeRequired
	}

	if strings.TrimSpace(reasonDetail) == "" {
		return ErrReasonRequired
	}

	return nil
}
