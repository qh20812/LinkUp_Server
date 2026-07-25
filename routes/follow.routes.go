package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterFollowRoutes(router *gin.Engine, followController *controllers.FollowController, env config.Env, db *gorm.DB) {
	follow := router.Group("/api/follow")
	{
		follow.POST("/:userID", middlewares.AuthMiddleware(env, db), followController.FollowToggle)
		follow.GET("/stats/:userID", middlewares.AuthMiddleware(env, db), followController.GetFollowStats)
	}
}
