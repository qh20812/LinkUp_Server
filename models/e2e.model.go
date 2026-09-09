package models

import "time"

// UserE2EKey stores the public half of a user's ECDH key pair, registered by
// the client. The private key never leaves the client.
type UserE2EKey struct {
	UserID    string    `json:"user_id" gorm:"type:varchar(36);primaryKey"`
	PublicKey string    `json:"public_key" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChatE2EKey stores the per-participant wrapped copy of the chat's symmetric
// AES key. The server can store and relay it but can never unwrap it.
type ChatE2EKey struct {
	ChatID     string    `json:"chat_id" gorm:"type:varchar(36);primaryKey"`
	UserID     string    `json:"user_id" gorm:"type:varchar(36);primaryKey"`
	WrappedKey string    `json:"wrapped_key" gorm:"type:text;not null"`
	Nonce      string    `json:"nonce" gorm:"type:varchar(64)"`
	CreatedAt  time.Time `json:"created_at"`
}

func (UserE2EKey) TableName() string {
	return "user_e2e_keys"
}

func (ChatE2EKey) TableName() string {
	return "chat_e2e_keys"
}
