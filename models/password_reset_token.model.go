package models

import "time"

type PasswordResetToken struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Token     string     `json:"token" db:"token"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty" db:"used_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

func NewPasswordResetToken(userID, token string, expiryDuration time.Duration) PasswordResetToken {
	return PasswordResetToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(expiryDuration),
		CreatedAt: time.Now().UTC(),
	}
}

func (t PasswordResetToken) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

func (t PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}
