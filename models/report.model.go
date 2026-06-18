package models

import (
	"strings"
	"time"
)

type ReportStatus string

const (
	ReportStatusPending  ReportStatus = "pending"
	ReportStatusReviewed ReportStatus = "reviewed"
	ReportStatusResolved ReportStatus = "resolved"
)

type Report struct {
	ID              string        `json:"id" db:"id"`
	ReporterID      string `json:"reporter_id" db:"reporter_id"`
	ReportType      string       `json:"report_type" db:"report_type"`
	TargetUserID    *string `json:"target_user_id,omitempty" db:"target_user_id"`
	TargetPostID    *string `json:"target_post_id,omitempty" db:"target_post_id"`
	TargetCommentID *string `json:"target_comment_id,omitempty" db:"target_comment_id"`
	ViolationRuleID *string `json:"violation_rule_id,omitempty" db:"violation_rule_id"`
	ReasonDetail    string       `json:"reason_detail" db:"reason_detail"`
	Status          ReportStatus `json:"status" db:"status"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
}

func NewReport(reporterID string, reportType, reasonDetail string, violationRuleID *string) Report {
	return Report{
		ReporterID:      reporterID,
		ReportType:      reportType,
		ViolationRuleID: violationRuleID,
		ReasonDetail:    reasonDetail,
		Status:          ReportStatusPending,
	}
}

func (s ReportStatus) String() string {
	return string(s)
}

func ParseReportStatus(value string) ReportStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ReportStatusPending):
		return ReportStatusPending
	case string(ReportStatusReviewed):
		return ReportStatusReviewed
	case string(ReportStatusResolved):
		return ReportStatusResolved
	default:
		return ReportStatusPending
	}
}
