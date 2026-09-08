package models

import "time"

type PushToken struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"index"`
	PushToken string    `json:"push_token" gorm:"uniqueIndex"`
	Platform  string    `json:"platform" gorm:"default:expo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
