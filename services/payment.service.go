package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"linkup/config"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"log"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const vnpayPaymentTTL = 15 * time.Minute

type vnpayConfig struct {
	tmnCode    string
	hashSecret string
	paymentURL string
	returnURL  string
}

func vnpayConfigFromEnv(env config.Env) vnpayConfig {
	return vnpayConfig{
		tmnCode:    strings.TrimSpace(env.VNPayTMNCode),
		hashSecret: strings.TrimSpace(env.VNPayHashSecret),
		paymentURL: strings.TrimSpace(env.VNPayPaymentURL),
		returnURL:  strings.TrimSpace(env.VNPayReturnURL),
	}
}

func (c vnpayConfig) enabled() bool {
	return c.tmnCode != "" && c.hashSecret != "" && c.paymentURL != "" && c.returnURL != ""
}

type PaymentService interface {
	CreateVNPayPayment(userID, packageID, ip string) (*dto.CreatePaymentResponse, error)
	// CreateWalletPayment mua gói bằng ví trả trước. idempotencyKey lấy từ header
	// Idempotency-Key; gửi lặp cùng khoá trả lại kết quả cũ thay vì trừ tiền lần nữa.
	CreateWalletPayment(userID, packageID, idempotencyKey string) (*dto.WalletPurchaseResponse, error)
	// GetPaymentMethods trả các phương thức thanh toán gói đang bật.
	GetPaymentMethods() dto.PaymentMethodsResponse
	ProcessVNPayCallback(values url.Values) dto.PaymentCallbackResult
	// ListPayments trả lịch sử giao dịch (partner: filter.UserID đã bị ép).
	ListPayments(filter dto.PaymentListFilter) ([]models.PaymentTransaction, int64, error)
}

type paymentServiceImpl struct {
	payments repository.PaymentRepository
	packages repository.PackageRepository
	env      config.Env
}

func NewPaymentService(payments repository.PaymentRepository, packages repository.PackageRepository, env config.Env) PaymentService {
	return &paymentServiceImpl{payments: payments, packages: packages, env: env}
}

func (s *paymentServiceImpl) CreateVNPayPayment(userID, packageID, ip string) (*dto.CreatePaymentResponse, error) {
	if !s.env.PaymentVNPayEnabled {
		return nil, errorsapp.New(errorsapp.ErrCodePaymentMethodOff)
	}
	cfg := vnpayConfigFromEnv(s.env)
	if !cfg.enabled() {
		return nil, errorsapp.New(errorsapp.ErrCodePaymentNotConfigured)
	}

	pkg, err := s.packages.GetPackageByID(packageID)
	if err != nil {
		return nil, err
	}
	if pkg.PriceMonthly <= 0 {
		return nil, errorsapp.New(errorsapp.ErrCodePaymentInvalidAmount)
	}

	activeSub, err := s.packages.GetActiveSubscription(userID)
	if err != nil {
		return nil, err
	}
	if activeSub != nil && activeSub.PackageID == pkg.ID {
		return nil, errorsapp.New(errorsapp.ErrCodePaymentAlreadyActive)
	}

	amountVND := int64(math.Round(pkg.PriceMonthly))
	gatewayAmount, err := utils.VNPayAmount(pkg.PriceMonthly)
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodePaymentInvalidAmount)
	}

	now := time.Now()
	expiresAt := now.Add(vnpayPaymentTTL)
	txnRef := newVNPayTxnRef()
	orderInfo := "Thanh toan goi quang cao " + pkg.Name

	txn := &models.PaymentTransaction{
		ID:        utils.GenerateUUID(),
		TxnRef:    txnRef,
		UserID:    userID,
		PackageID: pkg.ID,
		Amount:    amountVND,
		Status:    models.PaymentStatusPending,
		Gateway:   "vnpay",
		OrderInfo: orderInfo,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.payments.Create(txn); err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodePaymentCreateFailed, err)
	}

	params := map[string]string{
		"vnp_Version":    "2.1.0",
		"vnp_Command":    "pay",
		"vnp_TmnCode":    cfg.tmnCode,
		"vnp_Locale":     "vn",
		"vnp_CurrCode":   "VND",
		"vnp_TxnRef":     txnRef,
		"vnp_OrderInfo":  orderInfo,
		"vnp_OrderType":  "other",
		"vnp_Amount":     strconv.FormatInt(gatewayAmount, 10),
		"vnp_ReturnUrl":  cfg.returnURL,
		"vnp_IpAddr":     utils.VNPayClientIP(ip),
		"vnp_CreateDate": utils.VNPayTime(now),
		"vnp_ExpireDate": utils.VNPayTime(expiresAt),
	}
	params["vnp_SecureHash"] = utils.VNPaySign(params, cfg.hashSecret)

	return &dto.CreatePaymentResponse{
		PaymentURL: cfg.paymentURL + "?" + utils.VNPayParamsToQuery(params),
		TxnRef:     txnRef,
		Amount:     amountVND,
		ExpiresAt:  expiresAt,
	}, nil
}

// CreateWalletPayment mua gói bằng số dư ví trả trước. Toàn bộ việc trừ tiền,
// huỷ giao dịch gateway đang chờ cùng gói, ghi payment và kích hoạt gói nằm
// trong một transaction ở repository.
//
// Idempotency: txn_ref được suy ra từ (user, package, Idempotency-Key) nên
// request lặp lại (double click, retry mạng) không trừ tiền lần hai. Nếu client
// không gửi khoá, mỗi request là một giao dịch mới.
func (s *paymentServiceImpl) CreateWalletPayment(userID, packageID, idempotencyKey string) (*dto.WalletPurchaseResponse, error) {
	if !s.env.PaymentWalletEnabled {
		return nil, errorsapp.New(errorsapp.ErrCodePaymentMethodOff)
	}

	pkg, err := s.packages.GetPackageByID(packageID)
	if err != nil {
		return nil, err
	}
	if pkg.PriceMonthly < 0 {
		return nil, errorsapp.New(errorsapp.ErrCodePaymentInvalidAmount)
	}
	// Giá 0 là gói miễn phí: vẫn kích hoạt gói nhưng không ghi bút toán ví.
	amount := int64(math.Round(pkg.PriceMonthly))
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if len(idempotencyKey) > 64 {
		idempotencyKey = idempotencyKey[:64]
	}
	txnRef := walletTxnRef(userID, packageID, idempotencyKey)

	// Replay phải chạy TRƯỚC kiểm tra "gói đang hoạt động": lần gọi thứ hai
	// của cùng một yêu cầu sẽ thấy gói vừa được kích hoạt ở lần đầu.
	existing, err := s.payments.FindByTxnRef(txnRef)
	if err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodePaymentCreateFailed, err)
	}
	if existing != nil {
		return s.replayWalletPurchase(userID, existing)
	}

	activeSub, err := s.packages.GetActiveSubscription(userID)
	if err != nil {
		return nil, err
	}
	if activeSub != nil && activeSub.PackageID == pkg.ID {
		return nil, errorsapp.New(errorsapp.ErrCodePaymentAlreadyActive)
	}

	now := time.Now()
	newSub := newPartnerSubscription(userID, pkg)
	txn := &models.PaymentTransaction{
		ID:        utils.GenerateUUID(),
		TxnRef:    txnRef,
		RequestID: strings.TrimSpace(idempotencyKey),
		UserID:    userID,
		PackageID: pkg.ID,
		Amount:    amount,
		Status:    models.PaymentStatusPending,
		Gateway:   "wallet",
		OrderInfo: "Thanh toan goi quang cao " + pkg.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	wallet, err := s.payments.PurchaseWithWallet(userID, pkg, activeSub, newSub, txn, amount)
	if err != nil {
		// Request song song cùng Idempotency-Key: giao dịch kia đã thắng, đọc
		// lại và trả kết quả thay vì báo lỗi cho người dùng.
		if errors.Is(err, repository.ErrWalletTxnExists) {
			existing, findErr := s.payments.FindByTxnRef(txnRef)
			if findErr != nil {
				return nil, errorsapp.Wrap(errorsapp.ErrCodePaymentCreateFailed, findErr)
			}
			if existing != nil {
				return s.replayWalletPurchase(userID, existing)
			}
		}
		return nil, err
	}

	resp := &dto.WalletPurchaseResponse{
		Subscription: walletSubscriptionResponse(newSub, pkg),
		Payment:      *txn,
	}
	if wallet != nil {
		resp.Wallet = &dto.WalletBalance{
			Balance:   wallet.Balance,
			Reserved:  wallet.Reserved,
			Available: wallet.Available(),
			Currency:  wallet.Currency,
		}
	}
	return resp, nil
}

// replayWalletPurchase trả lại kết quả của giao dịch ví đã thành công trước đó.
func (s *paymentServiceImpl) replayWalletPurchase(userID string, txn *models.PaymentTransaction) (*dto.WalletPurchaseResponse, error) {
	if txn.UserID != userID || txn.Status != models.PaymentStatusSuccess {
		return nil, errorsapp.New(errorsapp.ErrCodePaymentDuplicate)
	}

	activeSub, err := s.packages.GetActiveSubscription(userID)
	if err != nil {
		return nil, err
	}
	if activeSub == nil || activeSub.PackageID != txn.PackageID {
		// Giao dịch đã xử lý nhưng gói đã đổi/hết hạn — không thể dựng lại
		// subscription của lần mua cũ, và cũng không được trừ tiền lại.
		return nil, errorsapp.New(errorsapp.ErrCodePaymentDuplicate)
	}

	pkg := activeSub.Package
	if pkg == nil {
		pkg = &models.AdPackage{ID: txn.PackageID}
	}
	return &dto.WalletPurchaseResponse{
		Subscription: walletSubscriptionResponse(activeSub, pkg),
		Payment:      *txn,
		Replayed:     true,
	}, nil
}

// walletSubscriptionResponse chuyển subscription vừa tạo sang DTO. Không dùng
// GetUserSubscription vì hàm đó đọc lại DB và trả lỗi 404 khi chưa thấy commit.
func walletSubscriptionResponse(sub *models.PartnerSubscription, pkg *models.AdPackage) dto.SubscriptionResponse {
	slotsLeft := pkg.MaxSlots - sub.SlotsUsed
	if slotsLeft < 0 {
		slotsLeft = 0
	}
	return dto.SubscriptionResponse{
		ID:           sub.ID,
		PackageID:    sub.PackageID,
		PackageName:  pkg.Name,
		MaxSlots:     pkg.MaxSlots,
		SlotsUsed:    sub.SlotsUsed,
		SlotsLeft:    slotsLeft,
		PriceMonthly: pkg.PriceMonthly,
		StartedAt:    sub.StartedAt,
		ExpiresAt:    sub.ExpiresAt,
		Status:       string(sub.Status),
	}
}

// GetPaymentMethods trả các phương thức thanh toán gói đang bật. MoMo/VNPay
// chỉ hiện khi vừa bật cờ vừa có đủ cấu hình merchant, tránh hiển thị nút
// bấm vào sẽ lỗi.
func (s *paymentServiceImpl) GetPaymentMethods() dto.PaymentMethodsResponse {
	options := []dto.PaymentMethodOption{
		{Key: "wallet", Enabled: s.env.PaymentWalletEnabled},
		{Key: "momo", Enabled: s.env.PaymentMoMoEnabled && momoConfigured(s.env)},
		{Key: "vnpay", Enabled: s.env.PaymentVNPayEnabled && vnpayConfigFromEnv(s.env).enabled()},
	}
	defaultMethod := ""
	for _, option := range options {
		if option.Enabled {
			defaultMethod = option.Key
			break
		}
	}
	return dto.PaymentMethodsResponse{Methods: options, Default: defaultMethod}
}

func momoConfigured(env config.Env) bool {
	return strings.TrimSpace(env.MoMoPartnerCode) != "" &&
		strings.TrimSpace(env.MoMoAccessKey) != "" &&
		strings.TrimSpace(env.MoMoSecretKey) != "" &&
		strings.TrimSpace(env.MoMoEndpoint) != ""
}

// walletTxnRef dựng mã giao dịch ví. Có Idempotency-Key thì băm (user, gói,
// khoá) để lặp lại cùng yêu cầu ra cùng mã; không có thì sinh mã ngẫu nhiên.
func walletTxnRef(userID, packageID, idempotencyKey string) string {
	key := strings.TrimSpace(idempotencyKey)
	if key == "" {
		return "LKW" + strings.ToUpper(strings.ReplaceAll(utils.GenerateUUID(), "-", ""))[:24]
	}
	sum := sha256.Sum256([]byte(userID + "|" + packageID + "|" + key))
	return "LKW" + strings.ToUpper(hex.EncodeToString(sum[:]))[:40]
}

// ProcessVNPayCallback xử lý cả return (trình duyệt) và IPN (server-to-server).
// Hàm chỉ kích hoạt gói khi chữ ký, mã merchant, giao dịch và số tiền đều hợp lệ;
// các callback lặp lại được xử lý idempotent qua CompletePayment.
func (s *paymentServiceImpl) ProcessVNPayCallback(values url.Values) dto.PaymentCallbackResult {
	params := utils.VNPayParamsFromValues(values)
	result := dto.PaymentCallbackResult{
		TxnRef:       params["vnp_TxnRef"],
		ResponseCode: params["vnp_ResponseCode"],
	}

	cfg := vnpayConfigFromEnv(s.env)
	if !cfg.enabled() {
		result.Message = "VNPay chưa được cấu hình"
		return result
	}
	if !utils.VNPayVerify(params, cfg.hashSecret, values.Get("vnp_SecureHash")) {
		result.Message = "Chữ ký không hợp lệ"
		return result
	}
	if params["vnp_TmnCode"] != cfg.tmnCode {
		result.Message = "Mã merchant không khớp"
		return result
	}
	result.Valid = true

	txn, err := s.payments.FindByTxnRef(result.TxnRef)
	if err != nil {
		log.Printf("[VNPay] lookup txn %s failed: %v", result.TxnRef, err)
		result.Internal = true
		result.Message = "Lỗi truy vấn giao dịch"
		return result
	}
	if txn == nil {
		result.NotFound = true
		result.Message = "Không tìm thấy giao dịch"
		return result
	}

	gatewayAmount, err := strconv.ParseInt(params["vnp_Amount"], 10, 64)
	if err != nil || gatewayAmount != txn.Amount*100 {
		result.AmountMismatch = true
		result.Message = "Số tiền không khớp"
		return result
	}

	if txn.Status == models.PaymentStatusSuccess {
		result.Success = true
		result.AlreadyConfirmed = true
		result.Message = "Giao dịch đã được xác nhận"
		return result
	}

	gatewayResult := repository.PaymentResult{
		ResponseCode:  params["vnp_ResponseCode"],
		TransactionNo: params["vnp_TransactionNo"],
		BankCode:      params["vnp_BankCode"],
		PayDate:       params["vnp_PayDate"],
		PaidAt:        time.Now(),
	}

	paid := params["vnp_ResponseCode"] == "00" &&
		(params["vnp_TransactionStatus"] == "" || params["vnp_TransactionStatus"] == "00")
	if !paid {
		if _, err := s.payments.MarkFailed(txn.TxnRef, gatewayResult); err != nil {
			log.Printf("[VNPay] mark txn %s failed: %v", txn.TxnRef, err)
			result.Internal = true
			result.Message = "Không thể cập nhật giao dịch"
			return result
		}
		result.Message = "Thanh toán không thành công"
		return result
	}

	pkg, err := s.packages.GetPackageByID(txn.PackageID)
	if err != nil {
		log.Printf("[VNPay] package %s of txn %s not found: %v", txn.PackageID, txn.TxnRef, err)
		result.Internal = true
		result.Message = "Không tìm thấy gói quảng cáo"
		return result
	}

	activeSub, err := s.packages.GetActiveSubscription(txn.UserID)
	if err != nil {
		log.Printf("[VNPay] active subscription of user %s: %v", txn.UserID, err)
		result.Internal = true
		result.Message = "Không thể kiểm tra gói hiện tại"
		return result
	}

	transitioned, err := s.payments.CompletePayment(
		txn.TxnRef,
		gatewayResult,
		activeSub,
		newPartnerSubscription(txn.UserID, pkg),
	)
	if err != nil {
		log.Printf("[VNPay] complete txn %s failed: %v", txn.TxnRef, err)
		result.Internal = true
		result.Message = "Không thể kích hoạt gói"
		return result
	}

	result.Success = true
	if !transitioned {
		result.AlreadyConfirmed = true
		result.Message = "Giao dịch đã được xác nhận"
		return result
	}
	result.Message = "Thanh toán thành công"
	return result
}

func newVNPayTxnRef() string {
	return "LK" + strings.ToUpper(strings.ReplaceAll(utils.GenerateUUID(), "-", ""))[:24]
}

func (s *paymentServiceImpl) ListPayments(filter dto.PaymentListFilter) ([]models.PaymentTransaction, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	size := filter.PageSize
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	filter.Status = strings.TrimSpace(strings.ToLower(filter.Status))
	filter.Gateway = strings.TrimSpace(strings.ToLower(filter.Gateway))
	return s.payments.List(filter, size, (page-1)*size)
}
