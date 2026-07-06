package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterCallRoutes(router *gin.Engine, ctrl *controllers.VoiceCallController, env config.Env) {
	callGroup := router.Group("/api/calls")
	callGroup.Use(middlewares.AuthMiddleware(env))
	{
		callGroup.GET("/ws", ctrl.HandleWebsocket)
		callGroup.GET("/history", ctrl.GetCallHistory)
		callGroup.GET("/:callID", ctrl.GetCallDetail)
		callGroup.POST("/initiate", ctrl.InitiateCall)
		callGroup.POST("/:callID/accept", ctrl.AcceptCall)
		callGroup.POST("/:callID/reject", ctrl.RejectCall)
		callGroup.POST("/:callID/mute", ctrl.ToggleMute)
	}
}
