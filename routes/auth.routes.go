package routes

import (
	"github.com/gin-gonic/gin"

	"linkup/middlewares"
	"linkup/controllers"
	"linkup/config"
)

func RegisterAuthRoutes(router *gin.Engine, authController *controllers.AuthController, env config.Env) {
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		
		protected := auth.Group("")
		protected.Use(middlewares.AuthMiddleware(env))
		{
			protected.POST("/change-password", authController.ChangePassword)
		}
		
	}
}
