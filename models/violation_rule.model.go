package models

type ViolationRule struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewViolationRule(title, description string) ViolationRule {
	return ViolationRule{Title: title, Description: description}
}
