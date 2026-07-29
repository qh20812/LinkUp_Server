package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"linkup/dto"
	"linkup/services"
	"linkup/validations"
)

type EmailVerificationController struct {
	service    *services.EmailVerificationService
	validation *validations.AuthValidation
}

func NewEmailVerificationController(service *services.EmailVerificationService, validation *validations.AuthValidation) *EmailVerificationController {
	return &EmailVerificationController{
		service:    service,
		validation: validation,
	}
}

func (ctrl *EmailVerificationController) VerifyEmail(c *gin.Context) {
	var input dto.VerifyEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token là bắt buộc"})
		return
	}

	response, err := ctrl.service.VerifyEmail(c.Request.Context(), input.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !response.Verified {
		c.JSON(http.StatusBadRequest, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *EmailVerificationController) ResendVerification(c *gin.Context) {
	var input dto.ResendVerificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email là bắt buộc"})
		return
	}

	if err := ctrl.validation.ValidateEmail(input.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := ctrl.service.ResendVerification(c.Request.Context(), input.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
