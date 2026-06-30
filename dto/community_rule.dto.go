package dto

import "linkup/models"

type CreateCommunityRuleInput struct {
	Category models.RuleCategory `json:"category" binding:"required,oneof=conduct prohibited guidelines"` // Loại quy tắc, bắt buộc và phải là một trong các giá trị: conduct, prohibited, guidelines
	Title    string              `json:"title" binding:"required,min=5,max=255"` // Tiêu đề của quy tắc, bắt buộc và có độ dài từ 5 đến 255 ký tự
	Content  string              `json:"content" binding:"max=2000"` // Nội dung của quy tắc, có thể để trống
	Position int                 `json:"position" binding:"min=0"` // Vị trí của quy tắc trong danh sách, bắt đầu từ 0
}

type UpdateCommunityRuleInput struct {
	Category models.RuleCategory `json:"category" binding:"omitempty,oneof=conduct prohibited guidelines"`
	Title    string              `json:"title" binding:"omitempty,min=5,max=255"`
	Content  string              `json:"content" binding:"max=2000"`
	Position *int                `json:"position,omitempty" binding:"omitempty,min=0"`
}
