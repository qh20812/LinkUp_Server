package models

import (
	"strings"
)

type RoleName string

const (
	RoleSuperAdmin RoleName = "SUPER_ADMIN"
	RoleAdmin      RoleName = "ADMIN"
	RoleUser       RoleName = "USER"
)

type Role struct {
	ID          string   `json:"id" db:"id"`
	Name        RoleName `json:"name" db:"name"`
	Description string   `json:"description" db:"description"`
}

func NewRole(name RoleName, description string) Role {
	return Role{
		Name:        name,
		Description: description,
	}
}

func (r RoleName) String() string {
	return string(r)
}

// ChatRole — participant roles within a chat
type ChatRole string

const (
	ChatRoleAdmin  ChatRole = "CHAT_ADMIN"
	ChatRoleMember ChatRole = "CHAT_MEMBER"
)

func (r ChatRole) String() string {
	return string(r)
}

func ParseChatRole(value string) ChatRole {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ChatRoleAdmin):
		return ChatRoleAdmin
	default:
		return ChatRoleMember
	}
}

// GroupRole — member roles within a community / group
type GroupRole string

const (
	GroupRoleAdmin  GroupRole = "GROUP_ADMIN"
	GroupRoleMod    GroupRole = "GROUP_MOD"
	GroupRoleMember GroupRole = "GROUP_MEMBER"
)

func (r GroupRole) String() string {
	return string(r)
}

func ParseGroupRole(value string) GroupRole {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(GroupRoleAdmin):
		return GroupRoleAdmin
	case string(GroupRoleMod):
		return GroupRoleMod
	default:
		return GroupRoleMember
	}
}

// CommunityRole — member roles within a community
type CommunityRole string

const (
	CommunityRoleAdmin  CommunityRole = "COMMUNITY_ADMIN"
	CommunityRoleMember CommunityRole = "COMMUNITY_MEMBER"
)

func (r CommunityRole) String() string {
	return string(r)
}

func ParseCommunityRole(value string) CommunityRole {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(CommunityRoleAdmin):
		return CommunityRoleAdmin
	default:
		return CommunityRoleMember
	}
}