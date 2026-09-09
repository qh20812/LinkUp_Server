package models

import "time"

type SystemConfig struct {
	Key       string    `gorm:"primaryKey;size:100"`
	Value     string    `gorm:"type:text;not null"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}