package models

import "time"

// ScopeType identifies what entity a role is scoped to.
// NULL means platform-level (global) role.
type ScopeType string

const (
	ScopeTypeCommunity ScopeType = "community"
	ScopeTypeChat      ScopeType = "chat"
)

// UserRole maps a user to a role, optionally scoped to a community or chat.
// - Platform roles (USER, ADMIN, SUPER_ADMIN): scope_id=NULL, scope_type=NULL
// - Community roles (GROUP_ADMIN, etc.): scope_id=communityID, scope_type="community"
// - Chat roles (CHAT_ADMIN, etc.): scope_id=chatID, scope_type="chat"
type UserRole struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	RoleID     string     `json:"role_id"`
	ScopeID    *string    `json:"scope_id,omitempty"`
	ScopeType  *ScopeType `json:"scope_type,omitempty"`
	AssignedAt time.Time  `json:"assigned_at"`
}

func NewUserRole(userID, roleID string) UserRole {
	return UserRole{
		UserID: userID,
		RoleID: roleID,
	}
}

func NewScopedUserRole(userID, roleID string, scopeID string, scopeType ScopeType) UserRole {
	return UserRole{
		UserID:    userID,
		RoleID:    roleID,
		ScopeID:   &scopeID,
		ScopeType: &scopeType,
	}
}
