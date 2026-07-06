package dto

import "encoding/json"

type CallInitiatePayload struct {
	CalleeID string `json:"callee_id"`
	CallType string `json:"call_type"`
}

type CallActionPayload struct {
	CallID string `json:"call_id"`
	Action string `json:"action"`
}

type CallIncomingPayload struct {
	CallID    string `json:"call_id"`
	CallerID  string `json:"caller_id"`
	CallType  string `json:"call_type"`
	Timestamp int64  `json:"timestamp"`
}

type CallStatusPayload struct {
	CallID    string  `json:"call_id"`
	Status    string  `json:"status"`
	CallerID  string  `json:"caller_id"`
	CalleeID  string  `json:"callee_id"`
	CallType  string  `json:"call_type"`
	StartedAt *int64  `json:"started_at,omitempty"`
	EndedAt   *int64  `json:"ended_at,omitempty"`
	Duration  int     `json:"duration,omitempty"`
}

type CallBusyPayload struct {
	CalleeID string `json:"callee_id"`
}

type CallSignalPayload struct {
	CallID   string          `json:"call_id"`
	SenderID string          `json:"sender_id,omitempty"`
	Signal   json.RawMessage `json:"signal"`
}

type ToggleMuteRequest struct {
	Muted bool `json:"muted"`
}
