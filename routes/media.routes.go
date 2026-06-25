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
		mediaGroup.GET("/storage", middlewares.AuthMiddleware(env), ctrl.GetStorageStatus)
<<<<<<< HEAD
		mediaGroup.GET("/user-media", middlewares.AuthMiddleware(env), ctrl.GetUserMedia)
=======
>>>>>>> 13551be0d51ec24a37c47a4610cf0d802911d6a6
	}
}
