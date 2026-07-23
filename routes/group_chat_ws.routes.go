package routes

import (
	"linkup/config"
	"linkup/groupws"
	"linkup/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterGroupChatWebSocketRoute(router *gin.Engine, hub *groupws.Hub, messageService *services.GroupMessageService, groupService *services.GroupChatService, env config.Env, db *gorm.DB) {
	router.GET("/api/group-chats/ws", groupws.ServeGroupWS(hub, messageService, groupService, env, db))
}

func RegisterGroupCallRoutes(router *gin.Engine, hub *groupws.Hub, messageService *services.GroupMessageService, groupService *services.GroupChatService, groupChatHub *groupws.Hub, env config.Env, db *gorm.DB) {
	router.GET("/api/group-calls/ws", groupws.ServeGroupCallWS(hub, messageService, groupService, groupChatHub, env, db))
}
