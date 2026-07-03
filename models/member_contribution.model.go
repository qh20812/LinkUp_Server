package models

import "time"

type MemberContribution struct {
	ID                  string     `json:"id"`
	CommunityID         string     `json:"community_id"`
	UserID              string     `json:"user_id"`
	ValidPosts          int        `json:"valid_posts"`
	QualityComments     int        `json:"quality_comments"`
	PositiveReactions   int        `json:"positive_reactions"`
	EventParticipations int        `json:"event_participations"`
	ContributionScore   int        `json:"contribution_score"`
	BadgeType           *string    `json:"badge_type"`
	PromotedToMod       bool       `json:"promoted_to_mod"`
	LastCalculatedAt    time.Time  `json:"last_calculated_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
}
