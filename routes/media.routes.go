package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterMediaRoutes(router *gin.Engine, ctrl *controllers.MediaController, env config.Env, db *gorm.DB) {
	mediaGroup := router.Group("/api/media")
	{
		mediaGroup.POST("/upload", middlewares.AuthMiddleware(env, db), ctrl.UploadMedia)
		mediaGroup.DELETE("/:id", middlewares.AuthMiddleware(env, db), ctrl.DeleteMedia)
		mediaGroup.GET("/storage", middlewares.AuthMiddleware(env, db), ctrl.GetStorageStatus)
		mediaGroup.GET("/user-media", middlewares.AuthMiddleware(env, db), ctrl.GetUserMedia)
	}
}
