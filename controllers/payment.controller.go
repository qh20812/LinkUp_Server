package controllers

import (
	"fmt"
	"linkup/config"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/services"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	service services.PaymentService
	env     config.Env
	issuer  TokenIssuer
}

func NewPaymentController(service services.PaymentService, env config.Env) *PaymentController {
	return &PaymentController{service: service, env: env}
}

// SetTokenIssuer gắn bộ cấp JWT để trả token mới khi mua gói nâng role.
func (ctrl *PaymentController) SetTokenIssuer(issuer TokenIssuer) {
	ctrl.issuer = issuer
}

func (ctrl *PaymentController) ListMyPayments(c *gin.Context) {
	var filter dto.PaymentListFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}
	filter.UserID = c.GetString("userID")
	ctrl.respondPaymentList(c, filter)
}

func (ctrl *PaymentController) ListAllPayments(c *gin.Context) {
	var filter dto.PaymentListFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}
	ctrl.respondPaymentList(c, filter)
}

func (ctrl *PaymentController) respondPaymentList(c *gin.Context, filter dto.PaymentListFilter) {
	rows, total, err := ctrl.service.ListPayments(filter)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      rows,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (ctrl *PaymentController) CreateVNPayPayment(c *gin.Context) {
	var input dto.CreatePaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	userID := c.GetString("userID")
	res, err := ctrl.service.CreateVNPayPayment(userID, input.PackageID, c.ClientIP())
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tạo giao dịch VNPay thành công",
		"data":    res,
	})
}

// CreateWalletPayment mua gói bằng ví trả trước. Header Idempotency-Key (tuỳ
// chọn) giúp retry không bị trừ tiền hai lần.
func (ctrl *PaymentController) CreateWalletPayment(c *gin.Context) {
	var input dto.CreatePaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	userID := c.GetString("userID")
	res, err := ctrl.service.CreateWalletPayment(
		userID,
		input.PackageID,
		c.GetHeader("Idempotency-Key"),
	)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}

	body := gin.H{"message": "Thanh toán gói bằng ví thành công", "data": res}
	if !res.Replayed {
		// Role DB có thể vừa nâng lên PARTNER — cấp lại token để claim khớp ngay.
		// Best-effort: lỗi cấp token không làm fail giao dịch đã trừ tiền.
		if ctrl.issuer != nil {
			sessionID := c.GetString("sessionID")
			if tokens, terr := ctrl.issuer.ReissueTokensForUser(c.Request.Context(), userID, sessionID); terr != nil {
				log.Printf("[WalletPayment] reissue tokens failed for user %s: %v", userID, terr)
			} else if tokens != nil {
				body["tokens"] = tokens
			}
		}
	}
	c.JSON(http.StatusOK, body)
}

// GetPaymentMethods trả phương thức thanh toán gói đang bật (ví/momo/vnpay)
// để UI chỉ hiện lựa chọn khả dụng.
func (ctrl *PaymentController) GetPaymentMethods(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": ctrl.service.GetPaymentMethods()})
}

// VNPayReturn là nơi trình duyệt quay về sau khi thanh toán. Endpoint công khai
// nên chỉ dùng để hiển thị kết quả; việc kích hoạt gói đã được xác thực bằng
// chữ ký và số tiền giống hệt IPN.
func (ctrl *PaymentController) VNPayReturn(c *gin.Context) {
	result := ctrl.service.ProcessVNPayCallback(c.Request.URL.Query())

	status := "failed"
	switch {
	case !result.Valid:
		status = "invalid"
	case result.Success:
		status = "success"
	}

	target := strings.TrimSpace(ctrl.env.VNPayFrontendReturnURL)
	if target == "" {
		target = "/packages"
	}
	separator := "?"
	if strings.Contains(target, "?") {
		separator = "&"
	}
	redirect := fmt.Sprintf("%s%spayment=%s", target, separator, url.QueryEscape(status))
	if result.TxnRef != "" {
		redirect += "&txn_ref=" + url.QueryEscape(result.TxnRef)
	}
	c.Redirect(http.StatusFound, redirect)
}

// VNPayIPN là callback server-to-server. Luôn trả HTTP 200 kèm RspCode theo
// đặc tả VNPay; mã khác 00 khiến VNPay gửi lại.
func (ctrl *PaymentController) VNPayIPN(c *gin.Context) {
	result := ctrl.service.ProcessVNPayCallback(c.Request.URL.Query())

	code, message := "00", "Confirm Success"
	switch {
	case result.Internal:
		code, message = "99", "Unknown error"
	case !result.Valid:
		code, message = "97", "Invalid Checksum"
	case result.NotFound:
		code, message = "01", "Order not found"
	case result.AmountMismatch:
		code, message = "04", "Invalid amount"
	case result.AlreadyConfirmed:
		code, message = "02", "Order already confirmed"
	}

	c.JSON(http.StatusOK, gin.H{"RspCode": code, "Message": message})
}
