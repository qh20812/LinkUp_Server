package dto

import "time"

type BlockUserInput struct {
	TargetUserID string `json:"target_user_id"`
}

type BlockUserResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type BlockedUserResponse struct {
	BlockedUserID string    `json:"blocked_user_id"`
	BlockedAt     time.Time `json:"blocked_at"`
}
