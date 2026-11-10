package models

import "time"

type AdPackage struct {
	ID                   string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name                 string    `json:"name" gorm:"size:100;not null"`
	Description          string    `json:"description" gorm:"type:text"`
	PriceMonthly         float64   `json:"price_monthly" gorm:"not null"`
	MaxSlots             int       `json:"max_slots" gorm:"not null"`
	MaxDurationDays      int       `json:"max_duration_days" gorm:"not null"`
	SupportsVideo        bool      `json:"supports_video" gorm:"default:false"`
	SupportsCarousel     bool      `json:"supports_carousel" gorm:"default:false"`
	HasAdvancedAnalytics bool      `json:"has_advanced_analytics" gorm:"default:false"`
	SortOrder            int       `json:"sort_order" gorm:"default:0"`
	// IsActive=false: ngừng bán nhưng vẫn giữ lịch sử subscription.
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
}
