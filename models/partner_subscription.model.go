package models

import "time"

type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
)

type PartnerSubscription struct {
	ID        string             `json:"id" gorm:"type:varchar(36);primaryKey"`
	UserID    string             `json:"user_id" gorm:"type:varchar(36);index;not null"`
	PackageID string             `json:"package_id" gorm:"type:varchar(36);index;not null"`
	SlotsUsed int                `json:"slots_used" gorm:"default:0"`
	StartedAt time.Time          `json:"started_at" gorm:"not null"`
	ExpiresAt time.Time          `json:"expires_at" gorm:"not null"`
	Status    SubscriptionStatus `json:"status" gorm:"default:'active'"`
	AutoRenew bool               `json:"auto_renew" gorm:"default:true"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`

	// Relational field
	Package *AdPackage `json:"package,omitempty" gorm:"foreignKey:PackageID"`
}
