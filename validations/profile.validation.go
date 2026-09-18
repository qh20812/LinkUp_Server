package validations

import (
	"strings"

	errorsapp "linkup/errors"
)

// WorkCodes are the accepted profile work enum values.
var WorkCodes = map[string]struct{}{
	"it": {}, "business": {}, "trading": {}, "marketing": {}, "finance": {},
	"education": {}, "healthcare": {}, "engineering": {}, "manufacturing": {}, "law": {},
	"arts": {}, "service": {}, "logistics": {}, "agriculture": {}, "realestate": {},
	"student": {}, "housewife": {}, "retired": {}, "other": {},
}

// EducationCodes are the accepted profile education enum values.
var EducationCodes = map[string]struct{}{
	"below_highschool": {}, "highschool": {}, "intermediate": {}, "college": {},
	"university": {}, "master": {}, "doctorate": {}, "other": {},
}

// ProfileValidation validates the new structured profile fields.
type ProfileValidation struct{}

func NewProfileValidation() *ProfileValidation {
	return &ProfileValidation{}
}

// ValidateWorkCode returns an error unless code is empty or a known work value.
func (v *ProfileValidation) ValidateWorkCode(code string) error {
	if code == "" {
		return nil
	}
	if _, ok := WorkCodes[code]; !ok {
		return errorsapp.New(errorsapp.ErrCodeWorkInvalid)
	}
	return nil
}

// ValidateEducationCode returns an error unless code is empty or a known
// education value.
func (v *ProfileValidation) ValidateEducationCode(code string) error {
	if code == "" {
		return nil
	}
	if _, ok := EducationCodes[code]; !ok {
		return errorsapp.New(errorsapp.ErrCodeEducationInvalid)
	}
	return nil
}

// ValidateWorkOther enforces the "other" free-text rule: required when the
// work code is "other", and capped at 255 characters.
func (v *ProfileValidation) ValidateWorkOther(workCode, workOther string) error {
	if workCode == "other" && strings.TrimSpace(workOther) == "" {
		return errorsapp.New(errorsapp.ErrCodeWorkOtherRequired)
	}
	if len(workOther) > 255 {
		return errorsapp.New(errorsapp.ErrCodeWorkOtherTooLong)
	}
	return nil
}