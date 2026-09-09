package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"linkup/dto"
	errorsapp "linkup/errors"
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
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	response, err := ctrl.service.VerifyEmail(c.Request.Context(), input.Token)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
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
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := ctrl.validation.ValidateEmail(input.Email); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	response, err := ctrl.service.ResendVerification(c.Request.Context(), input.Email)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
