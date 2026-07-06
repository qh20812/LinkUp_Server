package models

import "time"

type CommunityPolicy struct {
	ID                          string     `json:"id"`
	CommunityID                 string     `json:"community_id"`
	PostWeight                  int        `json:"post_weight"`
	CommentWeight               int        `json:"comment_weight"`
	ReactionWeight              int        `json:"reaction_weight"`
	EventWeight                 int        `json:"event_weight"`
	TopContributorThreshold     int        `json:"top_contributor_threshold"`
	ModeratorPromotionThreshold int        `json:"moderator_promotion_threshold"`
	AutoPromoteEnabled          bool       `json:"auto_promote_enabled"`
	BadgeEnabled                bool       `json:"badge_enabled"`
	CreatedAt                   time.Time  `json:"created_at"`
	UpdatedAt                   *time.Time `json:"updated_at,omitempty"`
}
