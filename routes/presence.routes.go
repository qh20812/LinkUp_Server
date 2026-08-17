package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"
)

func RegisterPresenceRoutes(router *gin.Engine, ctrl *controllers.PresenceController, env config.Env, db *gorm.DB) {
	g := router.Group("/api/presence")
	g.Use(middlewares.AuthMiddleware(env, db))
	{
		g.GET("/online/:userID", ctrl.GetPresence)
		g.POST("/batch", ctrl.BatchGetPresence)
		g.GET("/online", ctrl.GetOnlineUsers)
		g.GET("/count", ctrl.GetOnlineCount)
		g.GET("/settings", ctrl.GetPresenceSettings)
		g.PUT("/settings", ctrl.UpdatePresenceSettings)
	}
}
