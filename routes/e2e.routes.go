package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterE2ERoutes(router *gin.Engine, ctrl *controllers.E2EController, env config.Env, db *gorm.DB) {
	e2eGroup := router.Group("/api/e2e")
	e2eGroup.Use(middlewares.AuthMiddleware(env, db))
	{
		e2eGroup.PUT("/keys", ctrl.RegisterUserKey)
		e2eGroup.GET("/keys/:userID", ctrl.GetUserKey)
		e2eGroup.POST("/chats/keys", ctrl.StoreChatKeys)
		e2eGroup.GET("/chats/:chatID/keys", ctrl.GetChatKey)
	}
}
