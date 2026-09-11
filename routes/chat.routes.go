package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterChatRoutes(router *gin.Engine, ctrl *controllers.ChatController, env config.Env, db *gorm.DB) {
	// WS dùng ?token= để trình duyệt có thể kết nối (WebSocket không đặt được header).
	router.GET("/api/chats/ws", ctrl.HandleWebsocket)

	chatGroup := router.Group("/api/chats")
	chatGroup.Use(middlewares.AuthMiddleware(env, db))
	{
		chatGroup.GET("", ctrl.ListChats)
		chatGroup.GET("/invites", ctrl.ListChatInvites)
		chatGroup.POST("/direct", ctrl.CreateDirectChat)
		chatGroup.POST("/invite", ctrl.CreateChatInvite)
		chatGroup.POST("/invite/respond", ctrl.ResponseChatInvite)

		chatGroup.POST("/media", ctrl.UploadChatMedia)
		chatGroup.GET("/messages/:messageID/download", ctrl.DownloadMessageMedia)
		chatGroup.DELETE("/:chatID", ctrl.DeleteChat)
		chatGroup.POST("/share", ctrl.SharePost)

		chatGroup.GET("/:chatID/background", ctrl.GetChatBackground)
		chatGroup.PUT("/:chatID/background", ctrl.UpdateChatBackground)
		chatGroup.DELETE("/:chatID/background", ctrl.DeleteChatBackground)

		chatGroup.GET("/:chatID/shared-content", ctrl.GetSharedContent)
	}
}
