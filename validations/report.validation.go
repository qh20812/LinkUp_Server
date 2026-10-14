package validations

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrTargetTypeRequired = errors.New("target_type là bắt buộc")
	ErrTargetTypeInvalid  = errors.New("target_type phải là 'user', 'post' hoặc 'comment'")
	ErrTargetIDRequired   = errors.New("target_id là bắt buộc")
	ErrReportTypeRequired = errors.New("report_type là bắt buộc")
	ErrReasonRequired     = errors.New("reason_detail là bắt buộc")
	ErrReasonTooLong      = errors.New("reason_detail không được vượt quá 1000 ký tự")
)

const maxReasonDetailRunes = 1000

type ReportValidation struct{}

func NewReportValidation() *ReportValidation {
	return &ReportValidation{}
}

// ValidateCreateReport kiểm tra đầu vào tạo report. Khi đã chọn violation rule
// cụ thể (hasRule=true) thì reason_detail là optional — phân loại đã nằm ở rule;
// khi không kèm rule (lý do khác) thì reason_detail bắt buộc.
func (v *ReportValidation) ValidateCreateReport(targetType, targetID, reportType, reasonDetail string, hasRule bool) error {
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

	if !hasRule && strings.TrimSpace(reasonDetail) == "" {
		return ErrReasonRequired
	}

	if utf8.RuneCountInString(reasonDetail) > maxReasonDetailRunes {
		return ErrReasonTooLong
	}

	return nil
}

func (v *ReportValidation) ValidateUpdateReport(reportType, reasonDetail string, hasRule bool) error {
	if strings.TrimSpace(reportType) == "" {
		return ErrReportTypeRequired
	}

	if !hasRule && strings.TrimSpace(reasonDetail) == "" {
		return ErrReasonRequired
	}

	if utf8.RuneCountInString(reasonDetail) > maxReasonDetailRunes {
		return ErrReasonTooLong
	}

	return nil
}
