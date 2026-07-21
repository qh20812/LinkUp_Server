package models

import "time"

type AdPackage struct {
	ID                   string    `json:"id" gorm:"primaryKey"`
	Name                 string    `json:"name" gorm:"size:100;not null"`
	Description          string    `json:"description" gorm:"type:text"`
	PriceMonthly         float64   `json:"price_monthly" gorm:"not null"`
	MaxSlots             int       `json:"max_slots" gorm:"not null"`
	MaxDurationDays      int       `json:"max_duration_days" gorm:"not null"`
	SupportsVideo        bool      `json:"supports_video" gorm:"default:false"`
	SupportsCarousel     bool      `json:"supports_carousel" gorm:"default:false"`
	HasAdvancedAnalytics bool      `json:"has_advanced_analytics" gorm:"default:false"`
	SortOrder            int       `json:"sort_order" gorm:"default:0"`
	CreatedAt            time.Time `json:"created_at"`
}
