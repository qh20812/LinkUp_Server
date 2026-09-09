package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterFriendRoutes(router *gin.Engine, friendController *controllers.FriendController, env config.Env, db *gorm.DB) {
	friend := router.Group("/api/friend-requests")
	friend.Use(middlewares.AuthMiddleware(env, db))
	{
		friend.GET("", friendController.GetFriendRequests)
		friend.POST("/:userID", friendController.ToggleFriendRequest)
		friend.PUT("/:id/accept", friendController.AcceptFriendRequest)
		friend.DELETE("/:id", friendController.RejectFriendRequest)
	}

	friends := router.Group("/api/friends")
	friends.Use(middlewares.AuthMiddleware(env, db))
	{
		friends.GET("", friendController.GetFriends)
		friends.GET("/search", friendController.SearchFriends)
		friends.GET("/suggestions", friendController.GetFriendSuggestions)
		friends.DELETE("/:userID", friendController.Unfriend)
	}

	router.GET("/api/friend-status/:userID", middlewares.AuthMiddleware(env, db), friendController.GetFriendStatus)
}
