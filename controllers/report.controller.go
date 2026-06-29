package controllers

import (
	"linkup/dto"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReportController struct {
	reportService *services.ReportService
}

func NewReportController(reportService *services.ReportService) *ReportController {
	return &ReportController{
		reportService: reportService,
	}
}

func (h *ReportController) CreateReport(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "người dùng chưa xác thực"})
		return
	}

	var input dto.CreateReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}

	response, err := h.reportService.CreateReport(c.Request.Context(), userID.(string), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}
