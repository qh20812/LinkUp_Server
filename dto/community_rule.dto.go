package dto

import "linkup/models"

type CreateCommunityRuleInput struct {
	Category models.RuleCategory `json:"category" binding:"required,oneof=conduct prohibited guidelines"`
	Title    string              `json:"title" binding:"required,min=5,max=255"`
	Content  string              `json:"content" binding:"max=2000"`
	Position int                 `json:"position" binding:"min=0"`
}

type UpdateCommunityRuleInput struct {
	Category models.RuleCategory `json:"category" binding:"omitempty,oneof=conduct prohibited guidelines"`
	Title    string              `json:"title" binding:"omitempty,min=5,max=255"`
	Content  string              `json:"content" binding:"max=2000"`
	Position *int                `json:"position,omitempty" binding:"omitempty,min=0"`
}
