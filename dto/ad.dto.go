package dto

import "time"

type CreateAdInput struct {
	Title           string     `json:"title"`
	Content         string     `json:"content"`
	Format          string     `json:"format"`
	TargetURL       string     `json:"target_url"`
	Budget          float64    `json:"budget"`
	DailyBudget     float64    `json:"daily_budget"`
	CPMPrice        float64    `json:"cpm_price"`
	CPCPrice        float64    `json:"cpc_price"`
	MaxImpressions  int        `json:"max_impressions"`
	StartedAt       *time.Time `json:"started_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	TargetGender    string     `json:"target_gender"`
	TargetAgeMin    int        `json:"target_age_min"`
	TargetAgeMax    int        `json:"target_age_max"`
	TargetLocations []string   `json:"target_locations"`
}

type UpdateAdStatusInput struct {
	Status string `json:"status" binding:"required,oneof=active paused completed pending rejected"`
}

type UpdateAdInput struct {
	Title           string     `json:"title"`
	Content         string     `json:"content"`
	TargetURL       string     `json:"target_url"`
	Budget          float64    `json:"budget"`
	DailyBudget     float64    `json:"daily_budget"`
	CPMPrice        float64    `json:"cpm_price"`
	CPCPrice        float64    `json:"cpc_price"`
	MaxImpressions  int        `json:"max_impressions"`
	StartedAt       *time.Time `json:"started_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	TargetGender    string     `json:"target_gender"`
	TargetAgeMin    int        `json:"target_age_min"`
	TargetAgeMax    int        `json:"target_age_max"`
	TargetLocations []string   `json:"target_locations"`
}

type AdListInput struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
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
	Interactions    int64   `json:"interactions"`

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

type PartnerAdListItem struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Status          string     `json:"status"`
	Budget          float64    `json:"budget"`
	TotalSpent      float64    `json:"total_spent"`
	Impressions     int64      `json:"impressions"`
	Clicks          int64      `json:"clicks"`
	CTR             float64    `json:"ctr"`
	MediaURI        string     `json:"media_uri"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
}

type AdOverviewResponse struct {
	TotalBudget      float64    `json:"total_budget"`
	TotalSpent       float64    `json:"total_spent"`
	RemainingBudget  float64    `json:"remaining_budget"`
	Impressions      int64      `json:"impressions"`
	Clicks           int64      `json:"clicks"`
	CTR              float64    `json:"ctr"`
	ActiveAds        int        `json:"active_ads"`
	SlotsUsed        int        `json:"slots_used"`
	MaxSlots         int        `json:"max_slots"`
	SubscriptionName string     `json:"subscription_name"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
}
