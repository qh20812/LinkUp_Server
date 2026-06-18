package models

import "time"

type GroupMember struct {
	ID          string    `json:"id" db:"id"`
	CommunityID string    `json:"community_id" db:"community_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Role        GroupRole `json:"role" db:"role"`
	Points      int       `json:"points" db:"points"`
	JoinedAt    time.Time `json:"joined_at" db:"joined_at"`
}

func NewGroupMember(communityID, userID string, role GroupRole) GroupMember {
	if role == "" {
		role = GroupRoleMember
	}
	return GroupMember{CommunityID: communityID, UserID: userID, Role: role}
}
