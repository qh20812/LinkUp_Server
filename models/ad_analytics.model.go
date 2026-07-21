package models

import "time"

type AdActionType string

const (
	ActionImpression AdActionType = "impression"  // Quảng cáo hiển thị
	ActionView       AdActionType = "view"        // Đã xem (Viewable impression)
	ActionClick      AdActionType = "click"       // Click quảng cáo
	ActionSwipe      AdActionType = "swipe"       // Vuốt slide carousel
	ActionVideoStart AdActionType = "video_start" // Bắt đầu xem video
	ActionVideoEnd   AdActionType = "video_end"   // Xem hết video
)

type AdAnalytics struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	AdID       string    `json:"ad_id" gorm:"index"`
	UserID     *string   `json:"user_id,omitempty" gorm:"index"`
	ActionType string    `json:"action_type"`
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewAdAnalytics(adID string, userID *string, actionType string, ipAddress string) AdAnalytics {
	return AdAnalytics{
		AdID:       adID,
		UserID:     userID,
		ActionType: actionType,
		IPAddress:  ipAddress,
	}
}
