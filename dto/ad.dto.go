package dto

import "time"

type CreateAdInput struct {
	Title     string     `json:"title" binding:"required,min=5,max=100"`
	Content   string     `json:"content" binding:"required,min=10"`
	MediaID   *string    `json:"media_id"`
	TargetURL string     `json:"target_url" binding:"required,url"`
	Budget    float64    `json:"budget" binding:"required,gt=0"`
	StartedAt *time.Time `json:"started_at" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at" binding:"required,gtfield=StartedAt"`
}

type UpdateAdStatusInput struct {
	Status string `json:"status" binding:"required,oneof=active paused completed"`
}

type AdPerformanceResponse struct {
	AdID         string  `json:"ad_id"`
	Title        string  `json:"title"`
	Status       string  `json:"status"`
	Budget       float64 `json:"budget"`
	Impressions  int64   `json:"impressions"`
	Clicks       int64   `json:"clicks"`
	Interactions int64   `json:"interactions"`
	CTR          float64 `json:"ctr"`
}

type TrackActionInput struct {
	ActionType string `json:"action_type" binding:"required,oneof=impression click interact"`
}
