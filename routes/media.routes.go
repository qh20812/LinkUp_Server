package routes

import (
	"linkup/config"
	"linkup/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterMediaRoutes(router *gin.Engine, ctrl *controllers.MediaController, env config.Env) {
	mediaGroup := router.Group("/api/media")
	{
		mediaGroup.POST("/upload", AuthMiddleware(env), ctrl.UploadMedia)
		mediaGroup.GET("/storage", AuthMiddleware(env), ctrl.GetStorageStatus)
	}
}
