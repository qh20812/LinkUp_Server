package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/services"

	"github.com/gin-gonic/gin"
)

// WalletController phục vụ ví trả trước và nạp tiền bằng chuyển khoản.
type WalletController struct {
	service services.WalletService
}

func NewWalletController(service services.WalletService) *WalletController {
	return &WalletController{service: service}
}

// respondWalletError trả lỗi có mã + HTTP status phù hợp.
func respondWalletError(c *gin.Context, err error) {
	if appErr, ok := errorsapp.IsAppError(err); ok {
		errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		return
	}
	errorsapp.Respond(c, http.StatusInternalServerError, err)
}

func queryInt(c *gin.Context, key string, fallback int) int {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func (ctrl *WalletController) GetWallet(c *gin.Context) {
	userID := c.GetString("userID")
	res, err := ctrl.service.GetWallet(c.Request.Context(), userID)
	if err != nil {
		respondWalletError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (ctrl *WalletController) ListTransactions(c *gin.Context) {
	userID := c.GetString("userID")
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)
	res, err := ctrl.service.ListTransactions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		respondWalletError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctrl *WalletController) ListTopups(c *gin.Context) {
	userID := c.GetString("userID")
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)
	res, err := ctrl.service.ListTopups(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		respondWalletError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctrl *WalletController) CreateTopup(c *gin.Context) {
	var input dto.CreateTopupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}
	userID := c.GetString("userID")
	res, err := ctrl.service.CreateTopup(c.Request.Context(), userID, input.Amount)
	if err != nil {
		respondWalletError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Đã tạo yêu cầu nạp tiền", "data": res})
}

func (ctrl *WalletController) ReportTopup(c *gin.Context) {
	var input dto.ReportTopupInput
	// Body là tuỳ chọn (không có ảnh biên lai vẫn báo được).
	_ = c.ShouldBindJSON(&input)

	userID := c.GetString("userID")
	res, err := ctrl.service.ReportTopup(c.Request.Context(), userID, c.Param("id"), input.ProofURL)
	if err != nil {
		respondWalletError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Đã ghi nhận bạn chuyển khoản. Quản trị viên sẽ đối chiếu và cộng ví.",
		"data":    res,
	})
}

func (ctrl *WalletController) CancelTopup(c *gin.Context) {
	userID := c.GetString("userID")
	if err := ctrl.service.CancelTopup(c.Request.Context(), userID, c.Param("id")); err != nil {
		respondWalletError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã huỷ yêu cầu nạp tiền"})
}

// ── Admin ────────────────────────────────────────────────────────────

func (ctrl *WalletController) AdminListTopups(c *gin.Context) {
	var filter dto.AdminTopupFilterInput
	if err := c.ShouldBindQuery(&filter); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}
	res, err := ctrl.service.AdminListTopups(c.Request.Context(), filter)
	if err != nil {
		respondWalletError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctrl *WalletController) AdminApproveTopup(c *gin.Context) {
	var input dto.ApproveTopupInput
	_ = c.ShouldBindJSON(&input)

	adminID := c.GetString("userID")
	res, err := ctrl.service.ApproveTopup(c.Request.Context(), adminID, c.Param("id"), input)
	if err != nil {
		respondWalletError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã duyệt và cộng tiền vào ví", "data": res})
}

func (ctrl *WalletController) AdminRejectTopup(c *gin.Context) {
	var input dto.RejectTopupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	adminID := c.GetString("userID")
	res, err := ctrl.service.RejectTopup(c.Request.Context(), adminID, c.Param("id"), input)
	if err != nil {
		respondWalletError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã từ chối yêu cầu nạp tiền", "data": res})
}

func (ctrl *WalletController) AdminListWallets(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)
	res, err := ctrl.service.AdminListWallets(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		respondWalletError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}
