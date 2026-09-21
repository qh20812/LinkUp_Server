package models

import "time"

// UserE2EKey stores the public half of a user's ECDH key pair, registered by
// the client. The private key never leaves the client.
// KeyVersion tăng mỗi khi public key đổi (identity/thiết bị mới) — client đối
// phương dựa vào đó để phát hiện thay đổi và re-key row của mình.
type UserE2EKey struct {
	UserID     string    `json:"user_id" gorm:"type:varchar(36);primaryKey"`
	PublicKey  string    `json:"public_key" gorm:"type:text;not null"`
	KeyVersion int       `json:"key_version" gorm:"type:int;not null;default:1"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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

// UserE2ERecovery stores the encrypted backup of the user's chat keys plus the
// salt and two verifier hashes (pin_check, recovery_check) used to unlock on a
// new device. The blob is AES-256-GCM encrypted client-side — the server can
// relay it but can never read it.
type UserE2ERecovery struct {
	UserID        string     `json:"user_id" gorm:"type:varchar(36);primaryKey"`
	Salt          string     `json:"salt" gorm:"type:text;not null"`
	Blob          string     `json:"blob" gorm:"type:longtext;not null"`
	PinCheck      string     `json:"pin_check" gorm:"type:varchar(255);not null"`
	RecoveryCheck string     `json:"recovery_check" gorm:"type:varchar(255);not null"`
	Attempts      int        `json:"attempts" gorm:"type:int;not null;default:0"`
	LockedUntil   *time.Time `json:"locked_until"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (UserE2EKey) TableName() string {
	return "user_e2e_keys"
}

func (ChatE2EKey) TableName() string {
	return "chat_e2e_keys"
}

func (UserE2ERecovery) TableName() string {
	return "user_e2e_recovery"
}
