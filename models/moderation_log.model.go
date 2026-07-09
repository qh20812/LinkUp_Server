package models

import (
	"strings"
	"time"
)

type ModerationAction string

const (
	ModerationActionWarn          ModerationAction = "warn"
	ModerationActionBan           ModerationAction = "ban"
	ModerationActionMute          ModerationAction = "mute"
	ModerationActionSuspend       ModerationAction = "suspend"
	ModerationActionDelete        ModerationAction = "delete"
	ModerationActionUpdate        ModerationAction = "update"
	ModerationActionHide          ModerationAction = "hide"
	ModerationActionTransferOwner ModerationAction = "transfer_owner"
)

type ModerationTargetType string

const (
	ModerationTargetUser      ModerationTargetType = "user"
	ModerationTargetPost      ModerationTargetType = "post"
	ModerationTargetComment   ModerationTargetType = "comment"
	ModerationTargetAd        ModerationTargetType = "ad"
	ModerationTargetReport    ModerationTargetType = "report"
	ModerationTargetCommunity ModerationTargetType = "community"
	ModerationTargetGroupChat ModerationTargetType = "group_chat"
)

type ModerationLog struct {
	ID          string               `json:"id"`
	ModeratorID string               `json:"moderator_id"`
	Action      ModerationAction     `json:"action"`
	TargetType  ModerationTargetType `json:"target_type"`
	TargetID    string               `json:"target_id"`
	Reason      string               `json:"reason"`
	CreatedAt   time.Time            `json:"created_at"`
}

func NewModerationLog(moderatorID string, action ModerationAction, targetType ModerationTargetType, targetID string, reason string) ModerationLog {
	return ModerationLog{ModeratorID: moderatorID, Action: action, TargetType: targetType, TargetID: targetID, Reason: reason}
}

func (a ModerationAction) String() string {
	return string(a)
}

func ParseModerationAction(value string) ModerationAction {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ModerationActionWarn):
		return ModerationActionWarn
	case string(ModerationActionBan):
		return ModerationActionBan
	case string(ModerationActionMute):
		return ModerationActionMute
	case string(ModerationActionSuspend):
		return ModerationActionSuspend
	case string(ModerationActionDelete):
		return ModerationActionDelete
	case string(ModerationActionUpdate):
		return ModerationActionUpdate
	case string(ModerationActionHide):
		return ModerationActionHide
	case string(ModerationActionTransferOwner):
		return ModerationActionTransferOwner
	default:
		return ModerationActionWarn
	}
}

func (t ModerationTargetType) String() string {
	return string(t)
}

func ParseModerationTargetType(value string) ModerationTargetType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ModerationTargetUser):
		return ModerationTargetUser
	case string(ModerationTargetPost):
		return ModerationTargetPost
	case string(ModerationTargetComment):
		return ModerationTargetComment
	case string(ModerationTargetAd):
		return ModerationTargetAd
	case string(ModerationTargetReport):
		return ModerationTargetReport
	case string(ModerationTargetCommunity):
		return ModerationTargetCommunity
	case string(ModerationTargetGroupChat):
		return ModerationTargetGroupChat
	default:
		return ModerationTargetUser
	}
}
