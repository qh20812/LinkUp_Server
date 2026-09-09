package utils

import (
	"context"
	"fmt"
	errorsapp "linkup/errors"
	"linkup/models"
)

// ─── Platform roles ─────────────────────────────────────────────────────────

type PlatformRole string

const (
	RoleUser       PlatformRole = "USER"
	RoleAdmin      PlatformRole = "ADMIN"
	RoleSuperAdmin PlatformRole = "SUPER_ADMIN"
)

// ─── Platform role hierarchy (USER < ADMIN < SUPER_ADMIN) ───────────────────

var platformRoleLevel = map[PlatformRole]int{
	RoleUser:       0,
	RoleAdmin:      1,
	RoleSuperAdmin: 2,
}

// HasPlatformRole returns true if userRole >= requiredRole.
func HasPlatformRole(userRole, requiredRole PlatformRole) bool {
	return platformRoleLevel[userRole] >= platformRoleLevel[requiredRole]
}

// RequirePlatformRole returns error if userRole < requiredRole.
func RequirePlatformRole(userRole, requiredRole PlatformRole) error {
	if !HasPlatformRole(userRole, requiredRole) {
		return errorsapp.Wrap(errorsapp.ErrCodeRbacPermissionDenied, fmt.Errorf("bạn cần quyền %s để thực hiện hành động này", requiredRole))
	}
	return nil
}

// ─── Group role hierarchy (GROUP_MEMBER < GROUP_MOD < GROUP_ADMIN) ──────────

var groupRoleLevel = map[models.GroupRole]int{
	models.GroupRoleMember: 0,
	models.GroupRoleMod:    1,
	models.GroupRoleAdmin:  2,
}

// HasGroupRole returns true if userRole >= requiredRole.
func HasGroupRole(userRole, requiredRole models.GroupRole) bool {
	return groupRoleLevel[userRole] >= groupRoleLevel[requiredRole]
}

// RequireGroupRole returns error if userRole < requiredRole.
func RequireGroupRole(userRole, requiredRole models.GroupRole) error {
	if !HasGroupRole(userRole, requiredRole) {
		return errorsapp.Wrap(errorsapp.ErrCodeRbacPermissionDenied, fmt.Errorf("bạn cần quyền %s trong cộng đồng để thực hiện hành động này", requiredRole))
	}
	return nil
}

// ─── GroupRoleChecker — DB-backed group role checks ─────────────────────────

// GroupRoleChecker queries user_roles table for community-level role checks.
type GroupRoleChecker struct {
	getUserRole func(ctx context.Context, communityID, userID string) (models.GroupRole, error)
}

// NewGroupRoleChecker creates a checker from any function that queries group roles.
// Usage: checker := NewGroupRoleChecker(communityRepo.GetUserRole)
func NewGroupRoleChecker(getUserRole func(context.Context, string, string) (models.GroupRole, error)) *GroupRoleChecker {
	return &GroupRoleChecker{getUserRole: getUserRole}
}

// GetUserRole returns the user's group role in a community.
func (c *GroupRoleChecker) GetUserRole(ctx context.Context, communityID, userID string) (models.GroupRole, error) {
	return c.getUserRole(ctx, communityID, userID)
}

// RequireRole returns error if user doesn't have required group role.
func (c *GroupRoleChecker) RequireRole(ctx context.Context, communityID, userID string, required models.GroupRole) error {
	role, err := c.getUserRole(ctx, communityID, userID)
	if err != nil {
		return err
	}
	return RequireGroupRole(role, required)
}

// IsAdmin returns true if user is GROUP_ADMIN in the community.
func (c *GroupRoleChecker) IsAdmin(ctx context.Context, communityID, userID string) (bool, error) {
	role, err := c.getUserRole(ctx, communityID, userID)
	if err != nil {
		return false, err
	}
	return HasGroupRole(role, models.GroupRoleAdmin), nil
}

// IsModOrAbove returns true if user is GROUP_MOD or GROUP_ADMIN.
func (c *GroupRoleChecker) IsModOrAbove(ctx context.Context, communityID, userID string) (bool, error) {
	role, err := c.getUserRole(ctx, communityID, userID)
	if err != nil {
		return false, err
	}
	return HasGroupRole(role, models.GroupRoleMod), nil
}