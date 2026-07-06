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
	CallStatusCalling   CallStatus = "calling"
	CallStatusRinging   CallStatus = "ringing"
	CallStatusConnected CallStatus = "connected"
	CallStatusEnded     CallStatus = "ended"
	CallStatusMissed    CallStatus = "missed"
	CallStatusRejected  CallStatus = "rejected"
	CallStatusBusy      CallStatus = "busy"
)

type Call struct {
	ID          string     `json:"id"`
	CallerID    string     `json:"caller_id"`
	CalleeID    string     `json:"callee_id"`
	CallType    CallType   `json:"call_type"`
	IsGroup     bool       `json:"is_group"`
	Status      CallStatus `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	Duration    int        `json:"duration"`
	MutedCaller        bool       `json:"muted_caller"`
	MutedCallee        bool       `json:"muted_callee"`
	VideoEnabledCaller bool       `json:"video_enabled_caller"`
	VideoEnabledCallee bool       `json:"video_enabled_callee"`
	CreatedAt          time.Time  `json:"created_at"`
}

func NewCall(callerID, calleeID string, callType CallType, isGroup bool) Call {
	if callType == "" {
		callType = CallTypeVoice
	}
	return Call{
		CallerID:           callerID,
		CalleeID:           calleeID,
		CallType:           callType,
		IsGroup:            isGroup,
		Status:             CallStatusCalling,
		VideoEnabledCaller: false,
		VideoEnabledCallee: false,
		CreatedAt:          time.Now().UTC(),
	}
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
	case string(CallStatusCalling):
		return CallStatusCalling
	case string(CallStatusRinging):
		return CallStatusRinging
	case string(CallStatusConnected):
		return CallStatusConnected
	case string(CallStatusEnded):
		return CallStatusEnded
	case string(CallStatusMissed):
		return CallStatusMissed
	case string(CallStatusRejected):
		return CallStatusRejected
	case string(CallStatusBusy):
		return CallStatusBusy
	default:
		return CallStatusEnded
	}
}
