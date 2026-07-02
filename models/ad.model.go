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
	ID        string     `json:"id"`
	PartnerID string     `json:"partner_id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	MediaID   *string    `json:"media_id,omitempty"`
	TargetURL string     `json:"target_url"`
	Status    AdStatus   `json:"status"`
	Budget    float64    `json:"budget"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
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
