package models

type ViolationRule struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
}

func NewViolationRule(title, description string) ViolationRule {
	return ViolationRule{Title: title, Description: description}
}
