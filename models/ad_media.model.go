package models

import "time"

type AdMedia struct {
	ID        string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	AdID      string    `json:"ad_id" gorm:"type:varchar(36);index;not null"`
	URL       string    `json:"url" gorm:"type:text;not null"`
	MediaType string    `json:"media_type" gorm:"size:20;default:'image'"` // image | video
	SortOrder int       `json:"sort_order" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
}
