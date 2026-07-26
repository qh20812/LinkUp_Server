package dto

import "time"

type SubscribePackageInput struct {
	PackageID string `json:"package_id" binding:"required"`
}

type SubscriptionResponse struct {
	ID           string    `json:"id"`
	PackageName  string    `json:"package_name"`
	MaxSlots     int       `json:"max_slots"`
	SlotsUsed    int       `json:"slots_used"`
	SlotsLeft    int       `json:"slots_left"`
	PriceMonthly float64   `json:"price_monthly"`
	StartedAt    time.Time `json:"started_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Status       string    `json:"status"`
}
