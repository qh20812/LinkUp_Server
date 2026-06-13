package models

import "time"

type UserRole struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	RoleID     int64     `json:"role_id" db:"role_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
}

func NewUserRole(userID, roleID int64) UserRole {
	return UserRole{
		UserID: userID,
		RoleID: roleID,
	}
}
