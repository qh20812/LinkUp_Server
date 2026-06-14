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
	ID        int64      `json:"id" db:"id"`
	ChatID    *int64     `json:"chat_id,omitempty" db:"chat_id"`
	CallerID  int64      `json:"caller_id" db:"caller_id"`
	CallType  CallType   `json:"call_type" db:"call_type"`
	IsGroup   bool       `json:"is_group" db:"is_group"`
	Status    CallStatus `json:"status" db:"status"`
	StartedAt time.Time  `json:"started_at" db:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty" db:"ended_at"`
}

func NewCall(callerID int64, callType CallType, isGroup bool) Call {
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
