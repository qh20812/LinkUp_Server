package dto

import "time"

type CreateAdInput struct {
	Title          string     `json:"title"`
	Content        string     `json:"content"`
	Format         string     `json:"format"` // image | video | carousel
	TargetURL      string     `json:"target_url"`
	Budget         float64    `json:"budget"`
	DailyBudget    float64    `json:"daily_budget"`
	CPMPrice       float64    `json:"cpm_price"`
	CPCPrice       float64    `json:"cpc_price"`
	MaxImpressions int        `json:"max_impressions"`
	StartedAt      *time.Time `json:"started_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
}

type UpdateAdStatusInput struct {
	Status string `json:"status" binding:"required,oneof=active paused completed"`
}

type AdPerformanceResponse struct {
	AdID            string  `json:"ad_id"`
	Title           string  `json:"title"`
	Status          string  `json:"status"`
	Format          string  `json:"format"`
	Budget          float64 `json:"budget"`
	DailyBudget     float64 `json:"daily_budget"`
	TotalSpent      float64 `json:"total_spent"`
	RemainingBudget float64 `json:"remaining_budget"`
	Impressions     int64   `json:"impressions"`
	UniqueReach     int64   `json:"unique_reach"`
	Clicks          int64   `json:"clicks"`

	CTR float64 `json:"click_through_rate"`
	CPC float64 `json:"cost_per_click"`
	CPM float64 `json:"cost_per_thousand_impressions"` // (Mille = 1000 lượt hiển thị)

	VideoStarts      int64      `json:"video_starts,omitempty"`
	VideoCompletions int64      `json:"video_completions,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
}

type TrackActionInput struct {
	ActionType string `json:"action_type" binding:"required,oneof=impression view click swipe video_start video_end"`
}
