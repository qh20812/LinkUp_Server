package models

import "time"

type ReportStatus string

type Report struct {
	ID              int64        `json:"id" db:"id"`
	ReporterID      int64        `json:"reporter_id" db:"reporter_id"`
	ReportType      string       `json:"report_type" db:"report_type"`
	TargetUserID    *int64       `json:"target_user_id,omitempty" db:"target_user_id"`
	TargetPostID    *int64       `json:"target_post_id,omitempty" db:"target_post_id"`
	TargetCommentID *int64       `json:"target_comment_id,omitempty" db:"target_comment_id"`
	ViolationRuleID *int64       `json:"violation_rule_id,omitempty" db:"violation_rule_id"`
	ReasonDetail    string       `json:"reason_detail" db:"reason_detail"`
	Status          ReportStatus `json:"status" db:"status"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
}

const (
	ReportStatusPending  ReportStatus = "pending"
	ReportStatusReviewed ReportStatus = "reviewed"
	ReportStatusResolved ReportStatus = "resolved"
)

func NewReport(reporterID int64, reportType, reasonDetail string, violationRuleID *int64) Report {
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
