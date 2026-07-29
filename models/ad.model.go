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

type AdFormat string

const (
	AdFormatImage    AdFormat = "image"
	AdFormatVideo    AdFormat = "video"
	AdFormatCarousel AdFormat = "carousel"
)

type Ad struct {
	ID             string     `json:"id" gorm:"type:varchar(36);primaryKey"`
	PartnerID      string     `json:"partner_id" gorm:"type:varchar(36);index;not null"`
	PackageID      *string    `json:"package_id,omitempty" gorm:"type:varchar(36);index"`
	Title          string     `json:"title" gorm:"size:100;not null"`
	Content        string     `json:"content" gorm:"type:text;not null"`
	Format         AdFormat   `json:"format" gorm:"size:20;default:'image'"`
	TargetURL      string     `json:"target_url" gorm:"type:text;not null"`
	Status         AdStatus   `json:"status" gorm:"size:20;default:'active'"`
	Budget         float64    `json:"budget" gorm:"not null"`
	DailyBudget    float64    `json:"daily_budget" gorm:"default:0"`
	TotalSpent     float64    `json:"total_spent" gorm:"default:0"`
	CPMPrice       float64    `json:"cpm_price" gorm:"default:0"` // Đơn giá CPM tham chiếu từ gói
	CPCPrice       float64    `json:"cpc_price" gorm:"default:0"` // Đơn giá CPC tham chiếu từ gói
	MaxImpressions int        `json:"max_impressions" gorm:"default:0"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`

	// Relational fields
	MediaList []AdMedia  `json:"media_list,omitempty" gorm:"foreignKey:AdID;constraint:OnDelete:CASCADE"`
	Package   *AdPackage `json:"package,omitempty" gorm:"foreignKey:PackageID"`
}

func NewAd(title, content, targetURL string, budget float64, format AdFormat) Ad {
	return Ad{
		Title:     title,
		Content:   content,
		TargetURL: targetURL,
		Format:    format,
		Status:    AdStatusActive,
		Budget:    budget,
	}
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
