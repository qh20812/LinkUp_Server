package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterCallRoutes(router *gin.Engine, ctrl *controllers.VoiceCallController, env config.Env, db *gorm.DB) {
	// WS endpoint — auth handled via ?token= (not AuthMiddleware)
	router.GET("/api/calls/ws", ctrl.HandleWebsocket)

	callGroup := router.Group("/api/calls")
	callGroup.Use(middlewares.AuthMiddleware(env, db))
	{
		callGroup.GET("/ice-servers", ctrl.GetIceServers)
		callGroup.GET("/history", ctrl.GetCallHistory)

		// Missed-call badge — registered before /:callID to avoid route shadowing.
		callGroup.GET("/missed/count", ctrl.GetMissedCallCount)
		callGroup.POST("/missed/read", ctrl.MarkMissedRead)

		// Hide a call from history — before /:callID so "hide" doesn't match as a callID.
		callGroup.DELETE("/:callID/hide", ctrl.HideCall)

		callGroup.GET("/:callID", ctrl.GetCallDetail)
		callGroup.POST("/initiate", ctrl.InitiateCall)
		callGroup.POST("/:callID/accept", ctrl.AcceptCall)
		callGroup.POST("/:callID/reject", ctrl.RejectCall)
		callGroup.POST("/:callID/mute", ctrl.ToggleMute)
		callGroup.POST("/:callID/video", ctrl.ToggleVideo)
	}
}
