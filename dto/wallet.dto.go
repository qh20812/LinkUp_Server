package dto

import (
	"time"

	"linkup/models"
)

// WalletResponse là số dư ví của người dùng đang đăng nhập.
type WalletResponse struct {
	Balance    int64  `json:"balance"`
	Reserved   int64  `json:"reserved"`
	Available  int64  `json:"available"`
	Currency   string `json:"currency"`
	TotalTopup int64  `json:"total_topup"`
	TotalSpent int64  `json:"total_spent"`
}

// BankInfo là thông tin tài khoản nhận tiền hiển thị trong trang nạp ví.
// Chỉ trả cho người dùng đã đăng nhập, không đưa ra API công khai.
type BankInfo struct {
	BankName    string `json:"bank_name"`
	BankBIN     string `json:"bank_bin"`
	AccountNo   string `json:"account_no"`
	AccountName string `json:"account_name"`
}

// CreateTopupInput là body tạo yêu cầu nạp tiền.
type CreateTopupInput struct {
	Amount int64 `json:"amount" binding:"required"`
}

// ReportTopupInput là body báo đã chuyển khoản (ảnh biên lai tuỳ chọn).
type ReportTopupInput struct {
	ProofURL string `json:"proof_url"`
}

// ApproveTopupInput cho phép admin cộng theo số thực nhận (khác số yêu cầu).
type ApproveTopupInput struct {
	ConfirmedAmount *int64 `json:"confirmed_amount"`
	Note            string `json:"note"`
}

// RejectTopupInput bắt buộc lý do từ chối.
type RejectTopupInput struct {
	Reason string `json:"reason"`
}

// TopupItem là một yêu cầu nạp tiền trong danh sách của người dùng.
type TopupItem struct {
	ID              string     `json:"id"`
	Code            string     `json:"code"`
	Amount          int64      `json:"amount"`
	Status          string     `json:"status"`
	ExpiresAt       time.Time  `json:"expires_at"`
	UserReportedAt  *time.Time `json:"user_reported_at,omitempty"`
	ProofURL        string     `json:"proof_url,omitempty"`
	ConfirmedAmount *int64     `json:"confirmed_amount,omitempty"`
	AdminNote       string     `json:"admin_note,omitempty"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// TopupCreatedResponse là kết quả tạo yêu cầu nạp: thông tin chuyển khoản +
// URL ảnh QR động (chỉ là ảnh, không xác nhận thanh toán).
type TopupCreatedResponse struct {
	ID               string    `json:"id"`
	Code             string    `json:"code"`
	Amount           int64     `json:"amount"`
	Status           string    `json:"status"`
	ExpiresAt        time.Time `json:"expires_at"`
	ExpiresInSeconds int64     `json:"expires_in_seconds"`
	QRURL            string    `json:"qr_url"`
	Bank             BankInfo  `json:"bank"`
	Memo             string    `json:"memo"`
}

// AdminTopupItem là yêu cầu nạp kèm thông tin người dùng cho hàng đợi admin.
type AdminTopupItem struct {
	TopupItem
	UserID      string `json:"user_id"`
	Username    string `json:"username,omitempty"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURI   string `json:"avatar_uri,omitempty"`
}

// TopupListResponse là danh sách yêu cầu nạp có phân trang.
type TopupListResponse struct {
	Data     []TopupItem `json:"data"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// AdminTopupListResponse là hàng đợi nạp tiền cho admin.
type AdminTopupListResponse struct {
	Data     []AdminTopupItem `json:"data"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// AdminTopupFilterInput là bộ lọc hàng đợi nạp tiền.
type AdminTopupFilterInput struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
	Reported bool   `form:"reported"`
}

// AdminWalletItem là một dòng trong danh sách ví của admin.
type AdminWalletItem struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username,omitempty"`
	Email       string    `json:"email,omitempty"`
	DisplayName string    `json:"display_name,omitempty"`
	AvatarURI   string    `json:"avatar_uri,omitempty"`
	Balance     int64     `json:"balance"`
	Reserved    int64     `json:"reserved"`
	Available   int64     `json:"available"`
	TotalTopup  int64     `json:"total_topup"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AdminWalletListResponse tổng hợp ví toàn hệ thống cho admin.
type AdminWalletListResponse struct {
	Data          []AdminWalletItem `json:"data"`
	Total         int64             `json:"total"`
	Page          int               `json:"page"`
	PageSize      int               `json:"page_size"`
	TotalBalance  int64             `json:"total_balance"`
	TotalReserved int64             `json:"total_reserved"`
	TotalTopupAll int64             `json:"total_topup_all"`
}

// WalletBalance là số dư tối thiểu trả kèm kết quả thanh toán bằng ví.
type WalletBalance struct {
	Balance   int64  `json:"balance"`
	Reserved  int64  `json:"reserved"`
	Available int64  `json:"available"`
	Currency  string `json:"currency"`
}

// WalletPurchaseResponse là kết quả mua gói bằng ví trả trước.
// Wallet chỉ có ở lần mua đầu; khi replay idempotent client giữ số dư cũ và
// tự refetch ví.
type WalletPurchaseResponse struct {
	Subscription SubscriptionResponse      `json:"subscription"`
	Payment      models.PaymentTransaction `json:"payment"`
	Wallet       *WalletBalance            `json:"wallet,omitempty"`
	Replayed     bool                      `json:"replayed"`
}

// WalletTransactionItem là một dòng lịch sử ví hiển thị cho người dùng.
type WalletTransactionItem struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	Amount        int64     `json:"amount"`
	BalanceAfter  int64     `json:"balance_after"`
	ReservedAfter int64     `json:"reserved_after"`
	RefType       string    `json:"ref_type,omitempty"`
	RefID         string    `json:"ref_id,omitempty"`
	Note          string    `json:"note,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// WalletTransactionListResponse là lịch sử sổ cái có phân trang.
type WalletTransactionListResponse struct {
	Data     []WalletTransactionItem `json:"data"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}
