package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"linkup/controllers"
	"linkup/middlewares"
	"linkup/config"
)

func RegisterNotificationRoutes(router *gin.Engine, ctrl *controllers.NotificationController, env config.Env, db *gorm.DB) {
	g := router.Group("/api/notifications")
	g.Use(middlewares.AuthMiddleware(env, db))
	{
		g.GET("", ctrl.GetNotifications)
		g.PUT("/:id/read", ctrl.MarkAsRead)
		g.PUT("/read-all", ctrl.MarkAllAsRead)
		g.GET("/unread-count", ctrl.GetUnreadCount)
		g.GET("/preferences", ctrl.GetPreferences)
		g.PUT("/preferences", ctrl.UpdatePreferences)
	}
}
