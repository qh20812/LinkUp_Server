package models

import "time"

type RuleCategory string

const (
	RuleConduct    RuleCategory = "conduct"
	RuleProhibited RuleCategory = "prohibited"
	RuleGuidelines RuleCategory = "guidelines"
)

type CommunityRule struct {
	ID          string       `json:"id"`
	CommunityID string       `json:"community_id"`
	Category    RuleCategory `json:"category"`
	Title       string       `json:"title"`
	Content     string       `json:"content"`
	Position    int          `json:"position"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   *time.Time   `json:"updated_at,omitempty"`
}

func NewCommunityRule(communityID string, category RuleCategory, title, content string, position int) CommunityRule {
	return CommunityRule{
		CommunityID: communityID,
		Category:    category,
		Title:       title,
		Content:     content,
		Position:    position,
	}
}

func (c RuleCategory) String() string {
	return string(c)
}
