package models

import "time"

// UserSession represents a logged-in session. The session ID is embedded in the
// access/refresh tokens as the JWT `jti` claim and used to enforce per-session
// revocation (list / revoke specific devices).
type UserSession struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	UserID       string     `json:"user_id" gorm:"index"`
	DeviceName   string     `json:"device_name"`
	IPAddress    string     `json:"ip_address"`
	UserAgent    string     `json:"user_agent"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
	LastActiveAt time.Time  `json:"last_active_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}

func (s UserSession) IsActive() bool {
	return s.RevokedAt == nil && time.Now().UTC().Before(s.ExpiresAt)
}
