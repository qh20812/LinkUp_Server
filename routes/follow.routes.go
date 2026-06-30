package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterFollowRoutes(router *gin.Engine, followController *controllers.FollowController, env config.Env) {
	follow := router.Group("/api/follow")
	{
		follow.POST("/:userID", middlewares.AuthMiddleware(env), followController.FollowToggle)
		follow.GET("/stats/:userID", middlewares.AuthMiddleware(env), followController.GetFollowStats)
	}
}
