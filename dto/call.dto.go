package dto

import (
	"encoding/json"
	"fmt"
)

const (
	// maxSignalSize is the maximum allowed size for a WebRTC signal payload.
	// Prevents malicious clients from forwarding huge payloads via HandleSignal.
	maxSignalSize = 8192
	// maxGroupSignalSize is the cap for group-call signals. An SDP offer with
	// audio+video plus inline candidates is routinely larger than 8 KB once it
	// is JSON-escaped, so the direct-call cap would reject/close valid calls.
	maxGroupSignalSize = 32 * 1024
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

// Validate checks that the signal payload is within acceptable size bounds.
func (p *CallSignalPayload) Validate() error {
	if len(p.Signal) > maxSignalSize {
		return fmt.Errorf("signal payload quá lớn (tối đa %d bytes)", maxSignalSize)
	}
	return nil
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
	CallType       string   `json:"call_type,omitempty"`
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
	CallID string `json:"call_id"`
	// TargetUserID giới hạn signal cho đúng một peer. Bắt buộc với cuộc gọi
	// mesh ≥3 người: nếu broadcast, offer dành cho C cũng bị B áp vào PC A–B.
	TargetUserID string          `json:"target_user_id,omitempty"`
	Signal       json.RawMessage `json:"signal"`
}

// Validate checks that a group-call signal stays within the transport budget.
func (p *GroupCallSignalPayload) Validate() error {
	if len(p.Signal) > maxGroupSignalSize {
		return fmt.Errorf("signal payload quá lớn (tối đa %d bytes)", maxGroupSignalSize)
	}
	return nil
}

type GroupCallEndPayload struct {
	CallID string `json:"call_id"`
}

// ─── Call History ────────────────────────────────────────────────

// CallHistoryQuery binds to GET query params for call history list.
type CallHistoryQuery struct {
	Limit  int     `form:"limit,default=20"`
	Offset int     `form:"offset,default=0"`
	Type   *string `form:"type"`
	Status *string `form:"status"`
	Sort   string  `form:"sort,default=created_at"`
	Order  string  `form:"order,default=desc"`
}

// UserBrief is the minimal user info embedded in history items.
type UserBrief struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// CallHistoryItem is a single entry in the call history list.
type CallHistoryItem struct {
	ID        string    `json:"id"`
	OtherUser UserBrief `json:"other_user"`
	CallType  string    `json:"call_type"`
	Direction string    `json:"direction"` // "outgoing" | "incoming"
	Status    string    `json:"status"`
	IsMissed  bool      `json:"is_missed"`
	Duration  int       `json:"duration"`
	StartedAt *int64    `json:"started_at,omitempty"`
	EndedAt   *int64    `json:"ended_at,omitempty"`
	CreatedAt int64     `json:"created_at"`
}

// CallMissedPayload is the real-time WS event sent when a call is missed.
type CallMissedPayload struct {
	CallID    string `json:"call_id"`
	CallerID  string `json:"caller_id"`
	Timestamp int64  `json:"timestamp"`
}

// Nâng cấp groupcall
type GroupCallToggleMutePayload struct {
	CallID       string `json:"call_id"`
	TargetUserID string `json:"target_user_id"`
	Muted        bool   `json:"muted"`
}

// GroupCallToggleMicPayload lets a call participant change their own microphone state.
type GroupCallToggleMicPayload struct {
	CallID string `json:"call_id"`
	Muted  bool   `json:"muted"`
}

type GroupCallToggleVideoPayload struct {
	CallID       string `json:"call_id"`
	VideoEnabled bool   `json:"video_enabled"`
}

type GroupCallParticipantsPayload struct {
	CallID string `json:"call_id"`
}

type GroupCallParticipantsResponse struct {
	CallID             string   `json:"call_id"`
	Participants       []string `json:"participants"`
	Joined             []string `json:"joined"`
	ActiveParticipants []string `json:"active_participants"`
	// Profiles cho phép client hiển thị tên/ảnh thật thay vì UUID.
	Profiles map[string]GroupCallProfile `json:"profiles,omitempty"`
}

// GroupCallProfile là thông tin tối thiểu để vẽ tile của một người trong cuộc gọi nhóm.
type GroupCallProfile struct {
	DisplayName string `json:"display_name"`
	AvatarURI   string `json:"avatar_uri"`
}
