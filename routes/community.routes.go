package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterCommunityRoutes(router *gin.Engine, ctrl *controllers.CommunityController, env config.Env) {
	communityGroup := router.Group("/api/communities")
	communityGroup.Use(middlewares.AuthMiddleware(env))
	{
		communityGroup.POST("", ctrl.CreateCommunity)
		communityGroup.PUT("/:communityID/background", ctrl.SetCommunityBackground)
		communityGroup.POST("/:communityID/join", ctrl.RequestJoin)
		communityGroup.GET("/:communityID/join-requests", ctrl.ListPendingJoinRequests)
		communityGroup.PUT("/:communityID/join-requests/:requestID/approve", ctrl.ApproveJoinRequest)
		communityGroup.PUT("/:communityID/join-requests/:requestID/reject", ctrl.RejectJoinRequest)
	}
}
