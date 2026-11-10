package wallet_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"linkup/config"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/services"
)

// fakeWalletRepo chỉ cài đặt các method service dùng tới; phần còn lại nhúng
// interface nil nên nếu bị gọi sẽ panic (giúp phát hiện phụ thuộc ngoài ý muốn).
type fakeWalletRepo struct {
	repository.WalletRepository

	wallet    *models.Wallet
	topup     *models.WalletTopup
	created   []*models.WalletTopup
	openCount int64
	today     int64

	approvedCredit int64
	approvedNote   string
	rejectedReason string
	reportedProof  string
}

func (f *fakeWalletRepo) GetOrCreate(_ context.Context, userID string) (*models.Wallet, error) {
	if f.wallet != nil {
		return f.wallet, nil
	}
	return &models.Wallet{UserID: userID, Currency: "VND"}, nil
}

func (f *fakeWalletRepo) SumLedger(_ context.Context, _ string, _ models.WalletTransactionType) (int64, error) {
	return 0, nil
}

func (f *fakeWalletRepo) CountOpenTopups(_ context.Context, _ string) (int64, error) {
	return f.openCount, nil
}

func (f *fakeWalletRepo) CountTopupsSince(_ context.Context, _ string, _ time.Time) (int64, error) {
	return f.today, nil
}

func (f *fakeWalletRepo) CreateTopup(_ context.Context, topup *models.WalletTopup) error {
	f.topup = topup
	f.created = append(f.created, topup)
	return nil
}

func (f *fakeWalletRepo) FindTopupByID(_ context.Context, _ string) (*models.WalletTopup, error) {
	if f.topup == nil {
		return nil, errorsapp.New(errorsapp.ErrCodeWalletTopupNotFound)
	}
	return f.topup, nil
}

func (f *fakeWalletRepo) MarkTopupReported(_ context.Context, _ string, _ string, proof string) error {
	f.reportedProof = proof
	if f.topup != nil {
		now := time.Now()
		f.topup.UserReportedAt = &now
		f.topup.ProofURL = proof
	}
	return nil
}

func (f *fakeWalletRepo) ApproveTopup(_ context.Context, _ string, _ string, credit int64, note string) error {
	f.approvedCredit = credit
	f.approvedNote = note
	if f.topup != nil {
		now := time.Now()
		f.topup.Status = models.WalletTopupPaid
		f.topup.ConfirmedAmount = &credit
		f.topup.AdminNote = note
		f.topup.PaidAt = &now
	}
	return nil
}

func (f *fakeWalletRepo) RejectTopup(_ context.Context, _ string, _ string, reason string) error {
	f.rejectedReason = reason
	if f.topup != nil {
		f.topup.Status = models.WalletTopupRejected
		f.topup.AdminNote = reason
	}
	return nil
}

func bankEnv() config.Env {
	return config.Env{
		BankBIN:           "970407",
		BankAccountNo:     "23712299999999",
		BankAccountName:   "HUYNH NGOC QUI",
		BankShortName:     "Techcombank",
		TopupQRBaseURL:    "https://img.vietqr.io/image",
		TopupExpiresMin:   30,
		TopupMinAmount:    10000,
		TopupMaxAmount:    50000000,
		TopupDailyLimit:   10,
		TopupPendingLimit: 5,
	}
}

func codeOf(t *testing.T, err error) string {
	t.Helper()
	appErr, ok := errorsapp.IsAppError(err)
	if !ok {
		t.Fatalf("err = %v, want AppError", err)
	}
	return appErr.Code
}

func TestCreateTopup_NotConfigured(t *testing.T) {
	svc := services.NewWalletService(&fakeWalletRepo{}, config.Env{})
	_, err := svc.CreateTopup(context.Background(), "u1", 100000)
	if code := codeOf(t, err); code != errorsapp.ErrCodeWalletNotConfigured {
		t.Fatalf("code = %s, want %s", code, errorsapp.ErrCodeWalletNotConfigured)
	}
}

func TestCreateTopup_Success(t *testing.T) {
	repo := &fakeWalletRepo{}
	svc := services.NewWalletService(repo, bankEnv())

	res, err := svc.CreateTopup(context.Background(), "u1", 200000)
	if err != nil {
		t.Fatalf("CreateTopup: %v", err)
	}
	if len(res.Code) != 8 || !strings.HasPrefix(res.Code, "LK") {
		t.Fatalf("code = %q, want 8 chars LKxxxxxx", res.Code)
	}
	if res.Amount != 200000 {
		t.Fatalf("amount = %d, want 200000", res.Amount)
	}
	if !strings.Contains(res.QRURL, "amount=200000") || !strings.Contains(res.QRURL, "addInfo="+res.Code) {
		t.Fatalf("qr = %q, want amount + memo", res.QRURL)
	}
	if res.Bank.AccountNo != "23712299999999" {
		t.Fatalf("bank = %+v", res.Bank)
	}
	if len(repo.created) != 1 || repo.created[0].Status != models.WalletTopupPending {
		t.Fatalf("created = %+v, want one pending topup", repo.created)
	}
	if got := time.Until(res.ExpiresAt); got < 29*time.Minute || got > 31*time.Minute {
		t.Fatalf("expiry = %v, want ~30m", got)
	}
}

func TestCreateTopup_AmountValidation(t *testing.T) {
	svc := services.NewWalletService(&fakeWalletRepo{}, bankEnv())
	_, err := svc.CreateTopup(context.Background(), "u1", 5000)
	if code := codeOf(t, err); code != errorsapp.ErrCodeWalletAmountTooSmall {
		t.Fatalf("code = %s, want %s", code, errorsapp.ErrCodeWalletAmountTooSmall)
	}
}

func TestCreateTopup_RateLimitedByOpenTopups(t *testing.T) {
	repo := &fakeWalletRepo{openCount: 5}
	svc := services.NewWalletService(repo, bankEnv())
	_, err := svc.CreateTopup(context.Background(), "u1", 100000)
	if code := codeOf(t, err); code != errorsapp.ErrCodeWalletTopupRateLimited {
		t.Fatalf("code = %s, want %s", code, errorsapp.ErrCodeWalletTopupRateLimited)
	}
}

func TestReportTopup_Expired(t *testing.T) {
	repo := &fakeWalletRepo{topup: &models.WalletTopup{
		ID:        "t1",
		UserID:    "u1",
		Code:      "LKABCDEF",
		Amount:    100000,
		Status:    models.WalletTopupPending,
		ExpiresAt: time.Now().Add(-time.Minute),
	}}
	svc := services.NewWalletService(repo, bankEnv())

	_, err := svc.ReportTopup(context.Background(), "u1", "t1", "")
	if code := codeOf(t, err); code != errorsapp.ErrCodeWalletTopupExpired {
		t.Fatalf("code = %s, want %s", code, errorsapp.ErrCodeWalletTopupExpired)
	}
}

func TestReportTopup_NotOwner(t *testing.T) {
	repo := &fakeWalletRepo{topup: &models.WalletTopup{
		ID:        "t1",
		UserID:    "u1",
		Status:    models.WalletTopupPending,
		ExpiresAt: time.Now().Add(time.Minute),
	}}
	svc := services.NewWalletService(repo, bankEnv())

	_, err := svc.ReportTopup(context.Background(), "u2", "t1", "")
	if code := codeOf(t, err); code != errorsapp.ErrCodeWalletNotOwner {
		t.Fatalf("code = %s, want %s", code, errorsapp.ErrCodeWalletNotOwner)
	}
}

func TestReportTopup_RequiresProof(t *testing.T) {
	repo := &fakeWalletRepo{topup: &models.WalletTopup{
		ID:        "t1",
		UserID:    "u1",
		Status:    models.WalletTopupPending,
		ExpiresAt: time.Now().Add(time.Minute),
	}}
	svc := services.NewWalletService(repo, bankEnv())

	_, err := svc.ReportTopup(context.Background(), "u1", "t1", "   ")
	if code := codeOf(t, err); code != errorsapp.ErrCodeWalletProofRequired {
		t.Fatalf("code = %s, want %s", code, errorsapp.ErrCodeWalletProofRequired)
	}
	if repo.reportedProof != "" {
		t.Fatalf("repo must not be called without proof, got %q", repo.reportedProof)
	}
}

func TestReportTopup_WithProof(t *testing.T) {
	repo := &fakeWalletRepo{topup: &models.WalletTopup{
		ID:        "t1",
		UserID:    "u1",
		Status:    models.WalletTopupPending,
		ExpiresAt: time.Now().Add(time.Minute),
	}}
	svc := services.NewWalletService(repo, bankEnv())

	const proof = "https://res.cloudinary.com/demo/bill.jpg"
	if _, err := svc.ReportTopup(context.Background(), "u1", "t1", proof); err != nil {
		t.Fatalf("ReportTopup returned error: %v", err)
	}
	if repo.reportedProof != proof {
		t.Fatalf("reportedProof = %q, want %q", repo.reportedProof, proof)
	}
}

func TestApproveTopup_UsesConfirmedAmount(t *testing.T) {
	repo := &fakeWalletRepo{topup: &models.WalletTopup{
		ID:     "t1",
		UserID: "u1",
		Code:   "LKABCDEF",
		Amount: 100000,
		Status: models.WalletTopupPending,
	}}
	svc := services.NewWalletService(repo, bankEnv())

	credit := int64(90000)
	item, err := svc.ApproveTopup(context.Background(), "admin1", "t1", dto.ApproveTopupInput{
		ConfirmedAmount: &credit,
		Note:            "Nhận 90k",
	})
	if err != nil {
		t.Fatalf("ApproveTopup: %v", err)
	}
	if repo.approvedCredit != 90000 {
		t.Fatalf("credited = %d, want 90000", repo.approvedCredit)
	}
	if item.Status != string(models.WalletTopupPaid) {
		t.Fatalf("status = %s, want paid", item.Status)
	}
}

func TestApproveTopup_NotPending(t *testing.T) {
	repo := &fakeWalletRepo{topup: &models.WalletTopup{
		ID:     "t1",
		UserID: "u1",
		Status: models.WalletTopupPaid,
	}}
	svc := services.NewWalletService(repo, bankEnv())

	_, err := svc.ApproveTopup(context.Background(), "admin1", "t1", dto.ApproveTopupInput{})
	if code := codeOf(t, err); code != errorsapp.ErrCodeWalletTopupNotPending {
		t.Fatalf("code = %s, want %s", code, errorsapp.ErrCodeWalletTopupNotPending)
	}
}

func TestRejectTopup_RequiresReason(t *testing.T) {
	repo := &fakeWalletRepo{topup: &models.WalletTopup{
		ID:     "t1",
		UserID: "u1",
		Status: models.WalletTopupPending,
	}}
	svc := services.NewWalletService(repo, bankEnv())

	_, err := svc.RejectTopup(context.Background(), "admin1", "t1", dto.RejectTopupInput{})
	if code := codeOf(t, err); code != errorsapp.ErrCodeWalletReasonRequired {
		t.Fatalf("code = %s, want %s", code, errorsapp.ErrCodeWalletReasonRequired)
	}
}
