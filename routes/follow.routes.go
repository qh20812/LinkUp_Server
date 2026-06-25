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
<<<<<<< HEAD
		follow.GET("/stats/:userID", middlewares.AuthMiddleware(env),followController.GetFollowStats)
=======
		follow.GET("/stats/:userID", middlewares.AuthMiddleware(env), followController.GetFollowStats)
>>>>>>> 13551be0d51ec24a37c47a4610cf0d802911d6a6
	}
}
