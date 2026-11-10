package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"linkup/config"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const momoPaymentTTL = 15 * time.Minute

type momoConfig struct {
	partnerCode string
	accessKey   string
	secretKey   string
	endpoint    string
	redirectURL string
	ipnURL      string
}

func momoConfigFromEnv(env config.Env) momoConfig {
	return momoConfig{
		partnerCode: strings.TrimSpace(env.MoMoPartnerCode),
		accessKey:   strings.TrimSpace(env.MoMoAccessKey),
		secretKey:   strings.TrimSpace(env.MoMoSecretKey),
		endpoint:    strings.TrimSpace(env.MoMoEndpoint),
		redirectURL: strings.TrimSpace(env.MoMoRedirectURL),
		ipnURL:      strings.TrimSpace(env.MoMoIPNURL),
	}
}

func (c momoConfig) enabled() bool {
	return c.partnerCode != "" && c.accessKey != "" && c.secretKey != "" &&
		c.endpoint != "" && c.redirectURL != "" && c.ipnURL != ""
}

type momoCreateResponse struct {
	PartnerCode  string `json:"partnerCode"`
	OrderID      string `json:"orderId"`
	RequestID    string `json:"requestId"`
	Amount       int64  `json:"amount"`
	ResponseTime int64  `json:"responseTime"`
	Message      string `json:"message"`
	ResultCode   int    `json:"resultCode"`
	PayURL       string `json:"payUrl"`
	Deeplink     string `json:"deeplink"`
	QRCodeURL    string `json:"qrCodeUrl"`
}

type MoMoService interface {
	CreateMoMoPayment(userID, packageID string) (*dto.CreatePaymentResponse, error)
	ProcessMoMoCallback(values map[string]string) dto.PaymentCallbackResult
}

type momoServiceImpl struct {
	payments repository.PaymentRepository
	packages repository.PackageRepository
	env      config.Env
	client   *http.Client
}

func NewMoMoService(payments repository.PaymentRepository, packages repository.PackageRepository, env config.Env) MoMoService {
	return &momoServiceImpl{
		payments: payments,
		packages: packages,
		env:      env,
		client:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *momoServiceImpl) CreateMoMoPayment(userID, packageID string) (*dto.CreatePaymentResponse, error) {
	if !s.env.PaymentMoMoEnabled {
		return nil, errorsapp.New(errorsapp.ErrCodePaymentMethodOff)
	}
	cfg := momoConfigFromEnv(s.env)
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
	now := time.Now()
	expiresAt := now.Add(momoPaymentTTL)
	orderID := newMoMoOrderID()
	requestID := newMoMoRequestID()
	orderInfo := "Thanh toan goi quang cao " + pkg.Name

	txn := &models.PaymentTransaction{
		ID:        utils.GenerateUUID(),
		TxnRef:    orderID,
		RequestID: requestID,
		UserID:    userID,
		PackageID: pkg.ID,
		Amount:    amountVND,
		Status:    models.PaymentStatusPending,
		Gateway:   "momo",
		OrderInfo: orderInfo,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.payments.Create(txn); err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodePaymentCreateFailed, err)
	}

	const requestType = "captureWallet"
	rawData := utils.MoMoCreateRawData(
		cfg.accessKey,
		strconv.FormatInt(amountVND, 10),
		"",
		cfg.ipnURL,
		orderID,
		orderInfo,
		cfg.partnerCode,
		cfg.redirectURL,
		requestID,
		requestType,
	)

	payload := map[string]any{
		"partnerCode": cfg.partnerCode,
		"partnerName": "LinkUp",
		"storeId":     "LinkUp",
		"requestId":   requestID,
		"amount":      amountVND,
		"orderId":     orderID,
		"orderInfo":   orderInfo,
		"redirectUrl": cfg.redirectURL,
		"ipnUrl":      cfg.ipnURL,
		"lang":        "vi",
		"requestType": requestType,
		"autoCapture": true,
		"extraData":   "",
		"signature":   utils.MoMoSign(rawData, cfg.secretKey),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		s.markFailed(orderID, "SIGN_ERROR", "")
		return nil, errorsapp.Wrap(errorsapp.ErrCodePaymentCreateFailed, err)
	}

	req, err := http.NewRequest(http.MethodPost, cfg.endpoint, bytes.NewReader(body))
	if err != nil {
		s.markFailed(orderID, "REQUEST_ERROR", "")
		return nil, errorsapp.Wrap(errorsapp.ErrCodePaymentCreateFailed, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.markFailed(orderID, "HTTP_ERROR", "")
		return nil, errorsapp.Wrap(errorsapp.ErrCodePaymentCreateFailed, err)
	}
	defer resp.Body.Close()

	rawResp, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		s.markFailed(orderID, "READ_ERROR", "")
		return nil, errorsapp.Wrap(errorsapp.ErrCodePaymentCreateFailed, err)
	}

	var result momoCreateResponse
	if err := json.Unmarshal(rawResp, &result); err != nil {
		s.markFailed(orderID, "PARSE_ERROR", "")
		return nil, errorsapp.Wrap(errorsapp.ErrCodePaymentCreateFailed, err)
	}
	if result.ResultCode != 0 || result.PayURL == "" {
		s.markFailed(orderID, strconv.Itoa(result.ResultCode), result.Message)
		return nil, errorsapp.Wrap(errorsapp.ErrCodePaymentCreateFailed, fmt.Errorf("MoMo: %s", result.Message))
	}

	return &dto.CreatePaymentResponse{
		PaymentURL: result.PayURL,
		TxnRef:     orderID,
		Amount:     amountVND,
		ExpiresAt:  expiresAt,
	}, nil
}

func (s *momoServiceImpl) ProcessMoMoCallback(values map[string]string) dto.PaymentCallbackResult {
	result := dto.PaymentCallbackResult{
		TxnRef:       values["orderId"],
		ResponseCode: values["resultCode"],
	}

	cfg := momoConfigFromEnv(s.env)
	if !cfg.enabled() {
		result.Message = "MoMo chưa được cấu hình"
		return result
	}

	rawData := utils.MoMoCallbackRawData(
		cfg.accessKey,
		values["amount"],
		values["extraData"],
		values["message"],
		values["orderId"],
		values["orderInfo"],
		values["orderType"],
		values["partnerCode"],
		values["payType"],
		values["requestId"],
		values["responseTime"],
		values["resultCode"],
		values["transId"],
	)
	if !utils.MoMoVerify(rawData, cfg.secretKey, values["signature"]) {
		result.Message = "Chữ ký không hợp lệ"
		return result
	}
	if values["partnerCode"] != cfg.partnerCode {
		result.Message = "Mã merchant không khớp"
		return result
	}
	result.Valid = true

	txn, err := s.payments.FindByTxnRef(result.TxnRef)
	if err != nil {
		log.Printf("[MoMo] lookup txn %s failed: %v", result.TxnRef, err)
		result.Internal = true
		result.Message = "Lỗi truy vấn giao dịch"
		return result
	}
	if txn == nil || txn.Gateway != "momo" {
		result.NotFound = true
		result.Message = "Không tìm thấy giao dịch"
		return result
	}
	if txn.RequestID != "" && txn.RequestID != values["requestId"] {
		result.Message = "Request ID không khớp"
		return result
	}

	gatewayAmount, err := strconv.ParseInt(values["amount"], 10, 64)
	if err != nil || gatewayAmount != txn.Amount {
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
		ResponseCode:  values["resultCode"],
		TransactionNo: values["transId"],
		BankCode:      values["payType"],
		PayDate:       values["responseTime"],
		PaidAt:        time.Now(),
	}

	if values["resultCode"] != "0" {
		if _, err := s.payments.MarkFailed(txn.TxnRef, gatewayResult); err != nil {
			log.Printf("[MoMo] mark txn %s failed: %v", txn.TxnRef, err)
			result.Internal = true
			result.Message = "Không thể cập nhật giao dịch"
			return result
		}
		result.Message = "Thanh toán không thành công"
		return result
	}

	pkg, err := s.packages.GetPackageByID(txn.PackageID)
	if err != nil {
		log.Printf("[MoMo] package %s of txn %s not found: %v", txn.PackageID, txn.TxnRef, err)
		result.Internal = true
		result.Message = "Không tìm thấy gói quảng cáo"
		return result
	}

	activeSub, err := s.packages.GetActiveSubscription(txn.UserID)
	if err != nil {
		log.Printf("[MoMo] active subscription of user %s: %v", txn.UserID, err)
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
		log.Printf("[MoMo] complete txn %s failed: %v", txn.TxnRef, err)
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

func (s *momoServiceImpl) markFailed(orderID, responseCode, message string) {
	if _, err := s.payments.MarkFailed(orderID, repository.PaymentResult{ResponseCode: responseCode}); err != nil {
		log.Printf("[MoMo] mark txn %s failed (%s): %v", orderID, message, err)
	}
}

func newMoMoOrderID() string {
	return "LK" + strings.ToUpper(strings.ReplaceAll(utils.GenerateUUID(), "-", ""))[:24]
}

func newMoMoRequestID() string {
	return "LKR" + strings.ToUpper(strings.ReplaceAll(utils.GenerateUUID(), "-", ""))[:23]
}
