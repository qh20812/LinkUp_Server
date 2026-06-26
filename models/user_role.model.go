package models

import "time"

type UserRole struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	RoleID     string    `json:"role_id"`
	AssignedAt time.Time `json:"assigned_at"`
}

func NewUserRole(userID, roleID string) UserRole {
	return UserRole{
		UserID: userID,
		RoleID: roleID,
	}
}
