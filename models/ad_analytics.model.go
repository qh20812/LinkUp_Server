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
	ID         string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	AdID       string    `json:"ad_id" gorm:"type:varchar(36);index"`
	UserID     *string   `json:"user_id,omitempty" gorm:"type:varchar(36);index"`
	ActionType string    `json:"action_type"`
	IPAddress  string    `json:"ip_address"`
	// Cost là chi phí của riêng hành động này (CPM/1000 hoặc CPC). Cột đã có
	// trong DB từ trước; bổ sung vào model để API/analytics đọc được.
	Cost      float64   `json:"cost" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAdAnalytics(adID string, userID *string, actionType string, ipAddress string) AdAnalytics {
	return AdAnalytics{
		AdID:       adID,
		UserID:     userID,
		ActionType: actionType,
		IPAddress:  ipAddress,
	}
}
