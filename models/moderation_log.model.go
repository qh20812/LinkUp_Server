package models

import "time"

type ModerationLog struct {
	ID          int64     `json:"id" db:"id"`
	ModeratorID int64     `json:"moderator_id" db:"moderator_id"`
	Action      string    `json:"action" db:"action"`
	TargetType  string    `json:"target_type" db:"target_type"`
	TargetID    int64     `json:"target_id" db:"target_id"`
	Reason      string    `json:"reason" db:"reason"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

func NewModerationLog(moderatorID int64, action, targetType string, targetID int64, reason string) ModerationLog {
	return ModerationLog{ModeratorID: moderatorID, Action: action, TargetType: targetType, TargetID: targetID, Reason: reason}
}
