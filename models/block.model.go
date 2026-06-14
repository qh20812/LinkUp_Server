package models

import "time"

type Block struct {
	ID            int64     `json:"id" db:"id"`
	UserID        int64     `json:"user_id" db:"user_id"`
	BlockedUserID int64     `json:"blocked_user_id" db:"blocked_user_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

func NewBlock(userID, blockedUserID int64) Block {
	return Block{UserID: userID, BlockedUserID: blockedUserID}
}
