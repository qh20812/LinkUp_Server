package models

import "time"

type CommunityInviteCode struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	CommunityID string     `json:"community_id" gorm:"type:varchar(36);index;not null"`
	Code        string     `json:"code" gorm:"type:varchar(6);uniqueIndex;not null"`
	CreatedBy   string     `json:"created_by" gorm:"type:varchar(36);not null"`
	MaxUses     int        `json:"max_uses" gorm:"default:0"`
	UsedCount   int        `json:"used_count" gorm:"default:0"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at"`
}
