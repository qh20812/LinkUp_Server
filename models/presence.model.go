package models

import (
	"strings"
	"time"
)

// PresenceStatus represents the online/offline status of a user.
type PresenceStatus string

const (
	PresenceStatusOnline  PresenceStatus = "online"
	PresenceStatusOffline PresenceStatus = "offline"
)

// ParsePresenceStatus converts a string to PresenceStatus.
func ParsePresenceStatus(value string) PresenceStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(PresenceStatusOnline):
		return PresenceStatusOnline
	case string(PresenceStatusOffline):
		return PresenceStatusOffline
	default:
		return PresenceStatusOffline
	}
}

// LastSeenVisibility controls who can see a user's last seen timestamp.
type LastSeenVisibility string

const (
	LastSeenVisibilityAllFriends LastSeenVisibility = "all_friends"
	LastSeenVisibilityDmOnly     LastSeenVisibility = "dm_only"
	LastSeenVisibilityNobody     LastSeenVisibility = "nobody"
)

// ParseLastSeenVisibility converts a string to LastSeenVisibility.
func ParseLastSeenVisibility(value string) LastSeenVisibility {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(LastSeenVisibilityAllFriends):
		return LastSeenVisibilityAllFriends
	case string(LastSeenVisibilityDmOnly):
		return LastSeenVisibilityDmOnly
	case(string(LastSeenVisibilityNobody)):
		return LastSeenVisibilityNobody
	default:
		return LastSeenVisibilityAllFriends
	}
}

// PresenceCacheEntry is an in-memory cache entry for a user's presence.
// This is NOT stored in the database — it lives in the WebSocket Hub.
type PresenceCacheEntry struct {
	Status    PresenceStatus
	LastSeen  time.Time
	UpdatedAt time.Time
}

// UserPresence is the presence data returned to clients.
type UserPresence struct {
	UserID   string         `json:"user_id"`
	Status   PresenceStatus `json:"status"`
	LastSeen *time.Time     `json:"last_seen,omitempty"`
}
