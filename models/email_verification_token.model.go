package models

import "time"

type EmailVerificationToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func NewEmailVerificationToken(userID, token string, expiryDuration time.Duration) EmailVerificationToken {
	return EmailVerificationToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(expiryDuration),
		CreatedAt: time.Now().UTC(),
	}
}

func (t EmailVerificationToken) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

func (t EmailVerificationToken) IsUsed() bool {
	return t.UsedAt != nil
}
