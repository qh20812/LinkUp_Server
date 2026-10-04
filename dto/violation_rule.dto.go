package dto

type CreateViolationRuleInput struct {
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description"`
	ApplicableTo string `json:"applicable_to"`
	Severity     string `json:"severity"`
	SortOrder    int    `json:"sort_order"`
}

type UpdateViolationRuleInput struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	ApplicableTo string `json:"applicable_to"`
	Severity     string `json:"severity"`
	SortOrder    *int   `json:"sort_order"`
}

type ViolationRuleResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ApplicableTo string `json:"applicable_to"`
	Severity     string `json:"severity"`
	SortOrder    int    `json:"sort_order"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
}

type ListViolationRulesResponse struct {
	Rules []ViolationRuleResponse `json:"rules"`
	Total int64                   `json:"total"`
}
