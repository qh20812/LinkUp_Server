package models

import (
	"strings"
	"time"
)

// ApplicableTo: phạm vi áp dụng của rule. "all" dùng cho mọi loại target.
type ViolationApplicable string

const (
	ViolationApplicableAll     ViolationApplicable = "all"
	ViolationApplicablePost    ViolationApplicable = "post"
	ViolationApplicableComment ViolationApplicable = "comment"
	ViolationApplicableUser    ViolationApplicable = "user"
)

// Severity: mức độ nghiêm trọng để admin ưu tiên duyệt.
type ViolationSeverity string

const (
	ViolationSeverityLow    ViolationSeverity = "low"
	ViolationSeverityMedium ViolationSeverity = "medium"
	ViolationSeverityHigh   ViolationSeverity = "high"
)

type ViolationRule struct {
	ID           string              `json:"id"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	ApplicableTo ViolationApplicable `json:"applicable_to"`
	Severity     ViolationSeverity   `json:"severity"`
	SortOrder    int                 `json:"sort_order"`
	IsActive     bool                `json:"is_active"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    *time.Time          `json:"updated_at,omitempty"`
}

func (ViolationRule) TableName() string {
	return "violation_rules"
}

func NewViolationRule(title, description string) ViolationRule {
	return ViolationRule{
		Title:        title,
		Description:  description,
		ApplicableTo: ViolationApplicableAll,
		Severity:     ViolationSeverityMedium,
		IsActive:     true,
	}
}

// MatchesTarget: rule có áp dụng cho loại target này không.
func (r ViolationRule) MatchesTarget(targetType string) bool {
	if r.ApplicableTo == ViolationApplicableAll {
		return true
	}
	return string(r.ApplicableTo) == strings.TrimSpace(strings.ToLower(targetType))
}

func ParseViolationApplicable(value string) ViolationApplicable {
	switch ViolationApplicable(strings.TrimSpace(strings.ToLower(value))) {
	case ViolationApplicablePost:
		return ViolationApplicablePost
	case ViolationApplicableComment:
		return ViolationApplicableComment
	case ViolationApplicableUser:
		return ViolationApplicableUser
	default:
		return ViolationApplicableAll
	}
}

func ParseViolationSeverity(value string) ViolationSeverity {
	switch ViolationSeverity(strings.TrimSpace(strings.ToLower(value))) {
	case ViolationSeverityLow:
		return ViolationSeverityLow
	case ViolationSeverityHigh:
		return ViolationSeverityHigh
	default:
		return ViolationSeverityMedium
	}
}
