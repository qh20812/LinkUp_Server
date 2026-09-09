package controllers

import (
	"linkup/dto"
	errorsapp "linkup/errors"
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
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (ctrl *PackageController) Subscribe(c *gin.Context) {
	var input dto.SubscribePackageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	userID := c.GetString("userID")
	sub, err := ctrl.service.SubscribePackage(userID, input.PackageID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đăng ký gói quảng cáo thành công", "data": sub})
}

func (ctrl *PackageController) GetMySubscription(c *gin.Context) {
	userID := c.GetString("userID")
	res, err := ctrl.service.GetUserSubscription(userID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusNotFound, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res})
}
