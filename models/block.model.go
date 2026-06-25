package models

import "time"

type Block struct {
	ID            string     `json:"id"`
	UserID        string `json:"user_id"`
	BlockedUserID string `json:"blocked_user_id"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewBlock(userID, blockedUserID string) Block {
	return Block{UserID: userID, BlockedUserID: blockedUserID}
}
