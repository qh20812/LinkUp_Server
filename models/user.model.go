package models

import (
	"strings"
	"time"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusBanned    UserStatus = "banned"
	UserStatusSuspended UserStatus = "suspended"
)

type User struct {
	ID                string     `json:"id" db:"id"`
	Username          string     `json:"username" db:"username"`
	Email             string     `json:"email" db:"email"`
	PasswordHash      string     `json:"password_hash" db:"password_hash"`
	Status            UserStatus `json:"status" db:"status"`
	StorageQuotaBytes float64    `json:"storage_quota_bytes" db:"storage_quota_bytes"` // e.g., 1GB = 1073741824
	StorageUsedBytes  float64    `json:"storage_used_bytes" db:"storage_used_bytes"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

func NewUser(username, email, passwordHash string) User {
	return User{
		Email:        email,
		PasswordHash: passwordHash,
		Status:       UserStatusActive,
	}
}

func (u User) IsActive() bool {
	return u.Status == UserStatusActive
}

func (u User) IsBanned() bool {
	return u.Status == UserStatusBanned
}

func (u User) IsSuspended() bool {
	return u.Status == UserStatusSuspended
}

func ParseUserStatus(value string) UserStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(UserStatusActive):
		return UserStatusActive
	case string(UserStatusBanned):
		return UserStatusBanned
	case string(UserStatusSuspended):
		return UserStatusSuspended
	default:
		return UserStatusActive
	}
}

func (u User) AvailableStorageBytes() float64 {
	used := u.StorageUsedBytes
	if used < 0 {
		used = 0
	}
	available := u.StorageQuotaBytes - used
	if available < 0 {
		return 0
	}
	return available
}
