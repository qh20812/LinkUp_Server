package validations

import (
	"errors"
	"linkup/models"
	"strings"
	"unicode/utf8"
)

var (
	ErrRuleCategoryRequired = errors.New("danh mục nội quy không hợp lệ")
	ErrRuleTitleRequired    = errors.New("tiêu đề nội quy không được để trống")
	ErrRuleTitleMinLength   = errors.New("tiêu đề nội quy phải có ít nhất 5 ký tự")
	ErrRuleTitleMaxLength   = errors.New("tiêu đề nội quy không được vượt quá 255 ký tự")
	ErrRuleContentMaxLength = errors.New("nội dung nội quy không được vượt quá 2000 ký tự")
	ErrRulePositionNegative = errors.New("vị trí không được âm")
)

type CommunityRuleValidation struct{}

func NewCommunityRuleValidation() *CommunityRuleValidation {
	return &CommunityRuleValidation{}
}

func (v *CommunityRuleValidation) ValidateCreateRule(category models.RuleCategory, title, content string, position int) error {
	if err := v.ValidateCategory(category); err != nil {
		return err
	}
	if err := v.ValidateTitle(title); err != nil {
		return err
	}
	if err := v.ValidateContent(content); err != nil {
		return err
	}
	if position < 0 {
		return ErrRulePositionNegative
	}
	return nil
}

func (v *CommunityRuleValidation) ValidateCategory(category models.RuleCategory) error {
	switch category {
	case models.RuleConduct, models.RuleProhibited, models.RuleGuidelines:
		return nil
	}
	return ErrRuleCategoryRequired
}

func (v *CommunityRuleValidation) ValidateTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrRuleTitleRequired
	}
	if utf8.RuneCountInString(title) < 5 {
		return ErrRuleTitleMinLength
	}
	if utf8.RuneCountInString(title) > 255 {
		return ErrRuleTitleMaxLength
	}
	return nil
}

func (v *CommunityRuleValidation) ValidateContent(content string) error {
	if content == "" {
		return nil
	}
	if utf8.RuneCountInString(content) > 2000 {
		return ErrRuleContentMaxLength
	}
	return nil
}
