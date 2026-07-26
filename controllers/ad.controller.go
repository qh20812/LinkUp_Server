package controllers

import (
	"linkup/dto"
	"linkup/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AdController struct {
	service services.AdService
}

func NewAdController(service services.AdService) *AdController {
	return &AdController{service: service}
}

func (ctrl *AdController) CreateAd(c *gin.Context) {
	err := c.Request.ParseMultipartForm(500 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không thể parse dữ liệu form upload"})
		return
	}

	title := c.PostForm("title")
	content := c.PostForm("content")
	targetURL := c.PostForm("target_url")
	format := c.PostForm("format")
	if format == "" {
		format = "image"
	}

	budget, _ := strconv.ParseFloat(c.PostForm("budget"), 64)
	dailyBudget, _ := strconv.ParseFloat(c.PostForm("daily_budget"), 64)
	cpmPrice, _ := strconv.ParseFloat(c.PostForm("cpm_price"), 64)
	cpcPrice, _ := strconv.ParseFloat(c.PostForm("cpc_price"), 64)
	maxImpressions, _ := strconv.Atoi(c.PostForm("max_impressions"))

	var startedAt, expiresAt *time.Time
	if val := c.PostForm("started_at"); val != "" {
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			startedAt = &t
		}
	}
	if val := c.PostForm("expires_at"); val != "" {
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			expiresAt = &t
		}
	}

	input := dto.CreateAdInput{
		Title:          title,
		Content:        content,
		TargetURL:      targetURL,
		Format:         format,
		Budget:         budget,
		DailyBudget:    dailyBudget,
		CPMPrice:       cpmPrice,
		CPCPrice:       cpcPrice,
		MaxImpressions: maxImpressions,
		StartedAt:      startedAt,
		ExpiresAt:      expiresAt,
	}

	form := c.Request.MultipartForm
	files := form.File["media"]

	currentUserID := c.GetString("userID")
	ad, err := ctrl.service.CreateAdWithMedia(c, input, files, currentUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Tạo quảng cáo thành công", "data": ad})
}

func (ctrl *AdController) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	partnerID := c.GetString("userID")

	var input dto.UpdateAdStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ad, err := ctrl.service.UpdateStatus(id, input.Status, partnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully", "data": ad})
}

func (ctrl *AdController) GetAnalytics(c *gin.Context) {
	id := c.Param("id")
	partnerID := c.GetString("userID")

	report, err := ctrl.service.GetAdPerformance(id, partnerID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": report})
}

func (ctrl *AdController) GetAdminList(c *gin.Context) {
	list, err := ctrl.service.GetDashboardList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (ctrl *AdController) GetUserFeed(c *gin.Context) {
	ads, err := ctrl.service.GetAdsForUserFeed()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ads})
}

func (ctrl *AdController) TrackAction(c *gin.Context) {
	adID := c.Param("id")
	var input dto.TrackActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userIDPtr *string
	if uid, exists := c.Get("userID"); exists {
		if strUID, ok := uid.(string); ok {
			userIDPtr = &strUID
		}
	}

	ipAddress := c.ClientIP()
	err := ctrl.service.TrackUserAction(adID, userIDPtr, input.ActionType, ipAddress)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}
