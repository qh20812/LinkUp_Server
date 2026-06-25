package models

import "time"

type Ban struct {
	ID        string      `json:"id"`
	UserID    string `json:"user_id"`
	AdminID   string `json:"admin_id"`
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func NewBan(userID, adminID string, reason string, expiresAt *time.Time) Ban {
	return Ban{UserID: userID, AdminID: adminID, Reason: reason, ExpiresAt: expiresAt}
}
