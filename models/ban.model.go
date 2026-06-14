package models

import "time"

type Ban struct {
	ID        int64      `json:"id" db:"id"`
	UserID    int64      `json:"user_id" db:"user_id"`
	AdminID   int64      `json:"admin_id" db:"admin_id"`
	Reason    string     `json:"reason" db:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

func NewBan(userID, adminID int64, reason string, expiresAt *time.Time) Ban {
	return Ban{UserID: userID, AdminID: adminID, Reason: reason, ExpiresAt: expiresAt}
}
