package models

type ViolationRule struct {
	ID          int64  `json:"id" db:"id"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
}

func NewViolationRule(title, description string) ViolationRule {
	return ViolationRule{Title: title, Description: description}
}
