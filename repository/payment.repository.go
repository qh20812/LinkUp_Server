package repository

import (
	"errors"
	"time"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"

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

// ErrWalletTxnExists báo một request khác cùng Idempotency-Key đã tạo giao dịch
// trước đó. Service dùng lỗi này để đọc lại kết quả thay vì báo lỗi DB.
var ErrWalletTxnExists = errors.New("wallet transaction already exists")

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
	// List trả giao dịch (kèm gói) cho partner hoặc admin, có lọc + phân trang.
	List(filter dto.PaymentListFilter, limit, offset int) ([]models.PaymentTransaction, int64, error)
	// PurchaseWithWallet trừ ví, huỷ các giao dịch gateway đang pending cùng gói,
	// ghi payment wallet thành công và kích hoạt subscription trong MỘT transaction.
	PurchaseWithWallet(
		userID string,
		pkg *models.AdPackage,
		activeSub, newSub *models.PartnerSubscription,
		txn *models.PaymentTransaction,
		amount int64,
	) (*models.Wallet, error)
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

// PurchaseWithWallet thực hiện toàn bộ việc mua gói bằng ví trong 1 transaction:
// khoá ví → kiểm tra available → trừ balance (reserved giữ nguyên) → ghi ledger →
// huỷ giao dịch MoMo/VNPay đang pending cùng gói → tạo payment gateway=wallet →
// đổi gói + nâng role. Lỗi ở bất kỳ bước nào đều rollback, không mất tiền.
func (r *paymentRepositoryImpl) PurchaseWithWallet(
	userID string,
	pkg *models.AdPackage,
	activeSub, newSub *models.PartnerSubscription,
	txn *models.PaymentTransaction,
	amount int64,
) (*models.Wallet, error) {
	var wallet *models.Wallet

	err := r.db.Transaction(func(tx *gorm.DB) error {
		w, err := lockWalletTx(tx, userID)
		if err != nil {
			return err
		}

		// Khoá ví đã tuần tự hoá các lần mua của cùng người dùng. Kiểm tra lại
		// txn_ref để hai request cùng Idempotency-Key không cùng trừ tiền.
		var duplicate models.PaymentTransaction
		dupErr := tx.Where("txn_ref = ?", txn.TxnRef).First(&duplicate).Error
		if dupErr == nil {
			return ErrWalletTxnExists
		}
		if !errors.Is(dupErr, gorm.ErrRecordNotFound) {
			return dupErr
		}

		// Đọc lại gói đang hoạt động trong transaction: request đồng thời có thể
		// đã kích hoạt đúng gói này sau khi service kiểm tra lần đầu.
		var current models.PartnerSubscription
		subErr := tx.Where("user_id = ? AND status = ? AND expires_at > ?",
			userID, models.SubscriptionStatusActive, time.Now()).
			First(&current).Error
		switch {
		case subErr == nil:
			if current.PackageID == pkg.ID {
				return errorsapp.New(errorsapp.ErrCodePaymentAlreadyActive)
			}
			activeSub = &current
		case !errors.Is(subErr, gorm.ErrRecordNotFound):
			return subErr
		}

		if amount > 0 {
			if w.Available() < amount {
				return errorsapp.Newf(errorsapp.ErrCodeWalletInsufficient, map[string]any{
					"required":  amount,
					"available": w.Available(),
				})
			}
			if err := applyWalletDeltaTx(
				tx, w,
				-amount, 0, -amount,
				models.WalletTxPackagePurchase, "package", txn.ID,
				"Thanh toán gói "+pkg.Name,
			); err != nil {
				return err
			}
		}

		now := time.Now()

		// Huỷ các giao dịch gateway đang chờ cùng gói để callback/IPN về sau
		// không kích hoạt gói lần hai.
		if err := tx.Model(&models.PaymentTransaction{}).
			Where("user_id = ? AND package_id = ? AND status = ? AND gateway IN ?",
				userID, pkg.ID, models.PaymentStatusPending, []string{"momo", "vnpay"}).
			Updates(map[string]any{
				"status":     models.PaymentStatusCancelled,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		txn.UserID = userID
		txn.PackageID = pkg.ID
		txn.Amount = amount
		txn.Status = models.PaymentStatusSuccess
		txn.Gateway = "wallet"
		txn.ResponseCode = "00"
		txn.PaidAt = &now
		txn.CreatedAt = now
		txn.UpdatedAt = now
		if err := tx.Create(txn).Error; err != nil {
			return err
		}

		if err := subscribeWithTx(tx, activeSub, newSub); err != nil {
			return err
		}

		wallet = w
		return nil
	})

	return wallet, err
}

func (r *paymentRepositoryImpl) List(filter dto.PaymentListFilter, limit, offset int) ([]models.PaymentTransaction, int64, error) {
	base := func() *gorm.DB {
		q := r.db.Model(&models.PaymentTransaction{})
		if filter.UserID != "" {
			q = q.Where("user_id = ?", filter.UserID)
		}
		if filter.Status != "" {
			q = q.Where("status = ?", filter.Status)
		}
		if filter.Gateway != "" {
			q = q.Where("gateway = ?", filter.Gateway)
		}
		return q
	}

	var total int64
	if err := base().Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]models.PaymentTransaction, 0)
	if err := base().Preload("Package").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
