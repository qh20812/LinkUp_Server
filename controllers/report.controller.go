package controllers

import (
	"linkup/dto"
	errorsapp "linkup/errors"
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
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	var input dto.CreateReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	response, err := h.reportService.CreateReport(c.Request.Context(), userID.(string), input)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *ReportController) UpdateReport(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	reportID := c.Param("id")
	if reportID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	var input dto.UpdateReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	response, err := h.reportService.UpdateReport(c.Request.Context(), userID.(string), reportID, input)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
