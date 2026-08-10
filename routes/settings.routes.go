package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"
)

func RegisterSettingsRoutes(router *gin.Engine, ctrl *controllers.UserSettingsController, env config.Env, db *gorm.DB) {
	g := router.Group("/api/settings")
	g.Use(middlewares.AuthMiddleware(env, db))
	{
		g.GET("/privacy", ctrl.GetPrivacy)
		g.PUT("/privacy", ctrl.UpdatePrivacy)
		g.GET("/storage", ctrl.GetStorage)
		g.POST("/deactivate", ctrl.Deactivate)
		g.GET("/appearance", ctrl.GetAppearance)
		g.PUT("/appearance", ctrl.UpdateAppearance)
		g.GET("/sessions", ctrl.ListSessions)
		g.DELETE("/sessions/:id", ctrl.RevokeSession)
		g.POST("/sessions/revoke-others", ctrl.RevokeOtherSessions)
	}
}
