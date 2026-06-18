package routes

import (
	"github.com/gin-gonic/gin"

	"linkup/controllers"
)

func RegisterAuthRoutes(router *gin.Engine, authController *controllers.AuthController) {
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
	}
}
