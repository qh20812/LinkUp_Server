package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterFriendRoutes(router *gin.Engine, friendController *controllers.FriendController, env config.Env) {
	friend := router.Group("/api/friend-requests")
	friend.Use(middlewares.AuthMiddleware(env))
	{
		friend.GET("", friendController.GetFriendRequests)
		friend.POST("/:userID", friendController.ToggleFriendRequest)
		friend.PUT("/:id/accept", friendController.AcceptFriendRequest)
		friend.DELETE("/:id", friendController.RejectFriendRequest)
	}
}
