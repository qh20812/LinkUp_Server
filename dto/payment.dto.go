package dto

import "time"

type CreatePaymentInput struct {
	PackageID string `json:"package_id" binding:"required"`
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
