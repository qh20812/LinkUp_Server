package controllers

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "linkup/dto"
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
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "email là bắt buộc"})
        return
    }

    if err := c.validation.ValidateEmail(input.Email); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    response, err := c.service.ForgotPassword(ctx.Request.Context(), input)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, response)
}

// POST /api/auth/verify-reset-token
func (c *PasswordResetController) VerifyResetToken(ctx *gin.Context) {
    var input dto.VerifyResetTokenInput
    if err := ctx.ShouldBindJSON(&input); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "token là bắt buộc"})
        return
    }

    if input.Token == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "token không được để trống"})
        return
    }

    response, err := c.service.VerifyResetToken(ctx.Request.Context(), input)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, response)
}

// POST /api/auth/reset-password
func (c *PasswordResetController) ResetPassword(ctx *gin.Context) {
    var input dto.ResetPasswordInput
    if err := ctx.ShouldBindJSON(&input); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }

    if input.Token == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "token là bắt buộc"})
        return
    }

    if err := c.validation.ValidatePassword(input.NewPassword); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    response, err := c.service.ResetPassword(ctx.Request.Context(), input)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, response)
}