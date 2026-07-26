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

const DefaultStorageQuotaBytes float64 = 2147483648

type User struct {
	ID                string     `json:"id"`
	Username          string     `json:"username"`
	Email             string     `json:"email"`
	PasswordHash      string     `json:"password_hash"`
	Status            UserStatus `json:"status"`
	StorageQuotaBytes float64    `json:"storage_quota_bytes"`
	StorageUsedBytes  float64    `json:"storage_used_bytes"`
	TokenVersion      int        `json:"token_version" gorm:"default:0"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty"`
}

func NewUser(username, email, passwordHash string) User {
	return User{
		Email:        email,
		PasswordHash: passwordHash,
		Status:       UserStatusActive,
		StorageQuotaBytes: DefaultStorageQuotaBytes,
		StorageUsedBytes: 0,
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
