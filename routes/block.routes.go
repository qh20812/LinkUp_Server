package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterBlockRoutes(router *gin.Engine, blockController *controllers.BlockController, env config.Env) {
	block := router.Group("/api/blocks")
	block.Use(middlewares.AuthMiddleware(env))
	{
		block.POST("", blockController.ToggleBlock)
		block.GET("", blockController.GetBlockedUsers)
	}
}
