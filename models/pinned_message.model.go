package models

import "time"

type PinnedMessage struct {
	ID        string    `json:"id"`
	ChatID    string    `json:"chat_id"`
	MessageID string    `json:"message_id"`
	PinnedBy  string    `json:"pinned_by"`
	PinnedAt  time.Time `json:"pinned_at"`
}
