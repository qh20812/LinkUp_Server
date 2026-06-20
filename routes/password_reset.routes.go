package routes

import (
	"github.com/gin-gonic/gin"

	"linkup/controllers"
)

func RegisterPasswordResetRoutes(router *gin.Engine, controller *controllers.PasswordResetController) {
	auth := router.Group("/api/auth")
	{
		auth.POST("/forgot-password", controller.ForgotPassword)
		auth.POST("/verify-reset-token", controller.VerifyResetToken)
		auth.POST("/reset-password", controller.ResetPassword)
	}
}
