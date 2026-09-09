package dto

type RegisterE2EKeyRequest struct {
	PublicKey string `json:"public_key" binding:"required"`
}

type ChatE2EKeyInput struct {
	ChatID     string `json:"chat_id" binding:"required"`
	UserID     string `json:"user_id" binding:"required"`
	WrappedKey string `json:"wrapped_key" binding:"required"`
	Nonce      string `json:"nonce"`
}

type ChatE2EKeyBatchRequest struct {
	Keys []ChatE2EKeyInput `json:"keys" binding:"required"`
}

type ChatE2EKeyResponse struct {
	ChatID     string `json:"chat_id"`
	WrappedKey string `json:"wrapped_key"`
	Nonce      string `json:"nonce"`
}
