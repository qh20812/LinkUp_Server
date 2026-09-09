package dto

import "time"

type PrivacySettingsResponse struct {
	DiscoverableInSearch  bool `json:"discoverable_in_search"`
	AllowStrangerMessages bool `json:"allow_stranger_messages"`
}

type UpdatePrivacyInput struct {
	DiscoverableInSearch  *bool `json:"discoverable_in_search"`
	AllowStrangerMessages *bool `json:"allow_stranger_messages"`
}

type StorageInfoResponse struct {
	QuotaBytes float64 `json:"quota_bytes"`
	UsedBytes  float64 `json:"used_bytes"`
	AvailBytes float64 `json:"avail_bytes"`
}

type DeactivateInput struct {
	Password string `json:"password" binding:"required"`
}

type AppearanceSettingsResponse struct {
	Theme    string `json:"theme"`
	Language string `json:"language"`
}

type UpdateAppearanceInput struct {
	Theme    *string `json:"theme"`
	Language *string `json:"language"`
}

type SessionDTO struct {
	ID           string    `json:"id"`
	DeviceName   string    `json:"device_name"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	IsCurrent    bool      `json:"is_current"`
}
