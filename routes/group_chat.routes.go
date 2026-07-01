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
	}
}
