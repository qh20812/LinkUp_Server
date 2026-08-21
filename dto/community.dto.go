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

// ── Invite Code ──

type CreateInviteCodeInput struct {
	MaxUses   int        `json:"max_uses" binding:"omitempty,min=0"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" binding:"omitempty"`
}

type InviteCodeResponse struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	MaxUses   int        `json:"max_uses"`
	UsedCount int        `json:"used_count"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}

// ── Invitation ──

type SendInvitationInput struct {
	InviteeID string `json:"invitee_id" binding:"required,uuid"`
}

type InvitationItem struct {
	ID            string    `json:"id"`
	CommunityID   string    `json:"community_id"`
	CommunityName string    `json:"community_name"`
	InviterID     string    `json:"inviter_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type RespondInvitationInput struct {
	Accept bool `json:"accept" binding:"required"`
}

// ── Join Community (mở rộng) ──

type JoinCommunityInput struct {
	InviteCode   string `json:"code,omitempty" binding:"omitempty,min=6,max=6"`
	InvitationID string `json:"invitation_id,omitempty" binding:"omitempty,uuid"`
}

// CommunityTransferOwnershipInput là input cho tính năng chuyển quyền sở hữu cộng đồng.
type CommunityTransferOwnershipInput struct {
	TargetUserID string `json:"target_user_id" binding:"required"`
	KeepAdmin    bool   `json:"keep_admin"`
}

// ── User-facing Community List/Detail ──

type CommunityListItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	AvatarURI   string    `json:"avatar_uri"`
	Privacy     string    `json:"privacy"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type CommunityListResponse struct {
	Communities []CommunityListItem `json:"communities"`
	Total       int64               `json:"total"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
}

type CommunityDetailResponse struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	AvatarURI        string    `json:"avatar_uri"`
	BackgroundURI    string    `json:"background_uri"`
	CreatorID        string    `json:"creator_id"`
	CreatorName      string    `json:"creator_name"`
	Privacy          string    `json:"privacy"`
	AutoApprove      bool      `json:"auto_approve"`
	MemberCount      int       `json:"member_count"`
	MembershipStatus string    `json:"membership_status"`
	UserMemberRole   string    `json:"user_member_role,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}
