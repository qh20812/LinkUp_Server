package payment_test

import (
	"net/url"
	"strconv"
	"testing"
	"time"

	"linkup/config"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/services"
	"linkup/utils"
)

type fakePackageRepository struct {
	pkg    *models.AdPackage
	active *models.PartnerSubscription
}

func (f *fakePackageRepository) GetAllPackages() ([]models.AdPackage, error) { return nil, nil }
func (f *fakePackageRepository) GetActivePackages() ([]models.AdPackage, error) {
	return nil, nil
}
func (f *fakePackageRepository) CreatePackage(*models.AdPackage) error { return nil }
func (f *fakePackageRepository) UpdatePackage(*models.AdPackage) error { return nil }
func (f *fakePackageRepository) FindExpiringSubscriptions(time.Time) ([]models.PartnerSubscription, error) {
	return nil, nil
}
func (f *fakePackageRepository) GetPackageByID(id string) (*models.AdPackage, error) {
	return f.pkg, nil
}
func (f *fakePackageRepository) GetActiveSubscription(userID string) (*models.PartnerSubscription, error) {
	return f.active, nil
}
func (f *fakePackageRepository) CreateSubscription(*models.PartnerSubscription) error { return nil }
func (f *fakePackageRepository) UpdateSubscription(*models.PartnerSubscription) error { return nil }
func (f *fakePackageRepository) IncrementSlotsUsed(string) error                      { return nil }
func (f *fakePackageRepository) DecrementSlotsUsed(string) error                      { return nil }
func (f *fakePackageRepository) UpdateUserRole(string, models.RoleName) error         { return nil }
func (f *fakePackageRepository) SubscribeWithTransaction(*models.PartnerSubscription, *models.PartnerSubscription) error {
	return nil
}
func (f *fakePackageRepository) FindExpiredSubscriptions() ([]models.PartnerSubscription, error) {
	return nil, nil
}
func (f *fakePackageRepository) ExpireSubscription(string) error   { return nil }
func (f *fakePackageRepository) DowngradePartnerRole(string) error { return nil }

type fakePaymentRepository struct {
	created      *models.PaymentTransaction
	txn          *models.PaymentTransaction
	completed    bool
	completedSub *models.PartnerSubscription
}

func (f *fakePaymentRepository) Create(txn *models.PaymentTransaction) error {
	f.created = txn
	return nil
}
func (f *fakePaymentRepository) FindByTxnRef(txnRef string) (*models.PaymentTransaction, error) {
	return f.txn, nil
}
func (f *fakePaymentRepository) CompletePayment(
	txnRef string,
	result repository.PaymentResult,
	activeSub, newSub *models.PartnerSubscription,
) (bool, error) {
	f.completed = true
	f.completedSub = newSub
	return true, nil
}
func (f *fakePaymentRepository) MarkFailed(string, repository.PaymentResult) (bool, error) {
	return true, nil
}
func (f *fakePaymentRepository) List(dto.PaymentListFilter, int, int) ([]models.PaymentTransaction, int64, error) {
	return nil, 0, nil
}

func vnpayTestEnv() config.Env {
	return config.Env{
		VNPayTMNCode:    "TEST",
		VNPayHashSecret: "SECRET",
		VNPayPaymentURL: "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html",
		VNPayReturnURL:  "http://localhost:8080/api/payments/vnpay/return",
	}
}

func TestCreateVNPayPaymentBuildsSignedURL(t *testing.T) {
	pkgRepo := &fakePackageRepository{
		pkg: &models.AdPackage{ID: "pkg-1", Name: "Starter", PriceMonthly: 990000, MaxDurationDays: 15},
	}
	payRepo := &fakePaymentRepository{}
	svc := services.NewPaymentService(payRepo, pkgRepo, vnpayTestEnv())

	res, err := svc.CreateVNPayPayment("user-1", "pkg-1", "203.0.113.7")
	if err != nil {
		t.Fatalf("CreateVNPayPayment returned error: %v", err)
	}

	parsed, err := url.Parse(res.PaymentURL)
	if err != nil {
		t.Fatalf("payment URL is invalid: %v", err)
	}
	q := parsed.Query()
	if q.Get("vnp_Amount") != "99000000" {
		t.Errorf("vnp_Amount = %q, want 99000000", q.Get("vnp_Amount"))
	}
	if q.Get("vnp_TmnCode") != "TEST" {
		t.Errorf("vnp_TmnCode = %q, want TEST", q.Get("vnp_TmnCode"))
	}
	if q.Get("vnp_IpAddr") != "203.0.113.7" {
		t.Errorf("vnp_IpAddr = %q, want 203.0.113.7", q.Get("vnp_IpAddr"))
	}
	if q.Get("vnp_TxnRef") != res.TxnRef {
		t.Errorf("vnp_TxnRef = %q, want %q", q.Get("vnp_TxnRef"), res.TxnRef)
	}
	if !utils.VNPayVerify(utils.VNPayParamsFromValues(q), "SECRET", q.Get("vnp_SecureHash")) {
		t.Error("payment URL signature is invalid")
	}
	if payRepo.created == nil || payRepo.created.Amount != 990000 {
		t.Fatalf("unexpected created transaction: %#v", payRepo.created)
	}
}

func TestCreateVNPayPaymentRejectsActivePackage(t *testing.T) {
	pkgRepo := &fakePackageRepository{
		pkg:    &models.AdPackage{ID: "pkg-1", Name: "Starter", PriceMonthly: 990000, MaxDurationDays: 15},
		active: &models.PartnerSubscription{ID: "sub-1", UserID: "user-1", PackageID: "pkg-1"},
	}
	svc := services.NewPaymentService(&fakePaymentRepository{}, pkgRepo, vnpayTestEnv())

	if _, err := svc.CreateVNPayPayment("user-1", "pkg-1", "203.0.113.7"); err == nil {
		t.Fatal("expected active package to be rejected")
	}
}

func TestProcessVNPayCallbackActivatesSubscription(t *testing.T) {
	pkgRepo := &fakePackageRepository{
		pkg: &models.AdPackage{ID: "pkg-1", Name: "Starter", PriceMonthly: 990000, MaxDurationDays: 15},
	}
	payRepo := &fakePaymentRepository{
		txn: &models.PaymentTransaction{
			ID:        "txn-1",
			TxnRef:    "LK1",
			UserID:    "user-1",
			PackageID: "pkg-1",
			Amount:    990000,
			Status:    models.PaymentStatusPending,
		},
	}
	svc := services.NewPaymentService(payRepo, pkgRepo, vnpayTestEnv())

	params := map[string]string{
		"vnp_TmnCode":           "TEST",
		"vnp_TxnRef":            "LK1",
		"vnp_Amount":            "99000000",
		"vnp_ResponseCode":      "00",
		"vnp_TransactionStatus": "00",
		"vnp_TransactionNo":     "123456",
		"vnp_BankCode":          "NCB",
		"vnp_PayDate":           "20261031120000",
	}
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	values.Set("vnp_SecureHash", utils.VNPaySign(params, "SECRET"))

	result := svc.ProcessVNPayCallback(values)
	if !result.Valid || !result.Success || result.AlreadyConfirmed {
		t.Fatalf("unexpected result: %#v", result)
	}
	if !payRepo.completed || payRepo.completedSub == nil {
		t.Fatal("subscription was not activated")
	}
	if payRepo.completedSub.UserID != "user-1" || payRepo.completedSub.PackageID != "pkg-1" {
		t.Errorf("unexpected subscription: %#v", payRepo.completedSub)
	}
	wantExpiry := time.Now().AddDate(0, 0, 15)
	if diff := payRepo.completedSub.ExpiresAt.Sub(wantExpiry); diff > time.Minute || diff < -time.Minute {
		t.Errorf("expires_at = %v, want about %v", payRepo.completedSub.ExpiresAt, wantExpiry)
	}
}

func TestProcessVNPayCallbackRejectsAmountMismatch(t *testing.T) {
	pkgRepo := &fakePackageRepository{
		pkg: &models.AdPackage{ID: "pkg-1", Name: "Starter", PriceMonthly: 990000, MaxDurationDays: 15},
	}
	payRepo := &fakePaymentRepository{
		txn: &models.PaymentTransaction{
			ID:        "txn-1",
			TxnRef:    "LK1",
			UserID:    "user-1",
			PackageID: "pkg-1",
			Amount:    990000,
			Status:    models.PaymentStatusPending,
		},
	}
	svc := services.NewPaymentService(payRepo, pkgRepo, vnpayTestEnv())

	params := map[string]string{
		"vnp_TmnCode":           "TEST",
		"vnp_TxnRef":            "LK1",
		"vnp_Amount":            strconv.FormatInt(1000, 10),
		"vnp_ResponseCode":      "00",
		"vnp_TransactionStatus": "00",
	}
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	values.Set("vnp_SecureHash", utils.VNPaySign(params, "SECRET"))

	result := svc.ProcessVNPayCallback(values)
	if !result.Valid || !result.AmountMismatch || result.Success {
		t.Fatalf("unexpected result: %#v", result)
	}
	if payRepo.completed {
		t.Fatal("subscription must not be activated on amount mismatch")
	}
}
