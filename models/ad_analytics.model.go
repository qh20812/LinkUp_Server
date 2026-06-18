package models

import "time"

type AdAnalytics struct {
	ID         string    `json:"id" db:"id"`
	AdID       string    `json:"ad_id" db:"ad_id"`
	UserID     *string   `json:"user_id,omitempty" db:"user_id"`
	ActionType string    `json:"action_type" db:"action_type"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

func NewAdAnalytics(adID string, userID *string, actionType, ipAddress string) AdAnalytics {
	return AdAnalytics{AdID: adID, UserID: userID, ActionType: actionType, IPAddress: ipAddress}
}
