package controllers

import (
	"linkup/dto"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PackageController struct {
	service services.PackageService
}

func NewPackageController(service services.PackageService) *PackageController {
	return &PackageController{service: service}
}

func (ctrl *PackageController) GetPackages(c *gin.Context) {
	list, err := ctrl.service.GetPackages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (ctrl *PackageController) Subscribe(c *gin.Context) {
	var input dto.SubscribePackageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	sub, err := ctrl.service.SubscribePackage(userID, input.PackageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đăng ký gói quảng cáo thành công", "data": sub})
}

func (ctrl *PackageController) GetMySubscription(c *gin.Context) {
	userID := c.GetString("userID")
	res, err := ctrl.service.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res})
}
