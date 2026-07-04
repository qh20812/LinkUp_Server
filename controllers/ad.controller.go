package controllers

import (
	"linkup/dto"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdController struct {
	service services.AdService
}

func NewAdController(service services.AdService) *AdController {
	return &AdController{service: service}
}

func (ctrl *AdController) CreateAd(c *gin.Context) {
	var input dto.CreateAdInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized context"})
		return
	}

	ad, err := ctrl.service.CreateAd(input, currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": ad})
}

func (ctrl *AdController) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var input dto.UpdateAdStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ad, err := ctrl.service.UpdateStatus(id, input.Status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully", "data": ad})
}

func (ctrl *AdController) GetAnalytics(c *gin.Context) {
	id := c.Param("id")
	report, err := ctrl.service.GetAdPerformance(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
