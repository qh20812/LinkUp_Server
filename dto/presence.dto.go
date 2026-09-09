package dto

// BatchPresenceInput is the input for batch presence lookup.
type BatchPresenceInput struct {
	UserIDs []string `json:"user_ids" binding:"required"`
}

// UpdatePresenceSettingsInput is the input for updating presence settings.
type UpdatePresenceSettingsInput struct {
	ActivityStatusEnabled *bool   `json:"activity_status_enabled"`
	LastSeenVisibility    *string `json:"last_seen_visibility"`
}
