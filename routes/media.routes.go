package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterMediaRoutes(router *gin.Engine, ctrl *controllers.MediaController, env config.Env) {
	mediaGroup := router.Group("/api/media")
	{
		mediaGroup.POST("/upload", middlewares.AuthMiddleware(env), ctrl.UploadMedia)
		mediaGroup.DELETE("/:id", middlewares.AuthMiddleware(env), ctrl.DeleteMedia)
		mediaGroup.GET("/storage", middlewares.AuthMiddleware(env), ctrl.GetStorageStatus)
		mediaGroup.GET("/user-media", middlewares.AuthMiddleware(env), ctrl.GetUserMedia)
	}
}
