package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(router *gin.Engine, adminController *controllers.AdminController, env config.Env) {
	adminGroup := router.Group("/api/admin")
	adminGroup.Use(middlewares.AuthMiddleware(env))

	adminGroup.GET("/users", adminController.ListUsers)
	adminGroup.PUT("/users/:userID/status", adminController.UpdateUserStatus)
	adminGroup.POST("/users/:userID/ban", adminController.BanUser)
}
