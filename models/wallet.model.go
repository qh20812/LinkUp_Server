package models

import "time"

// WalletTransactionType là loại bút toán trong sổ cái ví. Mỗi thay đổi số dư
// hoặc số tiền đang giữ chỗ đều phải có đúng một dòng tương ứng.
type WalletTransactionType string

const (
	// WalletTxTopup: nạp tiền vào ví (balance tăng).
	WalletTxTopup WalletTransactionType = "topup"
	// WalletTxReserve: giữ chỗ ngân sách cho một quảng cáo (reserved tăng).
	WalletTxReserve WalletTransactionType = "reserve"
	// WalletTxRelease: trả lại phần giữ chỗ chưa tiêu (reserved giảm).
	WalletTxRelease WalletTransactionType = "release"
	// WalletTxSpend: trừ tiền thật khi quảng cáo phát sinh impression/click
	// (balance và reserved cùng giảm).
	WalletTxSpend WalletTransactionType = "spend"
	// WalletTxAdjust: điều chỉnh thủ công bởi admin (có ghi chú bắt buộc).
	WalletTxAdjust WalletTransactionType = "adjust"
	// WalletTxPackagePurchase: thanh toán phí gói quảng cáo bằng ví.
	WalletTxPackagePurchase WalletTransactionType = "package_purchase"
)

// Wallet là ví trả trước của người dùng. Số dư lưu bằng số nguyên VND (không
// dùng float) để tránh sai số khi cộng/trừ nhiều lần.
//
// Bất biến: balance >= 0, reserved >= 0, available = balance - reserved >= 0.
type Wallet struct {
	ID        string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:varchar(36);uniqueIndex;not null"`
	Balance   int64     `json:"balance" gorm:"not null;default:0"`
	Reserved  int64     `json:"reserved" gorm:"not null;default:0"`
	Currency  string    `json:"currency" gorm:"size:3;not null;default:'VND'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Available là số tiền còn dùng được để giữ chỗ cho quảng cáo mới.
func (w Wallet) Available() int64 {
	available := w.Balance - w.Reserved
	if available < 0 {
		return 0
	}
	return available
}

// WalletTransaction là một dòng sổ cái bất biến. Amount là delta có dấu theo
// loại bút toán: topup dương (balance), spend âm (balance), reserve dương
// (reserved), release âm (reserved), adjust âm/dương (balance).
type WalletTransaction struct {
	ID            string                `json:"id" gorm:"type:varchar(36);primaryKey"`
	UserID        string                `json:"user_id" gorm:"type:varchar(36);index;not null"`
	Type          WalletTransactionType `json:"type" gorm:"size:20;not null"`
	Amount        int64                 `json:"amount" gorm:"not null"`
	BalanceAfter  int64                 `json:"balance_after" gorm:"not null"`
	ReservedAfter int64                 `json:"reserved_after" gorm:"not null"`
	RefType       string                `json:"ref_type" gorm:"size:30;not null;default:''"`
	RefID         string                `json:"ref_id" gorm:"type:varchar(36);not null;default:''"`
	Note          string                `json:"note" gorm:"type:varchar(255);not null;default:''"`
	CreatedAt     time.Time             `json:"created_at"`
}

// WalletTopupStatus là trạng thái của một yêu cầu nạp tiền qua chuyển khoản.
type WalletTopupStatus string

const (
	WalletTopupPending  WalletTopupStatus = "pending"
	WalletTopupPaid     WalletTopupStatus = "paid"
	WalletTopupRejected WalletTopupStatus = "rejected"
	WalletTopupExpired  WalletTopupStatus = "expired"
	WalletTopupCanceled WalletTopupStatus = "cancelled"
)

// WalletTopup là một yêu cầu nạp tiền. Code là nội dung chuyển khoản duy nhất
// (unique) để admin đối chiếu sao kê; amount là số tiền yêu cầu, còn
// ConfirmedAmount là số tiền thực nhận khi admin duyệt.
type WalletTopup struct {
	ID              string            `json:"id" gorm:"type:varchar(36);primaryKey"`
	UserID          string            `json:"user_id" gorm:"type:varchar(36);index;not null"`
	Code            string            `json:"code" gorm:"type:varchar(20);uniqueIndex;not null"`
	Amount          int64             `json:"amount" gorm:"not null"`
	Status          WalletTopupStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	ExpiresAt       time.Time         `json:"expires_at" gorm:"not null;index"`
	UserReportedAt  *time.Time        `json:"user_reported_at,omitempty"`
	ProofURL        string            `json:"proof_url" gorm:"type:varchar(500);not null;default:''"`
	ConfirmedAmount *int64            `json:"confirmed_amount,omitempty"`
	ConfirmedBy     *string           `json:"confirmed_by,omitempty" gorm:"type:varchar(36)"`
	AdminNote       string            `json:"admin_note" gorm:"type:varchar(500);not null;default:''"`
	PaidAt          *time.Time        `json:"paid_at,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// IsOpen báo yêu cầu còn chờ xử lý (chưa paid/rejected/expired/cancelled).
func (t WalletTopup) IsOpen() bool {
	return t.Status == WalletTopupPending
}
