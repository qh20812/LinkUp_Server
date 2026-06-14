package models

import "time"

type AdAnalytics struct {
	ID        int64      `json:"id" db:"id"`
	AdID      int64      `json:"ad_id" db:"ad_id"`
	UserID    *int64     `json:"user_id,omitempty" db:"user_id"`
	ActionType string    `json:"action_type" db:"action_type"`
	IPAddress string     `json:"ip_address" db:"ip_address"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

func NewAdAnalytics(adID int64, userID *int64, actionType, ipAddress string) AdAnalytics {
	return AdAnalytics{AdID: adID, UserID: userID, ActionType: actionType, IPAddress: ipAddress}
}
