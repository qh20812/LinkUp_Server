package controllers

import (
	"linkup/dto"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "không có quyền truy cập"})
		return
	}

	result, err := ctrl.service.GetSettings(c.Request.Context(), adminID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminSettingsController) UpdateSettings(c *gin.Context) {
	adminID := c.GetString("userID")
	if adminID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "không có quyền truy cập"})
		return
	}

	var input dto.AdminSettingsUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu không hợp lệ"})
		return
	}

	if err := ctrl.service.UpdateSettings(c.Request.Context(), adminID, &input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lưu cài đặt thành công"})
}