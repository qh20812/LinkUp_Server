package models

// UserSetting stores per-user settings/preferences that are not part of the
// public profile. Absence of a row means defaults (DiscoverableInSearch=true,
// AllowStrangerMessages=false, Theme=light, Language=vi).
type UserSetting struct {
	UserID                string `json:"user_id" gorm:"type:varchar(36);primaryKey"`
	DiscoverableInSearch  bool   `json:"discoverable_in_search"`
	AllowStrangerMessages bool   `json:"allow_stranger_messages"`
	Theme                 string `json:"theme"`
	Language              string `json:"language"`
	// Presence settings
	ActivityStatusEnabled bool   `json:"activity_status_enabled" gorm:"default:true"`
	LastSeenVisibility    string `json:"last_seen_visibility" gorm:"type:VARCHAR(20);default:'all_friends'"`
}

func DefaultUserSetting(userID string) UserSetting {
	return UserSetting{
		UserID:                userID,
		DiscoverableInSearch:  true,
		AllowStrangerMessages: false,
		Theme:                 "light",
		Language:              "vi",
		ActivityStatusEnabled: true,
		LastSeenVisibility:    string(LastSeenVisibilityAllFriends),
	}
}
