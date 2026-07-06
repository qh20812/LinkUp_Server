package dto

import "time"

type CreatePolicyInput struct {
	PostWeight                  int  `json:"post_weight"`
	CommentWeight               int  `json:"comment_weight"`
	ReactionWeight              int  `json:"reaction_weight"`
	EventWeight                 int  `json:"event_weight"`
	TopContributorThreshold     int  `json:"top_contributor_threshold"`
	ModeratorPromotionThreshold int  `json:"moderator_promotion_threshold"`
	AutoPromoteEnabled          bool `json:"auto_promote_enabled"`
	BadgeEnabled                bool `json:"badge_enabled"`
}

type UpdatePolicyInput struct {
	PostWeight                  int  `json:"post_weight"`
	CommentWeight               int  `json:"comment_weight"`
	ReactionWeight              int  `json:"reaction_weight"`
	EventWeight                 int  `json:"event_weight"`
	TopContributorThreshold     int  `json:"top_contributor_threshold"`
	ModeratorPromotionThreshold int  `json:"moderator_promotion_threshold"`
	AutoPromoteEnabled          bool `json:"auto_promote_enabled"`
	BadgeEnabled                bool `json:"badge_enabled"`
}

type PolicyResponse struct {
	PostWeight                  int  `json:"post_weight"`
	CommentWeight               int  `json:"comment_weight"`
	ReactionWeight              int  `json:"reaction_weight"`
	EventWeight                 int  `json:"event_weight"`
	TopContributorThreshold     int  `json:"top_contributor_threshold"`
	ModeratorPromotionThreshold int  `json:"moderator_promotion_threshold"`
	AutoPromoteEnabled          bool `json:"auto_promote_enabled"`
	BadgeEnabled                bool `json:"badge_enabled"`
}

type ContributionResponse struct {
	UserID              string  `json:"user_id"`
	DisplayName         string  `json:"display_name"`
	AvatarURI           string  `json:"avatar_uri"`
	ValidPosts          int     `json:"valid_posts"`
	QualityComments     int     `json:"quality_comments"`
	PositiveReactions   int     `json:"positive_reactions"`
	EventParticipations int     `json:"event_participations"`
	ContributionScore   int     `json:"contribution_score"`
	BadgeType           *string `json:"badge_type"`
	PromotedToMod       bool    `json:"promoted_to_mod"`
}

type LeaderboardItem struct {
	Rank              int     `json:"rank"`
	UserID            string  `json:"user_id"`
	DisplayName       string  `json:"display_name"`
	AvatarURI         string  `json:"avatar_uri"`
	ContributionScore int     `json:"contribution_score"`
	BadgeType         *string `json:"badge_type"`
}

type CreateChallengeInput struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	Hashtag         string `json:"hashtag"`
	PointsPerPost   int    `json:"points_per_post"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	MaxParticipants *int   `json:"max_participants,omitempty"`
}

type ChallengeResponse struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	Hashtag           string    `json:"hashtag"`
	PointsPerPost     int       `json:"points_per_post"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	MaxParticipants   *int      `json:"max_participants,omitempty"`
	Status            string    `json:"status"`
	ParticipantsCount int       `json:"participants_count"`
}

type ChallengeParticipantItem struct {
	UserID            string    `json:"user_id"`
	DisplayName       string    `json:"display_name"`
	AvatarURI         string    `json:"avatar_uri"`
	PostsCount        int       `json:"posts_count"`
	TotalPointsEarned int       `json:"total_points_earned"`
	JoinedAt          time.Time `json:"joined_at"`
}
