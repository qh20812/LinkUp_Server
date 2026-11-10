package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	errorsapp "linkup/errors"
	"linkup/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AdminTopupRow là một yêu cầu nạp kèm thông tin người dùng cho hàng đợi admin.
type AdminTopupRow struct {
	models.WalletTopup
	Username    string `gorm:"column:username"`
	Email       string `gorm:"column:email"`
	DisplayName string `gorm:"column:display_name"`
	AvatarURI   string `gorm:"column:avatar_uri"`
}

// AdminWalletRow là một dòng ví kèm thông tin hồ sơ cho trang quản trị ví.
type AdminWalletRow struct {
	models.Wallet
	Username    string `gorm:"column:username"`
	Email       string `gorm:"column:email"`
	DisplayName string `gorm:"column:display_name"`
	AvatarURI   string `gorm:"column:avatar_uri"`
	TotalTopup  int64  `gorm:"column:total_topup"`
}

// WalletRepository quản lý ví, sổ cái và yêu cầu nạp tiền.
type WalletRepository interface {
	GetOrCreate(ctx context.Context, userID string) (*models.Wallet, error)
	GetByUserID(ctx context.Context, userID string) (*models.Wallet, error)
	SumLedger(ctx context.Context, userID string, txType models.WalletTransactionType) (int64, error)
	ListTransactions(ctx context.Context, userID string, limit, offset int) ([]models.WalletTransaction, int64, error)

	CreateTopup(ctx context.Context, topup *models.WalletTopup) error
	FindTopupByID(ctx context.Context, id string) (*models.WalletTopup, error)
	CountOpenTopups(ctx context.Context, userID string) (int64, error)
	CountTopupsSince(ctx context.Context, userID string, since time.Time) (int64, error)
	ListTopupsByUser(ctx context.Context, userID string, limit, offset int) ([]models.WalletTopup, int64, error)
	MarkTopupReported(ctx context.Context, userID, id, proofURL string) error
	CancelTopup(ctx context.Context, userID, id string) error

	ApproveTopup(ctx context.Context, adminID, id string, confirmedAmount int64, note string) error
	RejectTopup(ctx context.Context, adminID, id, reason string) error
	ExpireTopups(ctx context.Context, now time.Time) (int64, error)

	AdminListTopups(ctx context.Context, status, keyword string, reportedOnly bool, limit, offset int) ([]AdminTopupRow, int64, error)
	AdminListWallets(ctx context.Context, keyword string, limit, offset int) ([]AdminWalletRow, int64, error)
	AdminWalletTotals(ctx context.Context) (balance, reserved, topup int64, err error)
}

type walletRepositoryImpl struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepositoryImpl{db: db}
}

func (r *walletRepositoryImpl) GetOrCreate(ctx context.Context, userID string) (*models.Wallet, error) {
	var wallet *models.Wallet
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		w, err := lockWalletTx(tx, userID)
		if err != nil {
			return err
		}
		wallet = w
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("lấy ví: %w", err)
	}
	return wallet, nil
}

func (r *walletRepositoryImpl) GetByUserID(ctx context.Context, userID string) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &wallet, nil
}

// SumLedger trả tổng amount (có dấu) của một loại bút toán. Ví dụ topup trả
// tổng tiền đã nạp, spend trả số âm nên caller đổi dấu để hiển thị.
func (r *walletRepositoryImpl) SumLedger(ctx context.Context, userID string, txType models.WalletTransactionType) (int64, error) {
	var total *int64
	err := r.db.WithContext(ctx).Model(&models.WalletTransaction{}).
		Where("user_id = ? AND type = ?", userID, txType).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	if total == nil {
		return 0, nil
	}
	return *total, nil
}

func (r *walletRepositoryImpl) ListTransactions(ctx context.Context, userID string, limit, offset int) ([]models.WalletTransaction, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.WalletTransaction{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []models.WalletTransaction
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Limit(limit).Offset(offset).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *walletRepositoryImpl) CreateTopup(ctx context.Context, topup *models.WalletTopup) error {
	return r.db.WithContext(ctx).Create(topup).Error
}

func (r *walletRepositoryImpl) FindTopupByID(ctx context.Context, id string) (*models.WalletTopup, error) {
	var topup models.WalletTopup
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&topup).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorsapp.New(errorsapp.ErrCodeWalletTopupNotFound)
		}
		return nil, err
	}
	return &topup, nil
}

func (r *walletRepositoryImpl) CountOpenTopups(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.WalletTopup{}).
		Where("user_id = ? AND status = ? AND expires_at > ?", userID, models.WalletTopupPending, time.Now()).
		Count(&count).Error
	return count, err
}

func (r *walletRepositoryImpl) CountTopupsSince(ctx context.Context, userID string, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.WalletTopup{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Count(&count).Error
	return count, err
}

func (r *walletRepositoryImpl) ListTopupsByUser(ctx context.Context, userID string, limit, offset int) ([]models.WalletTopup, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.WalletTopup{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []models.WalletTopup
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Limit(limit).Offset(offset).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// MarkTopupReported đánh dấu người dùng đã chuyển khoản (không cộng ví).
func (r *walletRepositoryImpl) MarkTopupReported(ctx context.Context, userID, id, proofURL string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.WalletTopup{}).
		Where("id = ? AND user_id = ? AND status = ? AND user_reported_at IS NULL",
			id, userID, models.WalletTopupPending).
		Updates(map[string]interface{}{
			"user_reported_at": now,
			"proof_url":        proofURL,
			"updated_at":       now,
		})
	if result.Error != nil {
		return result.Error
	}
	return r.explainTopupNoRows(ctx, id, userID, result.RowsAffected)
}

func (r *walletRepositoryImpl) CancelTopup(ctx context.Context, userID, id string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.WalletTopup{}).
		Where("id = ? AND user_id = ? AND status = ? AND user_reported_at IS NULL",
			id, userID, models.WalletTopupPending).
		Updates(map[string]interface{}{
			"status":     models.WalletTopupCanceled,
			"updated_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	return r.explainTopupNoRows(ctx, id, userID, result.RowsAffected)
}

// explainTopupNoRows biến "0 dòng bị ảnh hưởng" thành lỗi có nghĩa thay vì im
// lặng thất bại.
func (r *walletRepositoryImpl) explainTopupNoRows(ctx context.Context, id, userID string, affected int64) error {
	if affected > 0 {
		return nil
	}
	topup, err := r.FindTopupByID(ctx, id)
	if err != nil {
		return err
	}
	if topup.UserID != userID {
		return errorsapp.New(errorsapp.ErrCodeWalletNotOwner)
	}
	if topup.Status != models.WalletTopupPending {
		return errorsapp.New(errorsapp.ErrCodeWalletTopupNotPending)
	}
	return errorsapp.New(errorsapp.ErrCodeWalletTopupReported)
}

// ApproveTopup cộng ví theo số thực nhận và đóng yêu cầu nạp trong MỘT
// transaction. Khoá cả dòng topup và dòng ví nên hai admin bấm duyệt đồng thời
// chỉ cộng tiền một lần.
func (r *walletRepositoryImpl) ApproveTopup(ctx context.Context, adminID, id string, confirmedAmount int64, note string) error {
	if confirmedAmount <= 0 {
		return errorsapp.New(errorsapp.ErrCodeWalletInvalidAmount)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var topup models.WalletTopup
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", id).First(&topup).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorsapp.New(errorsapp.ErrCodeWalletTopupNotFound)
			}
			return err
		}
		if topup.Status != models.WalletTopupPending {
			return errorsapp.New(errorsapp.ErrCodeWalletTopupNotPending)
		}

		wallet, err := lockWalletTx(tx, topup.UserID)
		if err != nil {
			return err
		}
		if err := applyWalletDeltaTx(
			tx, wallet,
			confirmedAmount, 0, confirmedAmount,
			models.WalletTxTopup, "wallet_topup", topup.ID,
			"Duyệt nạp ví "+topup.Code,
		); err != nil {
			return err
		}

		now := time.Now()
		confirmed := confirmedAmount
		admin := adminID
		return tx.Model(&models.WalletTopup{}).Where("id = ?", topup.ID).
			Updates(map[string]interface{}{
				"status":           models.WalletTopupPaid,
				"confirmed_amount": confirmed,
				"confirmed_by":     admin,
				"admin_note":       note,
				"paid_at":          now,
				"updated_at":       now,
			}).Error
	})
}

func (r *walletRepositoryImpl) RejectTopup(ctx context.Context, adminID, id, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return errorsapp.New(errorsapp.ErrCodeWalletReasonRequired)
	}
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.WalletTopup{}).
		Where("id = ? AND status = ?", id, models.WalletTopupPending).
		Updates(map[string]interface{}{
			"status":       models.WalletTopupRejected,
			"confirmed_by": adminID,
			"admin_note":   reason,
			"updated_at":   now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := r.db.WithContext(ctx).Model(&models.WalletTopup{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errorsapp.New(errorsapp.ErrCodeWalletTopupNotFound)
		}
		return errorsapp.New(errorsapp.ErrCodeWalletTopupNotPending)
	}
	return nil
}

// ExpireTopups đóng các yêu cầu pending đã quá hạn mà người dùng CHƯA báo
// chuyển khoản. Yêu cầu đã báo chuyển thì để admin xử lý thủ công.
func (r *walletRepositoryImpl) ExpireTopups(ctx context.Context, now time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.WalletTopup{}).
		Where("status = ? AND expires_at < ? AND user_reported_at IS NULL",
			models.WalletTopupPending, now).
		Updates(map[string]interface{}{
			"status":     models.WalletTopupExpired,
			"updated_at": now,
		})
	return result.RowsAffected, result.Error
}

func (r *walletRepositoryImpl) AdminListTopups(
	ctx context.Context, status, keyword string, reportedOnly bool, limit, offset int,
) ([]AdminTopupRow, int64, error) {
	base := r.db.WithContext(ctx).Table("wallet_topups AS t")

	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" {
		base = base.Where("t.status = ?", status)
	}
	if reportedOnly {
		base = base.Where("t.user_reported_at IS NOT NULL")
	}
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where(
			"(t.code LIKE ? OR t.user_id = ? OR u.email LIKE ? OR u.username LIKE ? OR p.display_name LIKE ?)",
			like, keyword, like, like, like,
		)
	}

	var total int64
	if err := base.Joins("LEFT JOIN users u ON u.id = t.user_id").
		Joins("LEFT JOIN profiles p ON p.user_id = t.user_id").
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []AdminTopupRow
	if err := base.
		Select(`t.*, COALESCE(u.username, '') AS username, COALESCE(u.email, '') AS email,
			COALESCE(p.display_name, '') AS display_name, COALESCE(p.avatar_uri, '') AS avatar_uri`).
		Order("CASE WHEN t.user_reported_at IS NULL THEN 1 ELSE 0 END, t.created_at DESC").
		Limit(limit).Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *walletRepositoryImpl) AdminListWallets(ctx context.Context, keyword string, limit, offset int) ([]AdminWalletRow, int64, error) {
	base := r.db.WithContext(ctx).Table("wallets AS w").
		Joins("LEFT JOIN users u ON u.id = w.user_id").
		Joins("LEFT JOIN profiles p ON p.user_id = w.user_id")

	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where("(w.user_id = ? OR u.email LIKE ? OR u.username LIKE ?)", keyword, like, like)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []AdminWalletRow
	if err := base.
		Select(`w.*, COALESCE(u.username, '') AS username, COALESCE(u.email, '') AS email,
			COALESCE(p.display_name, '') AS display_name, COALESCE(p.avatar_uri, '') AS avatar_uri,
			COALESCE((SELECT SUM(wt.amount) FROM wallet_transactions wt
				WHERE wt.user_id = w.user_id AND wt.type = 'topup'), 0) AS total_topup`).
		Order("w.updated_at DESC").
		Limit(limit).Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *walletRepositoryImpl) AdminWalletTotals(ctx context.Context) (balance, reserved, topup int64, err error) {
	type walletTotals struct {
		Balance  int64 `gorm:"column:balance"`
		Reserved int64 `gorm:"column:reserved"`
	}
	type topupTotal struct {
		Topup int64 `gorm:"column:topup"`
	}

	var w walletTotals
	if err = r.db.WithContext(ctx).Table("wallets").
		Select("COALESCE(SUM(balance), 0) AS balance, COALESCE(SUM(reserved), 0) AS reserved").
		Scan(&w).Error; err != nil {
		return 0, 0, 0, err
	}

	var t topupTotal
	if err = r.db.WithContext(ctx).Table("wallet_transactions").
		Where("type = ?", models.WalletTxTopup).
		Select("COALESCE(SUM(amount), 0) AS topup").
		Scan(&t).Error; err != nil {
		return 0, 0, 0, err
	}
	return w.Balance, w.Reserved, t.Topup, nil
}
