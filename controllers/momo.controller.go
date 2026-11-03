package controllers

import (
	"fmt"
	"io"
	"linkup/config"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/services"
	"linkup/utils"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type MoMoController struct {
	service services.MoMoService
	env     config.Env
}

func NewMoMoController(service services.MoMoService, env config.Env) *MoMoController {
	return &MoMoController{service: service, env: env}
}

func (ctrl *MoMoController) CreateMoMoPayment(c *gin.Context) {
	var input dto.CreatePaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	userID := c.GetString("userID")
	res, err := ctrl.service.CreateMoMoPayment(userID, input.PackageID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tạo giao dịch MoMo thành công",
		"data":    res,
	})
}

// MoMoReturn là nơi trình duyệt quay về sau khi thanh toán.
func (ctrl *MoMoController) MoMoReturn(c *gin.Context) {
	result := ctrl.service.ProcessMoMoCallback(utils.ValuesToMap(c.Request.URL.Query()))

	status := "failed"
	switch {
	case !result.Valid:
		status = "invalid"
	case result.Success:
		status = "success"
	}

	target := strings.TrimSpace(ctrl.env.MoMoFrontendReturnURL)
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

// MoMoIPN nhận callback POST JSON từ MoMo. Luôn trả HTTP 200; resultCode 0
// nghĩa là đã xử lý xong (kể cả giao dịch thất bại), resultCode 1 để MoMo retry.
func (ctrl *MoMoController) MoMoIPN(c *gin.Context) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"resultCode": 1, "message": "Invalid body"})
		return
	}

	values, err := utils.MoMoValuesFromJSON(body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"resultCode": 1, "message": "Invalid JSON"})
		return
	}

	result := ctrl.service.ProcessMoMoCallback(values)

	code := 0
	message := "Success"
	if !result.Valid || result.Internal || result.NotFound || result.AmountMismatch {
		code = 1
		message = result.Message
	}

	c.JSON(http.StatusOK, gin.H{
		"partnerCode": ctrl.env.MoMoPartnerCode,
		"orderId":     values["orderId"],
		"requestId":   values["requestId"],
		"resultCode":  code,
		"message":     message,
	})
}
