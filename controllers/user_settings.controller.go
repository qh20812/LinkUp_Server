package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"linkup/dto"
	"linkup/services"
)

type UserSettingsController struct {
	service *services.UserSettingsService
}

func NewUserSettingsController(service *services.UserSettingsService) *UserSettingsController {
	return &UserSettingsController{service: service}
}

func (h *UserSettingsController) GetPrivacy(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "người dùng chưa xác thực"})
		return
	}

	privacy, err := h.service.GetPrivacy(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, privacy)
}

func (h *UserSettingsController) UpdatePrivacy(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "người dùng chưa xác thực"})
		return
	}

	var input dto.UpdatePrivacyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}

	if input.DiscoverableInSearch == nil && input.AllowStrangerMessages == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cần ít nhất một trường để cập nhật"})
		return
	}

	privacy, err := h.service.UpdatePrivacy(c.Request.Context(), userID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, privacy)
}

func (h *UserSettingsController) GetStorage(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "người dùng chưa xác thực"})
		return
	}

	storage, err := h.service.GetStorage(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, storage)
}

func (h *UserSettingsController) Deactivate(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "người dùng chưa xác thực"})
		return
	}

	var input dto.DeactivateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}

	if err := h.service.Deactivate(c.Request.Context(), userID, input.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tài khoản đã được vô hiệu hóa tạm thời"})
}

func (h *UserSettingsController) GetAppearance(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "người dùng chưa xác thực"})
		return
	}

	appearance, err := h.service.GetAppearance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, appearance)
}

func (h *UserSettingsController) UpdateAppearance(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "người dùng chưa xác thực"})
		return
	}

	var input dto.UpdateAppearanceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}

	if input.Theme == nil && input.Language == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cần ít nhất một trường để cập nhật"})
		return
	}

	appearance, err := h.service.UpdateAppearance(c.Request.Context(), userID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, appearance)
}

func (h *UserSettingsController) ListSessions(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "người dùng chưa xác thực"})
		return
	}

	sessions, err := h.service.ListSessions(c.Request.Context(), userID, c.GetString("sessionID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

func (h *UserSettingsController) RevokeSession(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "người dùng chưa xác thực"})
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id là bắt buộc"})
		return
	}

	if err := h.service.RevokeSession(c.Request.Context(), userID, sessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã thu hồi phiên đăng nhập"})
}

func (h *UserSettingsController) RevokeOtherSessions(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "người dùng chưa xác thực"})
		return
	}

	if err := h.service.RevokeOtherSessions(c.Request.Context(), userID, c.GetString("sessionID")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã đăng xuất các thiết bị khác"})
}
