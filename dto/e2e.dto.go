package dto

import "time"

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

// RekeyChatKeyRequest — client ghi đè khóa bọc của CHÍNH nó (re-key sau khi
// đối phương đổi identity). Không làm mới khóa chat.
type RekeyChatKeyRequest struct {
	WrappedKey string `json:"wrapped_key" binding:"required"`
	Nonce      string `json:"nonce"`
}

type ChatE2EKeyResponse struct {
	ChatID     string `json:"chat_id"`
	WrappedKey string `json:"wrapped_key"`
	Nonce      string `json:"nonce"`
}

// PutRecoveryRequest — client gửi backup khóa chat đã mã hóa (blob) cùng salt
// và hai hash check (PIN / recovery key) để server lưu và đối chiếu khi mở
// khóa. Blob là dữ liệu opaque, server không đọc được.
type PutRecoveryRequest struct {
	Salt          string `json:"salt" binding:"required"`
	Blob          string `json:"blob" binding:"required"`
	PinCheck      string `json:"pin_check" binding:"required"`
	RecoveryCheck string `json:"recovery_check" binding:"required"`
}

// UnlockRecoveryRequest — client gửi giá trị check (PIN hoặc recovery key) đã
// băm; server đối chiếu với PinCheck lẫn RecoveryCheck đã lưu.
type UnlockRecoveryRequest struct {
	Check string `json:"check" binding:"required"`
}

type RecoveryMetaResponse struct {
	HasBlob   bool       `json:"has_blob"`
	Salt      string     `json:"salt,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
