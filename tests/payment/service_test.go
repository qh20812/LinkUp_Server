package payment_test

import (
	"net/url"
	"strconv"
	"testing"
	"time"

	"linkup/config"
	"linkup/dto"
	errorsapp "linkup/errors"
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
func (f *fakePackageRepository) ReserveSlot(string) (bool, error)      { return true, nil }
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

	wallet         *models.Wallet
	purchaseErr    error
	purchaseCalled bool
	purchaseSub    *models.PartnerSubscription
	purchaseTxn    *models.PaymentTransaction
	purchaseAmount int64
}

func (f *fakePaymentRepository) Create(txn *models.PaymentTransaction) error {
	f.created = txn
	return nil
}
func (f *fakePaymentRepository) FindByTxnRef(txnRef string) (*models.PaymentTransaction, error) {
	if f.txn != nil {
		return f.txn, nil
	}
	return f.created, nil
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
func (f *fakePaymentRepository) PurchaseWithWallet(
	userID string,
	pkg *models.AdPackage,
	activeSub, newSub *models.PartnerSubscription,
	txn *models.PaymentTransaction,
	amount int64,
) (*models.Wallet, error) {
	f.purchaseCalled = true
	if f.purchaseErr != nil {
		return nil, f.purchaseErr
	}
	f.purchaseAmount = amount
	f.purchaseSub = newSub
	f.purchaseTxn = txn

	txn.Status = models.PaymentStatusSuccess
	txn.Gateway = "wallet"
	txn.Amount = amount
	f.created = txn

	wallet := f.wallet
	if wallet == nil {
		wallet = &models.Wallet{UserID: userID, Balance: 0, Currency: "VND"}
	} else {
		wallet.Balance -= amount
	}
	return wallet, nil
}

func vnpayTestEnv() config.Env {
	return config.Env{
		VNPayTMNCode:         "TEST",
		VNPayHashSecret:      "SECRET",
		VNPayPaymentURL:      "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html",
		VNPayReturnURL:       "http://localhost:8080/api/payments/vnpay/return",
		PaymentVNPayEnabled:  true,
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

func walletTestEnv() config.Env {
	return config.Env{PaymentWalletEnabled: true}
}

func requireAppCode(t *testing.T, err error, want string) {
	t.Helper()
	appErr, ok := errorsapp.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError %s, got %v", want, err)
	}
	if appErr.Code != want {
		t.Fatalf("error code = %s, want %s", appErr.Code, want)
	}
}

func TestCreateWalletPaymentDeductsBalanceAndActivatesPackage(t *testing.T) {
	pkg := &models.AdPackage{ID: "pkg-1", Name: "Starter", PriceMonthly: 99000.4, MaxSlots: 3, MaxDurationDays: 30}
	pkgRepo := &fakePackageRepository{pkg: pkg}
	payRepo := &fakePaymentRepository{
		wallet: &models.Wallet{UserID: "user-1", Balance: 500000, Reserved: 50000, Currency: "VND"},
	}
	svc := services.NewPaymentService(payRepo, pkgRepo, walletTestEnv())

	res, err := svc.CreateWalletPayment("user-1", "pkg-1", "buy-1")
	if err != nil {
		t.Fatalf("CreateWalletPayment returned error: %v", err)
	}

	if payRepo.purchaseAmount != 99000 {
		t.Errorf("amount = %d, want 99000 (rounded)", payRepo.purchaseAmount)
	}
	if res.Payment.Gateway != "wallet" || res.Payment.Status != models.PaymentStatusSuccess {
		t.Errorf("unexpected payment: %#v", res.Payment)
	}
	if res.Wallet == nil || res.Wallet.Available != 500000-99000-50000 {
		t.Errorf("wallet = %#v, want available 351000", res.Wallet)
	}
	if res.Subscription.PackageID != "pkg-1" || res.Subscription.MaxSlots != 3 {
		t.Errorf("unexpected subscription: %#v", res.Subscription)
	}
	if res.Replayed {
		t.Error("first purchase must not be marked replayed")
	}
	wantExpiry := time.Now().AddDate(0, 0, 30)
	if diff := res.Subscription.ExpiresAt.Sub(wantExpiry); diff > time.Minute || diff < -time.Minute {
		t.Errorf("expires_at = %v, want about %v", res.Subscription.ExpiresAt, wantExpiry)
	}
	if payRepo.purchaseSub == nil || payRepo.purchaseSub.UserID != "user-1" {
		t.Errorf("subscription passed to repository = %#v", payRepo.purchaseSub)
	}
}

func TestCreateWalletPaymentRejectsInsufficientBalance(t *testing.T) {
	pkgRepo := &fakePackageRepository{
		pkg: &models.AdPackage{ID: "pkg-1", Name: "Starter", PriceMonthly: 99000, MaxSlots: 1, MaxDurationDays: 30},
	}
	payRepo := &fakePaymentRepository{
		purchaseErr: errorsapp.Newf(errorsapp.ErrCodeWalletInsufficient, map[string]any{
			"required": 99000, "available": 10000,
		}),
	}
	svc := services.NewPaymentService(payRepo, pkgRepo, walletTestEnv())

	_, err := svc.CreateWalletPayment("user-1", "pkg-1", "buy-2")
	requireAppCode(t, err, errorsapp.ErrCodeWalletInsufficient)
}

func TestCreateWalletPaymentAllowsFreePackage(t *testing.T) {
	pkgRepo := &fakePackageRepository{
		pkg: &models.AdPackage{ID: "pkg-1", Name: "Free", PriceMonthly: 0, MaxSlots: 1, MaxDurationDays: 30},
	}
	payRepo := &fakePaymentRepository{wallet: &models.Wallet{Balance: 500000, Currency: "VND"}}
	svc := services.NewPaymentService(payRepo, pkgRepo, walletTestEnv())

	res, err := svc.CreateWalletPayment("user-1", "pkg-1", "free-1")
	if err != nil {
		t.Fatalf("free package must activate without payment: %v", err)
	}
	if payRepo.purchaseAmount != 0 {
		t.Errorf("amount = %d, want 0", payRepo.purchaseAmount)
	}
	if res.Wallet == nil || res.Wallet.Balance != 500000 {
		t.Errorf("wallet changed for free package: %#v", res.Wallet)
	}
}

func TestCreateWalletPaymentDisabledByEnv(t *testing.T) {
	pkgRepo := &fakePackageRepository{
		pkg: &models.AdPackage{ID: "pkg-1", Name: "Starter", PriceMonthly: 99000},
	}
	payRepo := &fakePaymentRepository{}
	svc := services.NewPaymentService(payRepo, pkgRepo, config.Env{})

	_, err := svc.CreateWalletPayment("user-1", "pkg-1", "buy-3")
	requireAppCode(t, err, errorsapp.ErrCodePaymentMethodOff)
}

func TestCreateWalletPaymentIsIdempotentWithSameKey(t *testing.T) {
	pkg := &models.AdPackage{ID: "pkg-1", Name: "Starter", PriceMonthly: 99000, MaxSlots: 1, MaxDurationDays: 30}
	pkgRepo := &fakePackageRepository{pkg: pkg}
	payRepo := &fakePaymentRepository{wallet: &models.Wallet{Balance: 500000, Currency: "VND"}}
	svc := services.NewPaymentService(payRepo, pkgRepo, walletTestEnv())

	first, err := svc.CreateWalletPayment("user-1", "pkg-1", "same-key")
	if err != nil {
		t.Fatalf("first purchase failed: %v", err)
	}

	// Mô phỏng DB đã kích hoạt gói ở lần mua đầu.
	pkgRepo.active = &models.PartnerSubscription{
		ID: first.Subscription.ID, UserID: "user-1", PackageID: "pkg-1", Package: pkg,
	}
	payRepo.purchaseCalled = false

	second, err := svc.CreateWalletPayment("user-1", "pkg-1", "same-key")
	if err != nil {
		t.Fatalf("replay failed: %v", err)
	}
	if !second.Replayed {
		t.Error("second call with same key must be marked replayed")
	}
	if payRepo.purchaseCalled {
		t.Fatal("replay must not deduct the wallet again")
	}
	if second.Payment.TxnRef != first.Payment.TxnRef {
		t.Errorf("txn_ref changed between replay: %q vs %q", second.Payment.TxnRef, first.Payment.TxnRef)
	}
}

func TestCreateWalletPaymentRejectsActiveSamePackage(t *testing.T) {
	pkg := &models.AdPackage{ID: "pkg-1", Name: "Starter", PriceMonthly: 99000, MaxSlots: 1, MaxDurationDays: 30}
	pkgRepo := &fakePackageRepository{
		pkg:    pkg,
		active: &models.PartnerSubscription{ID: "sub-1", UserID: "user-1", PackageID: "pkg-1", Package: pkg},
	}
	payRepo := &fakePaymentRepository{wallet: &models.Wallet{Balance: 500000, Currency: "VND"}}
	svc := services.NewPaymentService(payRepo, pkgRepo, walletTestEnv())

	_, err := svc.CreateWalletPayment("user-1", "pkg-1", "buy-4")
	requireAppCode(t, err, errorsapp.ErrCodePaymentAlreadyActive)
	if payRepo.purchaseCalled {
		t.Fatal("must not deduct wallet when the package is already active")
	}
}

func TestGetPaymentMethodsReflectsEnv(t *testing.T) {
	svc := services.NewPaymentService(&fakePaymentRepository{}, &fakePackageRepository{}, walletTestEnv())
	res := svc.GetPaymentMethods()

	if res.Default != "wallet" {
		t.Errorf("default = %q, want wallet", res.Default)
	}
	if len(res.Methods) != 3 {
		t.Fatalf("methods = %d, want 3", len(res.Methods))
	}
	enabled := map[string]bool{}
	for _, m := range res.Methods {
		enabled[m.Key] = m.Enabled
	}
	if !enabled["wallet"] {
		t.Error("wallet must be enabled")
	}
	if enabled["momo"] || enabled["vnpay"] {
		t.Error("momo/vnpay must stay disabled by default")
	}

	// Bật cờ nhưng thiếu cấu hình merchant vẫn phải ẩn.
	svc = services.NewPaymentService(&fakePaymentRepository{}, &fakePackageRepository{}, config.Env{
		PaymentWalletEnabled: true,
		PaymentMoMoEnabled:   true,
		PaymentVNPayEnabled:  true,
	})
	for _, m := range svc.GetPaymentMethods().Methods {
		if m.Key != "wallet" && m.Enabled {
			t.Errorf("%s must stay hidden without merchant config", m.Key)
		}
	}
}
