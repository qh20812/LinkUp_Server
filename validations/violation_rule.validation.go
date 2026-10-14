package validations

import (
	errorsapp "linkup/errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrViolationRuleTitleRequired = errorsapp.New(errorsapp.ErrCodeViolationRuleTitleRequired)
	ErrViolationRuleTitleTooShort = errorsapp.New(errorsapp.ErrCodeViolationRuleTitleTooShort)
	ErrViolationRuleTitleTooLong  = errorsapp.New(errorsapp.ErrCodeViolationRuleTitleTooLong)
	ErrViolationRuleDescTooLong   = errorsapp.New(errorsapp.ErrCodeViolationRuleDescTooLong)
	ErrViolationRuleTitleDup      = errorsapp.New(errorsapp.ErrCodeViolationRuleTitleDup)
)

type ViolationRuleValidation struct{}

func NewViolationRuleValidation() *ViolationRuleValidation {
	return &ViolationRuleValidation{}
}

func (v *ViolationRuleValidation) ValidateTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrViolationRuleTitleRequired
	}
	if utf8.RuneCountInString(title) < 5 {
		return ErrViolationRuleTitleTooShort
	}
	if utf8.RuneCountInString(title) > 255 {
		return ErrViolationRuleTitleTooLong
	}
	return nil
}

func (v *ViolationRuleValidation) ValidateDescription(description string) error {
	if description == "" {
		return nil
	}
	if utf8.RuneCountInString(description) > 2000 {
		return ErrViolationRuleDescTooLong
	}
	return nil
}

// ValidateApplicable chuẩn hóa applicable_to, mặc định "all".
func (v *ViolationRuleValidation) ValidateApplicable(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "post", "comment", "user":
		return value
	default:
		return "all"
	}
}

// ValidateSeverity chuẩn hóa severity, mặc định "medium".
func (v *ViolationRuleValidation) ValidateSeverity(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "low", "high":
		return value
	default:
		return "medium"
	}
}
