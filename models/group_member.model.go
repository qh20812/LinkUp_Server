package models

import (
	"strings"
	"time"
)

type GroupMemberRole string

const (
	GroupMemberRoleAdmin  GroupMemberRole = "admin"
	GroupMemberRoleMod    GroupMemberRole = "mod"
	GroupMemberRoleMember GroupMemberRole = "member"
)

type GroupMember struct {
	ID          int64           `json:"id" db:"id"`
	CommunityID int64           `json:"community_id" db:"community_id"`
	UserID      int64           `json:"user_id" db:"user_id"`
	Role        GroupMemberRole `json:"role" db:"role"`
	Points      int             `json:"points" db:"points"`
	JoinedAt    time.Time       `json:"joined_at" db:"joined_at"`
}

func NewGroupMember(communityID, userID int64, role GroupMemberRole) GroupMember {
	if role == "" {
		role = GroupMemberRoleMember
	}
	return GroupMember{CommunityID: communityID, UserID: userID, Role: role}
}

func (r GroupMemberRole) String() string {
	return string(r)
}

func ParseGroupMemberRole(value string) GroupMemberRole {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(GroupMemberRoleAdmin):
		return GroupMemberRoleAdmin
	case string(GroupMemberRoleMod):
		return GroupMemberRoleMod
	case string(GroupMemberRoleMember):
		return GroupMemberRoleMember
	default:
		return GroupMemberRoleMember
	}
}
