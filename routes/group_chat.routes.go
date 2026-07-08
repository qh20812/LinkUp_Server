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
		groupChatGroup.POST("/:chatID/member-requests/:requestID/approve", ctrl.ApproveMemberRequest)
		groupChatGroup.POST("/:chatID/member-requests/:requestID/reject", ctrl.RejectMemberRequest)
		groupChatGroup.POST("/:chatID/ban", ctrl.BanMember)
		groupChatGroup.POST("/:chatID/messages", ctrl.SendGroupMessage)
		groupChatGroup.POST("/:chatID/transfer-admin", ctrl.TransferAdmin)
		groupChatGroup.GET("/:chatID/settings", ctrl.GetSettings)
		groupChatGroup.PUT("/:chatID/settings", ctrl.UpdateSettings)
		groupChatGroup.POST("/:chatID/mute", ctrl.MuteMember)
		groupChatGroup.POST("/:chatID/unmute", ctrl.UnmuteMember)
	}
}
