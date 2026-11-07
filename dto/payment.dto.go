package dto

import "time"

type CreatePaymentInput struct {
	PackageID string `json:"package_id" binding:"required"`
}

// PaymentListFilter dùng cho partner (bị ép user_id) và admin (lọc tùy chọn).
type PaymentListFilter struct {
	UserID   string `form:"user_id"`
	Status   string `form:"status"`
	Gateway  string `form:"gateway"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

// PackageInput dùng cho admin tạo/sửa gói quảng cáo.
type PackageInput struct {
	Name                 string  `json:"name"`
	Description          string  `json:"description"`
	PriceMonthly         float64 `json:"price_monthly"`
	MaxSlots             int     `json:"max_slots"`
	MaxDurationDays      int     `json:"max_duration_days"`
	SupportsVideo        bool    `json:"supports_video"`
	SupportsCarousel     bool    `json:"supports_carousel"`
	HasAdvancedAnalytics bool    `json:"has_advanced_analytics"`
	SortOrder            int     `json:"sort_order"`
	IsActive             *bool   `json:"is_active"`
}

type CreatePaymentResponse struct {
	PaymentURL string    `json:"payment_url"`
	TxnRef     string    `json:"txn_ref"`
	Amount     int64     `json:"amount"` // VND
	ExpiresAt  time.Time `json:"expires_at"`
}

// PaymentCallbackResult là kết quả xử lý return/IPN sau khi đã kiểm tra chữ ký,
// mã merchant, giao dịch và số tiền.
type PaymentCallbackResult struct {
	Valid            bool   `json:"valid"`
	Internal         bool   `json:"internal"`
	NotFound         bool   `json:"not_found"`
	AmountMismatch   bool   `json:"amount_mismatch"`
	Success          bool   `json:"success"`
	AlreadyConfirmed bool   `json:"already_confirmed"`
	TxnRef           string `json:"txn_ref"`
	ResponseCode     string `json:"response_code"`
	Message          string `json:"message"`
}
