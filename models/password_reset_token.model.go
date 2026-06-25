package models

import (
	"linkup/utils"
	"time"
)

type PasswordResetToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func NewPasswordResetToken(userID, token string, expiryDuration time.Duration) PasswordResetToken {
	return PasswordResetToken{
		ID:        utils.GenerateUUID(),
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
