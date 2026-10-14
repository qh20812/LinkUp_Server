package controllers

import (
	"encoding/json"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/services"
	"linkup/validations"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AdController struct {
	service     services.AdService
	profileRepo *repository.ProfileRepository
}

func NewAdController(service services.AdService, profileRepo *repository.ProfileRepository) *AdController {
	return &AdController{service: service, profileRepo: profileRepo}
}

func (ctrl *AdController) CreateAd(c *gin.Context) {
	err := c.Request.ParseMultipartForm(500 << 20)
	if err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
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

	if err := validations.ValidateCreateAd(title, content, targetURL, format, budget); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

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

	targetGender := c.PostForm("target_gender")
	targetAgeMin, _ := strconv.Atoi(c.PostForm("target_age_min"))
	targetAgeMax, _ := strconv.Atoi(c.PostForm("target_age_max"))
	var targetLocations []string
	if locStr := c.PostForm("target_locations"); locStr != "" {
		_ = json.Unmarshal([]byte(locStr), &targetLocations)
	}

	if err := validations.ValidateTargeting(targetGender, targetAgeMin, targetAgeMax); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	input := dto.CreateAdInput{
		Title:           title,
		Content:         content,
		TargetURL:       targetURL,
		Format:          format,
		Budget:          budget,
		DailyBudget:     dailyBudget,
		CPMPrice:        cpmPrice,
		CPCPrice:        cpcPrice,
		MaxImpressions:  maxImpressions,
		StartedAt:       startedAt,
		ExpiresAt:       expiresAt,
		TargetGender:    targetGender,
		TargetAgeMin:    targetAgeMin,
		TargetAgeMax:    targetAgeMax,
		TargetLocations: targetLocations,
	}

	form := c.Request.MultipartForm
	files := form.File["media"]

	currentUserID := c.GetString("userID")
	ad, err := ctrl.service.CreateAdWithMedia(c, input, files, currentUserID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}

	// 2.3 Moderation: Partner/Admin tạo ad → pending; SUPER_ADMIN → active
	role := c.GetString("userRole")
	if role != string(models.RoleSuperAdmin) {
		_ = ctrl.service.SetAdStatus(ad.ID, models.AdStatusPending)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Tạo quảng cáo thành công", "data": ad})
}

func (ctrl *AdController) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	partnerID := c.GetString("userID")

	var input dto.UpdateAdStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	ad, err := ctrl.service.UpdateStatus(id, input.Status, partnerID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully", "data": ad})
}

func (ctrl *AdController) GetAnalytics(c *gin.Context) {
	id := c.Param("id")
	partnerID := c.GetString("userID")

	report, err := ctrl.service.GetAdPerformance(id, partnerID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
		errorsapp.Respond(c, http.StatusForbidden, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": report})
}

func (ctrl *AdController) GetAdminList(c *gin.Context) {
	var input dto.AdListInput
	if err := c.ShouldBindQuery(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	role := c.GetString("userRole")
	currentUserID := c.GetString("userID")

	if role == string(models.RoleSuperAdmin) || role == string(models.RoleAdmin) {
		list, err := ctrl.service.GetDashboardList()
		if err != nil {
			errorsapp.Respond(c, http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": list})
		return
	}

	ads, total, err := ctrl.service.GetPartnerAds(currentUserID, input.Page, input.PageSize)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      ads,
		"total":     total,
		"page":      input.Page,
		"page_size": input.PageSize,
	})
}

func (ctrl *AdController) GetUserFeed(c *gin.Context) {
	userID := c.GetString("userID")

	gender, age, location := "all", 0, ""
	if userID != "" && ctrl.profileRepo != nil {
		demo, err := ctrl.profileRepo.GetUserDemographics(userID)
		if err == nil && demo != nil {
			gender = demo.Gender
			age = demo.Age
			location = demo.Location
		}
	}

	ads, err := ctrl.service.GetAdsForUserFeed(userID, gender, age, location)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ads})
}

func (ctrl *AdController) TrackAction(c *gin.Context) {
	adID := c.Param("id")
	var input dto.TrackActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
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
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
		errorsapp.Respond(c, http.StatusNotFound, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

func (ctrl *AdController) UpdateAd(c *gin.Context) {
	id := c.Param("id")
	partnerID := c.GetString("userID")

	if err := c.Request.ParseMultipartForm(500 << 20); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

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

	budget, _ := strconv.ParseFloat(c.PostForm("budget"), 64)
	dailyBudget, _ := strconv.ParseFloat(c.PostForm("daily_budget"), 64)
	cpmPrice, _ := strconv.ParseFloat(c.PostForm("cpm_price"), 64)
	cpcPrice, _ := strconv.ParseFloat(c.PostForm("cpc_price"), 64)
	maxImpressions, _ := strconv.Atoi(c.PostForm("max_impressions"))
	targetAgeMin, _ := strconv.Atoi(c.PostForm("target_age_min"))
	targetAgeMax, _ := strconv.Atoi(c.PostForm("target_age_max"))
	targetGender := c.PostForm("target_gender")
	var targetLocations []string
	if locStr := c.PostForm("target_locations"); locStr != "" {
		_ = json.Unmarshal([]byte(locStr), &targetLocations)
	}

	input := dto.UpdateAdInput{
		Title:           c.PostForm("title"),
		Content:         c.PostForm("content"),
		TargetURL:       c.PostForm("target_url"),
		Budget:          budget,
		DailyBudget:     dailyBudget,
		CPMPrice:        cpmPrice,
		CPCPrice:        cpcPrice,
		MaxImpressions:  maxImpressions,
		StartedAt:       startedAt,
		ExpiresAt:       expiresAt,
		TargetGender:    targetGender,
		TargetAgeMin:    targetAgeMin,
		TargetAgeMax:    targetAgeMax,
		TargetLocations: targetLocations,
	}

	if err := validations.ValidateCreateAd(input.Title, input.Content, input.TargetURL, "image", input.Budget); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	ad, err := ctrl.service.UpdateAd(id, partnerID, input)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật quảng cáo thành công", "data": ad})
}

func (ctrl *AdController) DeleteAd(c *gin.Context) {
	id := c.Param("id")
	partnerID := c.GetString("userID")

	if err := ctrl.service.DeleteAd(id, partnerID); err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Xóa quảng cáo thành công"})
}

func (ctrl *AdController) GetOverview(c *gin.Context) {
	partnerID := c.GetString("userID")

	overview, err := ctrl.service.GetOverview(partnerID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": overview})
}
