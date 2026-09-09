package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"linkup/dto"
	errorsapp "linkup/errors"
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
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	privacy, err := h.service.GetPrivacy(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, privacy)
}

func (h *UserSettingsController) UpdatePrivacy(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	var input dto.UpdatePrivacyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if input.DiscoverableInSearch == nil && input.AllowStrangerMessages == nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	privacy, err := h.service.UpdatePrivacy(c.Request.Context(), userID, input)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, privacy)
}

func (h *UserSettingsController) GetStorage(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	storage, err := h.service.GetStorage(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, storage)
}

func (h *UserSettingsController) Deactivate(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	var input dto.DeactivateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := h.service.Deactivate(c.Request.Context(), userID, input.Password); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tài khoản đã được vô hiệu hóa tạm thời"})
}

func (h *UserSettingsController) GetAppearance(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	appearance, err := h.service.GetAppearance(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, appearance)
}

func (h *UserSettingsController) UpdateAppearance(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	var input dto.UpdateAppearanceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if input.Theme == nil && input.Language == nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	appearance, err := h.service.UpdateAppearance(c.Request.Context(), userID, input)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, appearance)
}

func (h *UserSettingsController) ListSessions(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	sessions, err := h.service.ListSessions(c.Request.Context(), userID, c.GetString("sessionID"))
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

func (h *UserSettingsController) RevokeSession(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := h.service.RevokeSession(c.Request.Context(), userID, sessionID); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã thu hồi phiên đăng nhập"})
}

func (h *UserSettingsController) RevokeOtherSessions(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	if err := h.service.RevokeOtherSessions(c.Request.Context(), userID, c.GetString("sessionID")); err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã đăng xuất các thiết bị khác"})
}
