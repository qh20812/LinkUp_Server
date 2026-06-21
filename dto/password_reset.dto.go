package dto

type ForgotPasswordInput struct {
	Email string `json:"email"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

type VerifyResetTokenInput struct {
	Token string `json:"token"`
}

type VerifyResetTokenResponse struct {
	Message string `json:"message"`
	Valid   bool   `json:"valid"`
}

type ResetPasswordInput struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}
