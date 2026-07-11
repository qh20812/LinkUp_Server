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

func RegisterGroupCallRoutes(router *gin.Engine, hub *groupws.Hub, messageService *services.GroupMessageService, groupService *services.GroupChatService, groupChatHub *groupws.Hub, env config.Env) {
	router.GET("/api/group-calls/ws", groupws.ServeGroupCallWS(hub, messageService, groupService, groupChatHub, env))
}
