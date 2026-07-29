package dto

type VerifyEmailInput struct {
	Token string `json:"token" binding:"required"`
}

type VerifyEmailResponse struct {
	Message      string         `json:"message"`
	Verified     bool           `json:"verified"`
	AccessToken  string         `json:"access_token,omitempty"`
	RefreshToken string         `json:"refresh_token,omitempty"`
	Role         string         `json:"role,omitempty"`
}

type ResendVerificationInput struct {
	Email string `json:"email" binding:"required"`
}

type ResendVerificationResponse struct {
	Message string `json:"message"`
}
