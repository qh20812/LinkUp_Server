package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterChatRoutes(router *gin.Engine, ctrl *controllers.ChatController, env config.Env) {
	chatGroup := router.Group("/api/chats")
	chatGroup.Use(middlewares.AuthMiddleware(env))
	{
		chatGroup.POST("/direct", ctrl.CreateDirectChat)
		chatGroup.POST("/invite", ctrl.CreateChatInvite)
		chatGroup.POST("/invite/respond", ctrl.ResponseChatInvite)

		chatGroup.GET("/messages/:messageID/download", ctrl.DownloadMessageMedia)

		chatGroup.GET("/ws", ctrl.HandleWebsocket)
	}
}
