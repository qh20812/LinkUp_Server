package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"linkup/middlewares"
	"linkup/controllers"
	"linkup/config"
)

func RegisterAuthRoutes(router *gin.Engine, authController *controllers.AuthController, emailVerifyController *controllers.EmailVerificationController, env config.Env, db *gorm.DB) {
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/google", authController.GoogleLogin)
		auth.POST("/refresh", authController.RefreshToken)
		auth.POST("/verify-email", emailVerifyController.VerifyEmail)
		auth.POST("/resend-verification", emailVerifyController.ResendVerification)
		
		protected := auth.Group("")
		protected.Use(middlewares.AuthMiddleware(env, db))
		{
			protected.POST("/change-password", authController.ChangePassword)
			protected.POST("/logout", authController.Logout)
		}
		
	}
}
