package dto

import (
	"encoding/json"
)

type IceServer struct {
	URLs       string `json:"urls"`
	Username   string `json:"username,omitempty"`
	Credential string `json:"credential,omitempty"`
}

type IceServersResponse struct {
	IceServers []IceServer `json:"ice_servers"`
}

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
	CallID             string `json:"call_id"`
	Status             string `json:"status"`
	CallerID           string `json:"caller_id"`
	CalleeID           string `json:"callee_id"`
	CallType           string `json:"call_type"`
	VideoEnabledCaller bool   `json:"video_enabled_caller"`
	VideoEnabledCallee bool   `json:"video_enabled_callee"`
	StartedAt          *int64 `json:"started_at,omitempty"`
	EndedAt            *int64 `json:"ended_at,omitempty"`
	Duration           int    `json:"duration,omitempty"`
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

type ToggleVideoRequest struct {
	VideoEnabled bool `json:"video_enabled"`
}

// ==== GROUP CALL DTO ====
type GroupCallInitiatePayload struct {
	ChatID         string   `json:"chat_id"`
	ParticipantIDs []string `json:"participant_ids,omitempty"`
}

type GroupCallJoinRequestPayload struct {
	CallID string `json:"call_id"`
}

type GroupCallApprovePayload struct {
	CallID string `json:"call_id"`
	UserID string `json:"user_id"`
}

type GroupCallSignalPayload struct {
	CallID string          `json:"call_id"`
	Signal json.RawMessage `json:"signal"`
}

type GroupCallEndPayload struct {
	CallID string `json:"call_id"`
}
