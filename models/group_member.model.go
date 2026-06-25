package models

import "time"

type GroupMember struct {
	ID          string    `json:"id"`
	CommunityID string    `json:"community_id"`
	UserID      string    `json:"user_id"`
	Role        GroupRole `json:"role"`
	Points      int       `json:"points"`
	JoinedAt    time.Time `json:"joined_at"`
}

func NewGroupMember(communityID, userID string, role GroupRole) GroupMember {
	if role == "" {
		role = GroupRoleMember
	}
	return GroupMember{CommunityID: communityID, UserID: userID, Role: role}
}
