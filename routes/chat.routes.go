package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterChatRoutes(router *gin.Engine, ctrl *controllers.ChatController, env config.Env, db *gorm.DB) {
	chatGroup := router.Group("/api/chats")
	chatGroup.Use(middlewares.AuthMiddleware(env, db))
	{
		chatGroup.POST("/direct", ctrl.CreateDirectChat)
		chatGroup.POST("/invite", ctrl.CreateChatInvite)
		chatGroup.POST("/invite/respond", ctrl.ResponseChatInvite)

		chatGroup.GET("/messages/:messageID/download", ctrl.DownloadMessageMedia)
		chatGroup.DELETE("/:chatID", ctrl.DeleteChat)

		chatGroup.GET("/ws", ctrl.HandleWebsocket)
	}
}
