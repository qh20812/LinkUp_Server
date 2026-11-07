package models

import "time"

// AdminAuditLog ghi lại hành động quản trị trên tài nguyên (hiện dùng cho
// quảng cáo/gói) để truy vết khiếu nại hoặc tranh chấp.
type AdminAuditLog struct {
	ID         string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	AdminID    string    `json:"admin_id" gorm:"type:varchar(36);index;not null"`
	Action     string    `json:"action" gorm:"size:50;not null"`
	TargetType string    `json:"target_type" gorm:"size:30;not null"`
	TargetID   string    `json:"target_id" gorm:"type:varchar(36);not null"`
	Reason     string    `json:"reason" gorm:"type:text"`
	Metadata   string    `json:"metadata" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at"`
}
