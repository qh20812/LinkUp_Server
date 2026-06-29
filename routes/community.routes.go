package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterCommunityRoutes(router *gin.Engine, ctrl *controllers.CommunityController, env config.Env) {
	communityGroup := router.Group("/api/communities")
	communityGroup.Use(middlewares.AuthMiddleware(env))
	{
		communityGroup.POST("", ctrl.CreateCommunity)
	}
}
