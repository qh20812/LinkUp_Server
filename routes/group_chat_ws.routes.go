package routes

import (
	"linkup/config"
	"linkup/groupws"
	"linkup/services"

	"github.com/gin-gonic/gin"
)

func RegisterGroupChatWebSocketRoute(router *gin.Engine, hub *groupws.Hub, messageService *services.GroupMessageService, groupService *services.GroupChatService, env config.Env) {
	router.GET("/api/group-chats/ws", groupws.ServeGroupWS(hub, messageService, groupService, env))
}
