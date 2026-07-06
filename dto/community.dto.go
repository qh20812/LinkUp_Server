package dto

import "time"

type CreateCommunityInput struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
	AvatarURI   string `json:"avatar_uri" binding:"omitempty,url"`
}

type JoinRequestItem struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURI   string    `json:"avatar_uri"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type JoinRequestListResponse struct {
	Requests []JoinRequestItem `json:"requests"`
}

type UpdateMemberRoleInput struct {
	Role string `json:"role" binding:"required,oneof=GROUP_ADMIN GROUP_MOD GROUP_MEMBER"`
}

type CommunityMemberItem struct {
	UserID            string    `json:"user_id"`
	DisplayName       string    `json:"display_name"`
	AvatarURI         string    `json:"avatar_uri"`
	Role              string    `json:"role"`
	JoinedAt          time.Time `json:"joined_at"`
	ContributionScore int       `json:"contribution_score"`
	BadgeType         *string   `json:"badge_type,omitempty"`
}

type CommunityMemberListResponse struct {
	Members []CommunityMemberItem `json:"members"`
}

type KickMemberInput struct {
	Reason string `json:"reason" binding:"required,min=3,max=500"`
}

type JoinResult struct {
	RequestID    string `json:"request_id,omitempty"`
	AutoApproved bool   `json:"auto_approved"`
}

type LeaveCommunityInput struct {
	Quiet bool `json:"quiet" binding:"omitempty"`
}
