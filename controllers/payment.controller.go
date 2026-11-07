package controllers

import (
	"fmt"
	"linkup/config"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/services"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	service services.PaymentService
	env     config.Env
}

func NewPaymentController(service services.PaymentService, env config.Env) *PaymentController {
	return &PaymentController{service: service, env: env}
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
