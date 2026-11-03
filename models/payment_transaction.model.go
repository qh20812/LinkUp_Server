package models

import "time"

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSuccess   PaymentStatus = "success"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// PaymentTransaction lưu một lần thanh toán gói quảng cáo qua cổng thanh toán
// (VNPay hoặc MoMo). TxnRef là mã đơn gửi sang cổng và là khoá idempotent khi
// nhận return/IPN; RequestID là mã request riêng của MoMo.
type PaymentTransaction struct {
	ID            string        `json:"id" gorm:"type:varchar(36);primaryKey"`
	TxnRef        string        `json:"txn_ref" gorm:"type:varchar(64);uniqueIndex;not null"`
	RequestID     string        `json:"request_id" gorm:"type:varchar(64);index;default:''"`
	UserID        string        `json:"user_id" gorm:"type:varchar(36);index;not null"`
	PackageID     string        `json:"package_id" gorm:"type:varchar(36);index;not null"`
	Amount        int64         `json:"amount" gorm:"not null"` // VND, chưa nhân 100
	Status        PaymentStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	Gateway       string        `json:"gateway" gorm:"type:varchar(20);not null;default:'vnpay'"`
	ResponseCode  string        `json:"response_code" gorm:"type:varchar(10);default:''"`
	TransactionNo string        `json:"transaction_no" gorm:"type:varchar(50);default:''"`
	BankCode      string        `json:"bank_code" gorm:"type:varchar(30);default:''"`
	PayDate       string        `json:"pay_date" gorm:"type:varchar(20);default:''"`
	OrderInfo     string        `json:"order_info" gorm:"type:varchar(255);default:''"`
	PaidAt        *time.Time    `json:"paid_at" gorm:"type:datetime(3);null"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`

	Package *AdPackage `json:"package,omitempty" gorm:"foreignKey:PackageID"`
}
