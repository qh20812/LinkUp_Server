package models

import (
	"strings"
	"time"
)

type AdStatus string

const (
	AdStatusActive    AdStatus = "active"
	AdStatusPaused    AdStatus = "paused"
	AdStatusCompleted AdStatus = "completed"
)

type Ad struct {
	ID        string     `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	MediaID   *string `json:"media_id,omitempty" db:"media_id"`
	TargetURL string    `json:"target_url" db:"target_url"`
	Status    AdStatus  `json:"status" db:"status"`
	Budget    float64   `json:"budget" db:"budget"`
	StartedAt *time.Time `json:"started_at,omitempty" db:"started_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func NewAd(title, content string, targetURL string, budget float64) Ad {
	return Ad{
		Title:     title,
		Content:   content,
		TargetURL: targetURL,
		Status:    AdStatusActive,
		Budget:    budget,
	}
}

func (s AdStatus) String() string {
	return string(s)
}

func ParseAdStatus(value string) AdStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(AdStatusActive):
		return AdStatusActive
	case string(AdStatusPaused):
		return AdStatusPaused
	case string(AdStatusCompleted):
		return AdStatusCompleted
	default:
		return AdStatusActive
	}
}
