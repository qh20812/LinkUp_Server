package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterGroupChatRoutes(router *gin.Engine, ctrl *controllers.GroupChatController, env config.Env) {
	groupChatGroup := router.Group("/api/group-chats")
	groupChatGroup.Use(middlewares.AuthMiddleware(env))
	{
		groupChatGroup.POST("", ctrl.CreateGroup)
		groupChatGroup.POST("/:chatID/add-member", ctrl.AddMember)
		groupChatGroup.POST("/:chatID/leave", ctrl.LeaveGroup)
		groupChatGroup.POST("/:chatID/ban", ctrl.BanMember)
		groupChatGroup.POST("/:chatID/messages", ctrl.SendGroupMessage)
		groupChatGroup.POST("/:chatID/transfer-admin", ctrl.TransferAdmin)
		groupChatGroup.POST("/:chatID/transfer-ownership", ctrl.TransferOwnership)
		groupChatGroup.GET("/:chatID/settings", ctrl.GetSettings)
		groupChatGroup.PUT("/:chatID/settings", ctrl.UpdateSettings)
		groupChatGroup.POST("/:chatID/mute", ctrl.MuteMember)
		groupChatGroup.POST("/:chatID/unmute", ctrl.UnmuteMember)
	}
}
