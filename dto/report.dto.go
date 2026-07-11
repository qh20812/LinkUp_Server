package dto

type CreateReportInput struct {
	TargetType     string  `json:"target_type"`
	TargetID       string  `json:"target_id"`
	ReportType     string  `json:"report_type"`
	ViolationRuleID *string `json:"violation_rule_id,omitempty"`
	ReasonDetail   string  `json:"reason_detail"`
}

type CreateReportResponse struct {
	Message string `json:"message"`
}

type UpdateReportInput struct {
	ReportType     string  `json:"report_type"`
	ViolationRuleID *string `json:"violation_rule_id,omitempty"`
	ReasonDetail   string  `json:"reason_detail"`
}

type UpdateReportResponse struct {
	Message string `json:"message"`
}
