package services

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"linkup/config"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
)

const (
	// topupCodeAlphabet bỏ các ký tự dễ nhầm khi đọc/ghi tay (I, O, 0, 1).
	topupCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	topupCodeLength   = 6
	topupCodePrefix   = "LK"
	topupCodeAttempts = 5

	defaultTopupExpiresMinutes = 30
	defaultTopupMinAmount      = int64(10000)
	defaultTopupMaxAmount      = int64(50000000)
	defaultTopupDailyLimit     = 10
	defaultTopupPendingLimit   = 5
)

// WalletService là nghiệp vụ ví trả trước + nạp tiền bằng chuyển khoản.
type WalletService interface {
	GetWallet(ctx context.Context, userID string) (*dto.WalletResponse, error)
	ListTransactions(ctx context.Context, userID string, page, pageSize int) (*dto.WalletTransactionListResponse, error)
	ListTopups(ctx context.Context, userID string, page, pageSize int) (*dto.TopupListResponse, error)
	CreateTopup(ctx context.Context, userID string, amount int64) (*dto.TopupCreatedResponse, error)
	ReportTopup(ctx context.Context, userID, topupID, proofURL string) (*dto.TopupItem, error)
	CancelTopup(ctx context.Context, userID, topupID string) error

	AdminListTopups(ctx context.Context, filter dto.AdminTopupFilterInput) (*dto.AdminTopupListResponse, error)
	ApproveTopup(ctx context.Context, adminID, topupID string, input dto.ApproveTopupInput) (*dto.TopupItem, error)
	RejectTopup(ctx context.Context, adminID, topupID string, input dto.RejectTopupInput) (*dto.TopupItem, error)
	AdminListWallets(ctx context.Context, keyword string, page, pageSize int) (*dto.AdminWalletListResponse, error)

	ExpireTopups(ctx context.Context) (int, error)
}

type walletServiceImpl struct {
	repo                repository.WalletRepository
	auditRepo           *repository.AdminAuditRepository
	notificationService *NotificationService
	env                 config.Env
}

func NewWalletService(repo repository.WalletRepository, env config.Env) *walletServiceImpl {
	return &walletServiceImpl{repo: repo, env: env}
}

// SetAuditRepository ghi nhật ký hành động admin khi duyệt/từ chối nạp tiền.
func (s *walletServiceImpl) SetAuditRepository(repo *repository.AdminAuditRepository) {
	s.auditRepo = repo
}

// SetNotificationService gửi thông báo khi ví được cộng hoặc yêu cầu bị từ chối.
func (s *walletServiceImpl) SetNotificationService(service *NotificationService) {
	s.notificationService = service
}

// walletSettings chuẩn hoá cấu hình ví: thiếu/không hợp lệ thì dùng mặc định để
// một .env cũ không làm hỏng luồng nạp tiền.
type walletSettings struct {
	bankBIN      string
	accountNo    string
	accountName  string
	bankName     string
	qrBaseURL    string
	expiresMin   int
	minAmount    int64
	maxAmount    int64
	dailyLimit   int
	pendingLimit int
}

func (s *walletServiceImpl) settings() walletSettings {
	set := walletSettings{
		bankBIN:      strings.TrimSpace(s.env.BankBIN),
		accountNo:    strings.TrimSpace(s.env.BankAccountNo),
		accountName:  strings.TrimSpace(s.env.BankAccountName),
		bankName:     strings.TrimSpace(s.env.BankShortName),
		qrBaseURL:    strings.TrimSpace(s.env.TopupQRBaseURL),
		expiresMin:   s.env.TopupExpiresMin,
		minAmount:    s.env.TopupMinAmount,
		maxAmount:    s.env.TopupMaxAmount,
		dailyLimit:   s.env.TopupDailyLimit,
		pendingLimit: s.env.TopupPendingLimit,
	}
	if set.qrBaseURL == "" {
		set.qrBaseURL = "https://img.vietqr.io/image"
	}
	if set.expiresMin <= 0 {
		set.expiresMin = defaultTopupExpiresMinutes
	}
	if set.minAmount <= 0 {
		set.minAmount = defaultTopupMinAmount
	}
	if set.maxAmount <= 0 || set.maxAmount < set.minAmount {
		set.maxAmount = defaultTopupMaxAmount
	}
	if set.dailyLimit <= 0 {
		set.dailyLimit = defaultTopupDailyLimit
	}
	if set.pendingLimit <= 0 {
		set.pendingLimit = defaultTopupPendingLimit
	}
	return set
}

// bankConfigured báo đã đủ thông tin để tạo QR chưa. Thiếu cấu hình thì endpoint
// tạo yêu cầu nạp trả lỗi rõ ràng thay vì sinh QR sai.
func (set walletSettings) bankConfigured() bool {
	return set.bankBIN != "" && set.accountNo != "" && set.accountName != ""
}

func (set walletSettings) bankInfo() dto.BankInfo {
	return dto.BankInfo{
		BankName:    set.bankName,
		BankBIN:     set.bankBIN,
		AccountNo:   set.accountNo,
		AccountName: set.accountName,
	}
}

// GenerateTopupCode sinh mã nội dung chuyển khoản dạng LK + 6 ký tự ASCII
// uppercase, không dấu, không khoảng trắng. Bảng chữ 32 ký tự chia hết 256 nên
// lấy byte % 32 không lệch phân phối.
func GenerateTopupCode() (string, error) {
	buf := make([]byte, topupCodeLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("sinh mã nạp tiền: %w", err)
	}
	var b strings.Builder
	b.WriteString(topupCodePrefix)
	for _, v := range buf {
		b.WriteByte(topupCodeAlphabet[int(v)%len(topupCodeAlphabet)])
	}
	return b.String(), nil
}

// ValidateTopupAmount kiểm tra số tiền nạp có nằm trong [min, max] và là số
// nguyên dương. Tách riêng để test không cần DB.
func ValidateTopupAmount(amount, min, max int64) error {
	if amount <= 0 {
		return errorsapp.New(errorsapp.ErrCodeWalletInvalidAmount)
	}
	if min > 0 && amount < min {
		return errorsapp.Newf(errorsapp.ErrCodeWalletAmountTooSmall, map[string]any{
			"min": formatVND(min),
		})
	}
	if max > 0 && amount > max {
		return errorsapp.Newf(errorsapp.ErrCodeWalletAmountTooLarge, map[string]any{
			"max": formatVND(max),
		})
	}
	return nil
}

// BuildVietQRURL dựng URL ảnh VietQR động. Đây chỉ là ảnh chứa thông tin
// chuyển khoản (BIN + số tài khoản + số tiền + nội dung), KHÔNG kết nối ngân
// hàng và không xác nhận thanh toán.
func BuildVietQRURL(baseURL, bankBIN, accountNo, accountName string, amount int64, memo string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "https://img.vietqr.io/image"
	}
	bankBIN = strings.TrimSpace(bankBIN)
	accountNo = strings.TrimSpace(accountNo)

	parts := []string{"amount=" + fmt.Sprintf("%d", amount)}
	if memo = strings.TrimSpace(memo); memo != "" {
		parts = append(parts, "addInfo="+url.QueryEscape(memo))
	}
	if name := strings.TrimSpace(accountName); name != "" {
		// VietQR hiển thị tên có dấu; dùng %20 thay vì dấu + cho an toàn.
		parts = append(parts, "accountName="+strings.ReplaceAll(url.QueryEscape(name), "+", "%20"))
	}

	return fmt.Sprintf("%s/%s-%s-compact2.png?%s", base, bankBIN, accountNo, strings.Join(parts, "&"))
}

func (s *walletServiceImpl) GetWallet(ctx context.Context, userID string) (*dto.WalletResponse, error) {
	wallet, err := s.repo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}
	topup, err := s.repo.SumLedger(ctx, userID, models.WalletTxTopup)
	if err != nil {
		return nil, err
	}
	spent, err := s.repo.SumLedger(ctx, userID, models.WalletTxSpend)
	if err != nil {
		return nil, err
	}
	if spent < 0 {
		spent = -spent
	}
	return &dto.WalletResponse{
		Balance:    wallet.Balance,
		Reserved:   wallet.Reserved,
		Available:  wallet.Available(),
		Currency:   wallet.Currency,
		TotalTopup: topup,
		TotalSpent: spent,
	}, nil
}

func (s *walletServiceImpl) ListTransactions(ctx context.Context, userID string, page, pageSize int) (*dto.WalletTransactionListResponse, error) {
	page, pageSize = normalizePage(page, pageSize)
	rows, total, err := s.repo.ListTransactions(ctx, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]dto.WalletTransactionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.WalletTransactionItem{
			ID:            row.ID,
			Type:          string(row.Type),
			Amount:        row.Amount,
			BalanceAfter:  row.BalanceAfter,
			ReservedAfter: row.ReservedAfter,
			RefType:       row.RefType,
			RefID:         row.RefID,
			Note:          row.Note,
			CreatedAt:     row.CreatedAt,
		})
	}
	return &dto.WalletTransactionListResponse{
		Data:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *walletServiceImpl) ListTopups(ctx context.Context, userID string, page, pageSize int) (*dto.TopupListResponse, error) {
	page, pageSize = normalizePage(page, pageSize)
	rows, total, err := s.repo.ListTopupsByUser(ctx, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]dto.TopupItem, 0, len(rows))
	for i := range rows {
		items = append(items, toTopupItem(&rows[i]))
	}
	return &dto.TopupListResponse{
		Data:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *walletServiceImpl) CreateTopup(ctx context.Context, userID string, amount int64) (*dto.TopupCreatedResponse, error) {
	set := s.settings()
	if !set.bankConfigured() {
		return nil, errorsapp.New(errorsapp.ErrCodeWalletNotConfigured)
	}
	if err := ValidateTopupAmount(amount, set.minAmount, set.maxAmount); err != nil {
		return nil, err
	}

	open, err := s.repo.CountOpenTopups(ctx, userID)
	if err != nil {
		return nil, err
	}
	if open >= int64(set.pendingLimit) {
		return nil, errorsapp.New(errorsapp.ErrCodeWalletTopupRateLimited)
	}
	startOfToday := time.Now()
	since := time.Date(startOfToday.Year(), startOfToday.Month(), startOfToday.Day(), 0, 0, 0, 0, startOfToday.Location())
	todayCount, err := s.repo.CountTopupsSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if todayCount >= int64(set.dailyLimit) {
		return nil, errorsapp.New(errorsapp.ErrCodeWalletTopupRateLimited)
	}

	now := time.Now()
	topup := &models.WalletTopup{
		UserID:    userID,
		Amount:    amount,
		Status:    models.WalletTopupPending,
		ExpiresAt: now.Add(time.Duration(set.expiresMin) * time.Minute),
		CreatedAt: now,
		UpdatedAt: now,
	}

	var createErr error
	for attempt := 0; attempt < topupCodeAttempts; attempt++ {
		code, err := GenerateTopupCode()
		if err != nil {
			return nil, errorsapp.New(errorsapp.ErrCodeWalletTopupCodeFailed)
		}
		topup.ID = utils.GenerateUUID()
		topup.Code = code
		createErr = s.repo.CreateTopup(ctx, topup)
		if createErr == nil {
			break
		}
		if !isDuplicateKeyError(createErr) {
			return nil, createErr
		}
	}
	if createErr != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeWalletTopupCodeFailed)
	}

	qrURL := BuildVietQRURL(set.qrBaseURL, set.bankBIN, set.accountNo, set.accountName, amount, topup.Code)
	return &dto.TopupCreatedResponse{
		ID:               topup.ID,
		Code:             topup.Code,
		Amount:           topup.Amount,
		Status:           string(topup.Status),
		ExpiresAt:        topup.ExpiresAt,
		ExpiresInSeconds: int64(set.expiresMin) * 60,
		QRURL:            qrURL,
		Bank:             set.bankInfo(),
		Memo:             topup.Code,
	}, nil
}

func (s *walletServiceImpl) ReportTopup(ctx context.Context, userID, topupID, proofURL string) (*dto.TopupItem, error) {
	topup, err := s.repo.FindTopupByID(ctx, topupID)
	if err != nil {
		return nil, err
	}
	if topup.UserID != userID {
		return nil, errorsapp.New(errorsapp.ErrCodeWalletNotOwner)
	}
	if topup.Status != models.WalletTopupPending {
		return nil, errorsapp.New(errorsapp.ErrCodeWalletTopupNotPending)
	}
	if topup.UserReportedAt != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeWalletTopupReported)
	}
	if !topup.ExpiresAt.IsZero() && topup.ExpiresAt.Before(time.Now()) {
		return nil, errorsapp.New(errorsapp.ErrCodeWalletTopupExpired)
	}

	proofURL = strings.TrimSpace(proofURL)
	if len(proofURL) > 500 {
		return nil, errorsapp.New(errorsapp.ErrCodeInvalidInput)
	}
	// Ảnh biên lai chỉ để tham khảo nhưng vẫn phải là URL http(s) để admin
	// không mở nhầm scheme lạ (javascript:, data:...).
	if proofURL != "" && !strings.HasPrefix(proofURL, "http://") && !strings.HasPrefix(proofURL, "https://") {
		return nil, errorsapp.New(errorsapp.ErrCodeInvalidInput)
	}
	if err := s.repo.MarkTopupReported(ctx, userID, topupID, proofURL); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindTopupByID(ctx, topupID)
	if err != nil {
		return nil, err
	}
	item := toTopupItem(updated)
	return &item, nil
}

func (s *walletServiceImpl) CancelTopup(ctx context.Context, userID, topupID string) error {
	topup, err := s.repo.FindTopupByID(ctx, topupID)
	if err != nil {
		return err
	}
	if topup.UserID != userID {
		return errorsapp.New(errorsapp.ErrCodeWalletNotOwner)
	}
	return s.repo.CancelTopup(ctx, userID, topupID)
}

func (s *walletServiceImpl) AdminListTopups(ctx context.Context, filter dto.AdminTopupFilterInput) (*dto.AdminTopupListResponse, error) {
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	rows, total, err := s.repo.AdminListTopups(
		ctx,
		filter.Status,
		filter.Keyword,
		filter.Reported,
		pageSize,
		(page-1)*pageSize,
	)
	if err != nil {
		return nil, err
	}

	items := make([]dto.AdminTopupItem, 0, len(rows))
	for i := range rows {
		row := &rows[i]
		items = append(items, dto.AdminTopupItem{
			TopupItem:   toTopupItem(&row.WalletTopup),
			UserID:      row.UserID,
			Username:    row.Username,
			Email:       row.Email,
			DisplayName: row.DisplayName,
			AvatarURI:   row.AvatarURI,
		})
	}
	return &dto.AdminTopupListResponse{
		Data:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *walletServiceImpl) ApproveTopup(ctx context.Context, adminID, topupID string, input dto.ApproveTopupInput) (*dto.TopupItem, error) {
	topup, err := s.repo.FindTopupByID(ctx, topupID)
	if err != nil {
		return nil, err
	}
	if topup.Status != models.WalletTopupPending {
		return nil, errorsapp.New(errorsapp.ErrCodeWalletTopupNotPending)
	}

	credit := topup.Amount
	if input.ConfirmedAmount != nil {
		credit = *input.ConfirmedAmount
	}
	if credit <= 0 {
		return nil, errorsapp.New(errorsapp.ErrCodeWalletInvalidAmount)
	}
	note := strings.TrimSpace(input.Note)

	if err := s.repo.ApproveTopup(ctx, adminID, topupID, credit, note); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindTopupByID(ctx, topupID)
	if err != nil {
		return nil, err
	}

	if s.notificationService != nil {
		message := fmt.Sprintf("Ví của bạn đã được cộng %s từ yêu cầu nạp %s.", formatVND(credit), topup.Code)
		if credit != topup.Amount {
			message = fmt.Sprintf(
				"Ví của bạn đã được cộng %s (yêu cầu %s) từ yêu cầu nạp %s.",
				formatVND(credit), formatVND(topup.Amount), topup.Code,
			)
		}
		_, _ = s.notificationService.Create(ctx, topup.UserID, nil, models.NotificationTypeMessage, message, nil, nil, nil)
	}

	s.logAudit(ctx, adminID, "wallet_topup.approve", topupID, note, map[string]any{
		"code":             topup.Code,
		"requested_amount": topup.Amount,
		"confirmed_amount": credit,
	})

	item := toTopupItem(updated)
	return &item, nil
}

func (s *walletServiceImpl) RejectTopup(ctx context.Context, adminID, topupID string, input dto.RejectTopupInput) (*dto.TopupItem, error) {
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return nil, errorsapp.New(errorsapp.ErrCodeWalletReasonRequired)
	}

	topup, err := s.repo.FindTopupByID(ctx, topupID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.RejectTopup(ctx, adminID, topupID, reason); err != nil {
		return nil, err
	}

	if s.notificationService != nil {
		message := fmt.Sprintf("Yêu cầu nạp %s đã bị từ chối. Lý do: %s", topup.Code, reason)
		_, _ = s.notificationService.Create(ctx, topup.UserID, nil, models.NotificationTypeMessage, message, nil, nil, nil)
	}

	s.logAudit(ctx, adminID, "wallet_topup.reject", topupID, reason, map[string]any{
		"code":   topup.Code,
		"amount": topup.Amount,
	})

	updated, err := s.repo.FindTopupByID(ctx, topupID)
	if err != nil {
		return nil, err
	}
	item := toTopupItem(updated)
	return &item, nil
}

func (s *walletServiceImpl) AdminListWallets(ctx context.Context, keyword string, page, pageSize int) (*dto.AdminWalletListResponse, error) {
	page, pageSize = normalizePage(page, pageSize)
	rows, total, err := s.repo.AdminListWallets(ctx, keyword, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	balance, reserved, topup, err := s.repo.AdminWalletTotals(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]dto.AdminWalletItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.AdminWalletItem{
			UserID:      row.UserID,
			Username:    row.Username,
			Email:       row.Email,
			DisplayName: row.DisplayName,
			AvatarURI:   row.AvatarURI,
			Balance:     row.Balance,
			Reserved:    row.Reserved,
			Available:   row.Available(),
			TotalTopup:  row.TotalTopup,
			UpdatedAt:   row.UpdatedAt,
		})
	}

	return &dto.AdminWalletListResponse{
		Data:          items,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
		TotalBalance:  balance,
		TotalReserved: reserved,
		TotalTopupAll: topup,
	}, nil
}

func (s *walletServiceImpl) ExpireTopups(ctx context.Context) (int, error) {
	rows, err := s.repo.ExpireTopups(ctx, time.Now())
	return int(rows), err
}

// logAudit ghi audit best-effort: lỗi ghi log không được làm hỏng hành động chính.
func (s *walletServiceImpl) logAudit(ctx context.Context, adminID, action, targetID, reason string, metadata map[string]any) {
	if s.auditRepo == nil {
		return
	}
	raw := ""
	if len(metadata) > 0 {
		if data, err := json.Marshal(metadata); err == nil {
			raw = string(data)
		}
	}
	entry := &models.AdminAuditLog{
		ID:         utils.GenerateUUID(),
		AdminID:    adminID,
		Action:     action,
		TargetType: "wallet_topup",
		TargetID:   targetID,
		Reason:     reason,
		Metadata:   raw,
		CreatedAt:  time.Now(),
	}
	_ = s.auditRepo.Create(ctx, entry)
}

func toTopupItem(topup *models.WalletTopup) dto.TopupItem {
	return dto.TopupItem{
		ID:              topup.ID,
		Code:            topup.Code,
		Amount:          topup.Amount,
		Status:          string(topup.Status),
		ExpiresAt:       topup.ExpiresAt,
		UserReportedAt:  topup.UserReportedAt,
		ProofURL:        topup.ProofURL,
		ConfirmedAmount: topup.ConfirmedAmount,
		AdminNote:       topup.AdminNote,
		PaidAt:          topup.PaidAt,
		CreatedAt:       topup.CreatedAt,
	}
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "1062")
}

// formatVND định dạng số tiền theo kiểu Việt Nam (1.234.567) — dùng trong câu
// thông báo và thông báo lỗi, không dùng cho hiển thị UI (web tự format).
func formatVND(amount int64) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}
	digits := fmt.Sprintf("%d", amount)
	var b strings.Builder
	for i, ch := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(ch)
	}
	if negative {
		return "-" + b.String()
	}
	return b.String()
}
