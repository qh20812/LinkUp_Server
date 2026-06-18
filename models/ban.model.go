package models

import "time"

type Ban struct {
	ID        string      `json:"id" db:"id"`
	UserID    string `json:"user_id" db:"user_id"`
	AdminID   string `json:"admin_id" db:"admin_id"`
	Reason    string     `json:"reason" db:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

func NewBan(userID, adminID string, reason string, expiresAt *time.Time) Ban {
	return Ban{UserID: userID, AdminID: adminID, Reason: reason, ExpiresAt: expiresAt}
}
