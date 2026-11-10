package repository

import (
	"errors"
	"time"

	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Các helper dưới đây thao tác trên ví trong một transaction đã mở. Chúng được
// dùng chung bởi WalletRepository và AdRepository để việc giữ chỗ/trừ tiền
// ngân sách quảng cáo nằm cùng transaction với bản ghi ad/analytics.

// lockWalletTx khoá dòng ví của user (SELECT ... FOR UPDATE), tạo mới nếu chưa
// có. Tạo mới có thể race với request khác: khi INSERT trùng unique user_id thì
// đọc lại dòng vừa được tạo và khoá nó.
func lockWalletTx(tx *gorm.DB, userID string) (*models.Wallet, error) {
	var wallet models.Wallet
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&wallet).Error
	if err == nil {
		return &wallet, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	now := time.Now()
	wallet = models.Wallet{
		ID:        utils.GenerateUUID(),
		UserID:    userID,
		Currency:  "VND",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if createErr := tx.Create(&wallet).Error; createErr != nil {
		var existing models.Wallet
		if readErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&existing).Error; readErr == nil {
			return &existing, nil
		}
		return nil, createErr
	}
	return &wallet, nil
}

// applyWalletDeltaTx áp một bút toán lên ví và ghi sổ cái trong cùng transaction.
// amount là delta có dấu của bút toán (dùng để hiển thị lịch sử), còn
// deltaBalance/deltaReserved quyết định số dư thực tế.
//
// Từ chối nếu kết quả vi phạm bất biến: balance >= 0, reserved >= 0 và
// available = balance - reserved >= 0.
func applyWalletDeltaTx(
	tx *gorm.DB,
	wallet *models.Wallet,
	deltaBalance, deltaReserved, amount int64,
	txType models.WalletTransactionType,
	refType, refID, note string,
) error {
	if deltaBalance == 0 && deltaReserved == 0 {
		return nil
	}

	newBalance := wallet.Balance + deltaBalance
	newReserved := wallet.Reserved + deltaReserved
	if newBalance < 0 || newReserved < 0 || newBalance < newReserved {
		return errorsapp.New(errorsapp.ErrCodeWalletInvalidLedger)
	}

	now := time.Now()
	if err := tx.Model(&models.Wallet{}).
		Where("id = ?", wallet.ID).
		Updates(map[string]interface{}{
			"balance":    newBalance,
			"reserved":   newReserved,
			"updated_at": now,
		}).Error; err != nil {
		return err
	}

	ledger := models.WalletTransaction{
		ID:            utils.GenerateUUID(),
		UserID:        wallet.UserID,
		Type:          txType,
		Amount:        amount,
		BalanceAfter:  newBalance,
		ReservedAfter: newReserved,
		RefType:       refType,
		RefID:         refID,
		Note:          note,
		CreatedAt:     now,
	}
	if err := tx.Create(&ledger).Error; err != nil {
		return err
	}

	wallet.Balance = newBalance
	wallet.Reserved = newReserved
	wallet.UpdatedAt = now
	return nil
}
