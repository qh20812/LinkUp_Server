package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterBlockRoutes(router *gin.Engine, blockController *controllers.BlockController, env config.Env, db *gorm.DB) {
	block := router.Group("/api/blocks")
	block.Use(middlewares.AuthMiddleware(env, db))
	{
		block.POST("", blockController.ToggleBlock)
		block.GET("", blockController.GetBlockedUsers)
	}
}
