package models

import "time"

type BackgroundType string

const (
    BackgroundSolid    BackgroundType = "solid"
    BackgroundGradient BackgroundType = "gradient"
    BackgroundPreset   BackgroundType = "preset"
    BackgroundCustom   BackgroundType = "custom"
)

type ChatUserSettings struct {
    ChatID          string         `json:"chat_id" gorm:"primaryKey;column:chat_id"`
    UserID          string         `json:"user_id" gorm:"primaryKey;column:user_id"`
    BackgroundType  BackgroundType `json:"background_type" gorm:"column:background_type"`
    BackgroundValue string         `json:"background_value" gorm:"column:background_value;size:500"`
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
}

func (ChatUserSettings) TableName() string {
	return "chat_user_settings"
}

func ParseBackgroundType(s string) BackgroundType {
	switch BackgroundType(s) {
	case BackgroundSolid, BackgroundGradient, BackgroundPreset, BackgroundCustom:
		return BackgroundType(s)
	default:
		return ""
	}
}
