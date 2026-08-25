package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterCommunityRoutes(router *gin.Engine, ctrl *controllers.CommunityController, env config.Env, db *gorm.DB) {
	// Public routes — OptionalAuth so membership_status and is_creator are returned when token present
	router.GET("/api/communities", middlewares.OptionalAuth(env, db), ctrl.ListCommunities)
	router.GET("/api/communities/:communityID", middlewares.OptionalAuth(env, db), ctrl.GetCommunityDetail)
	router.GET("/api/communities/:communityID/posts", middlewares.OptionalAuth(env, db), ctrl.GetCommunityPosts)

	communityGroup := router.Group("/api/communities")
	communityGroup.Use(middlewares.AuthMiddleware(env, db))
	{
		communityGroup.GET("/joined", ctrl.ListJoinedCommunities)
		communityGroup.GET("/created", ctrl.ListCreatedCommunities)
		communityGroup.POST("", ctrl.CreateCommunity)
		communityGroup.PUT("/:communityID", ctrl.UpdateCommunity)
		communityGroup.PUT("/:communityID/background", ctrl.SetCommunityBackground)
		communityGroup.POST("/:communityID/join", ctrl.RequestJoin)
		communityGroup.GET("/:communityID/join-requests", ctrl.ListPendingJoinRequests)
		communityGroup.PUT("/:communityID/join-requests/:requestID/approve", ctrl.ApproveJoinRequest)
		communityGroup.PUT("/:communityID/join-requests/:requestID/reject", ctrl.RejectJoinRequest)
		communityGroup.GET("/:communityID/members", ctrl.GetCommunityMembers)
		communityGroup.PUT("/:communityID/members/:memberID/role", ctrl.UpdateMemberRole)
		communityGroup.DELETE("/:communityID/members/:memberID", ctrl.KickMember)
		communityGroup.DELETE("/:communityID/leave", ctrl.LeaveCommunity)
		communityGroup.POST("/:communityID/transfer-ownership", ctrl.TransferOwnership)

		// Invite codes management
		communityGroup.POST("/:communityID/invite-codes", ctrl.CreateInviteCode)
		communityGroup.GET("/:communityID/invite-codes", ctrl.ListInviteCodes)
		communityGroup.DELETE("/:communityID/invite-codes/:codeID", ctrl.DeactivateInviteCode)

		// Direct invitations
		communityGroup.POST("/:communityID/invitations", ctrl.SendInvitation)
		communityGroup.GET("/invitations", ctrl.ListMyInvitations)
		communityGroup.PUT("/invitations/:invitationID/respond", ctrl.RespondInvitation)
	}
}
