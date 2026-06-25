package models

import (
	"strings"
	"time"
)

type CallType string

type CallStatus string

const (
	CallTypeVoice CallType = "voice"
	CallTypeVideo CallType = "video"
)

const (
	CallStatusMissed    CallStatus = "missed"
	CallStatusCompleted CallStatus = "completed"
	CallStatusDeclined  CallStatus = "declined"
)

type Call struct {
	ID        string     `json:"id"`
	ChatID    *string    `json:"chat_id,omitempty"`
	CallerID  string     `json:"caller_id"`
	CallType  CallType   `json:"call_type"`
	IsGroup   bool       `json:"is_group"`
	Status    CallStatus `json:"status"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

func NewCall(callerID string, callType CallType, isGroup bool) Call {
	if callType == "" {
		callType = CallTypeVoice
	}
	return Call{CallerID: callerID, CallType: callType, IsGroup: isGroup, Status: CallStatusCompleted}
}

func (c CallType) String() string {
	return string(c)
}

func ParseCallType(value string) CallType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(CallTypeVoice):
		return CallTypeVoice
	case string(CallTypeVideo):
		return CallTypeVideo
	default:
		return CallTypeVoice
	}
}

func (s CallStatus) String() string {
	return string(s)
}

func ParseCallStatus(value string) CallStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(CallStatusMissed):
		return CallStatusMissed
	case string(CallStatusCompleted):
		return CallStatusCompleted
	case string(CallStatusDeclined):
		return CallStatusDeclined
	default:
		return CallStatusCompleted
	}
}
