package dto

import "time"

type RegisterInput struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthUserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshTTLIn int64  `json:"refresh_ttl_in"`
}

type AuthResponse struct {
	User        AuthUserResponse `json:"user"`
	Tokens      TokenResponse    `json:"tokens,omitempty"`
	Storage     StorageInfo      `json:"storage,omitempty"`
	VerifyEmail bool             `json:"verify_email,omitempty"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type ChangePasswordResponse struct {
	Message string `json:"message"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type StorageInfo struct {
	QuotaBytes float64 `json:"quota_bytes"`
	UsedBytes  float64 `json:"used_bytes"`
	AvailBytes float64 `json:"available_bytes"`
}
