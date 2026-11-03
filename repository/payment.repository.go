package repository

import (
	"errors"
	"linkup/models"
	"time"

	"gorm.io/gorm"
)

// PaymentResult là dữ liệu cổng thanh toán trả về cho một giao dịch.
type PaymentResult struct {
	ResponseCode  string
	TransactionNo string
	BankCode      string
	PayDate       string
	PaidAt        time.Time
}

type PaymentRepository interface {
	Create(txn *models.PaymentTransaction) error
	FindByTxnRef(txnRef string) (*models.PaymentTransaction, error)
	// CompletePayment chuyển giao dịch pending → success và kích hoạt gói
	// trong cùng một transaction. Trả về false nếu giao dịch đã được xử lý
	// trước đó (IPN/return gửi lặp).
	CompletePayment(txnRef string, result PaymentResult, activeSub, newSub *models.PartnerSubscription) (bool, error)
	// MarkFailed chuyển giao dịch pending → failed. Không ghi đè giao dịch
	// đã success.
	MarkFailed(txnRef string, result PaymentResult) (bool, error)
}

type paymentRepositoryImpl struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepositoryImpl{db: db}
}

func (r *paymentRepositoryImpl) Create(txn *models.PaymentTransaction) error {
	return r.db.Create(txn).Error
}

func (r *paymentRepositoryImpl) FindByTxnRef(txnRef string) (*models.PaymentTransaction, error) {
	var txn models.PaymentTransaction
	err := r.db.Preload("Package").Where("txn_ref = ?", txnRef).First(&txn).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &txn, nil
}

func paymentUpdateFields(status models.PaymentStatus, result PaymentResult) map[string]any {
	fields := map[string]any{
		"status":         status,
		"response_code":  result.ResponseCode,
		"transaction_no": result.TransactionNo,
		"bank_code":      result.BankCode,
		"pay_date":       result.PayDate,
		"updated_at":     time.Now(),
	}
	if status == models.PaymentStatusSuccess {
		fields["paid_at"] = result.PaidAt
	}
	return fields
}

func (r *paymentRepositoryImpl) CompletePayment(
	txnRef string,
	result PaymentResult,
	activeSub, newSub *models.PartnerSubscription,
) (bool, error) {
	transitioned := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Conditional UPDATE: chỉ một callback thắng cuộc, các callback lặp
		// lại thấy RowsAffected = 0 và không tạo thêm subscription.
		res := tx.Model(&models.PaymentTransaction{}).
			Where("txn_ref = ? AND status = ?", txnRef, models.PaymentStatusPending).
			Updates(paymentUpdateFields(models.PaymentStatusSuccess, result))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		transitioned = true
		return subscribeWithTx(tx, activeSub, newSub)
	})
	return transitioned, err
}

func (r *paymentRepositoryImpl) MarkFailed(txnRef string, result PaymentResult) (bool, error) {
	res := r.db.Model(&models.PaymentTransaction{}).
		Where("txn_ref = ? AND status = ?", txnRef, models.PaymentStatusPending).
		Updates(paymentUpdateFields(models.PaymentStatusFailed, result))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
