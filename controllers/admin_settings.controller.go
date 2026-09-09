package controllers

import (
	"linkup/dto"
	"linkup/errors"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminSettingsController struct {
	service *services.AdminSettingsService
}

func NewAdminSettingsController(service *services.AdminSettingsService) *AdminSettingsController {
	return &AdminSettingsController{service: service}
}

func (ctrl *AdminSettingsController) GetSettings(c *gin.Context) {
	adminID := c.GetString("userID")
	if adminID == "" {
		errors.RespondError(c, http.StatusUnauthorized, errors.New(errors.ErrCodeAdminNoAccess))
		return
	}

	result, err := ctrl.service.GetSettings(c.Request.Context(), adminID)
	if err != nil {
		errors.Respond(c, http.StatusForbidden, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminSettingsController) UpdateSettings(c *gin.Context) {
	adminID := c.GetString("userID")
	if adminID == "" {
		errors.RespondError(c, http.StatusUnauthorized, errors.New(errors.ErrCodeAdminNoAccess))
		return
	}

	var input dto.AdminSettingsUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errors.RespondError(c, http.StatusBadRequest, errors.New(errors.ErrCodeAdminInvalidInput))
		return
	}

	if err := ctrl.service.UpdateSettings(c.Request.Context(), adminID, &input); err != nil {
		errors.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lưu cài đặt thành công"})
}
