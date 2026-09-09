package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/services"
	"linkup/validations"
)

type PasswordResetController struct {
	service    *services.PasswordResetService
	validation *validations.AuthValidation
}

func NewPasswordResetController(service *services.PasswordResetService, validation *validations.AuthValidation) *PasswordResetController {
	return &PasswordResetController{
		service:    service,
		validation: validation,
	}
}

// POST /api/auth/forgot-password
func (c *PasswordResetController) ForgotPassword(ctx *gin.Context) {
	var input dto.ForgotPasswordInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(ctx, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := c.validation.ValidateEmail(input.Email); err != nil {
		errorsapp.Respond(ctx, http.StatusBadRequest, err)
		return
	}

	response, err := c.service.ForgotPassword(ctx.Request.Context(), input)
	if err != nil {
		errorsapp.Respond(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// POST /api/auth/verify-reset-token
func (c *PasswordResetController) VerifyResetToken(ctx *gin.Context) {
	var input dto.VerifyResetTokenInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(ctx, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if input.Token == "" {
		errorsapp.RespondError(ctx, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	response, err := c.service.VerifyResetToken(ctx.Request.Context(), input)
	if err != nil {
		errorsapp.Respond(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// POST /api/auth/reset-password
func (c *PasswordResetController) ResetPassword(ctx *gin.Context) {
	var input dto.ResetPasswordInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(ctx, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if input.Token == "" {
		errorsapp.RespondError(ctx, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := c.validation.ValidatePassword(input.NewPassword); err != nil {
		errorsapp.Respond(ctx, http.StatusBadRequest, err)
		return
	}

	response, err := c.service.ResetPassword(ctx.Request.Context(), input)
	if err != nil {
		errorsapp.Respond(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
