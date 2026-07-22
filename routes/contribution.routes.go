package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterContributionRoutes(router *gin.Engine, ctrl *controllers.ContributionController, env config.Env, db *gorm.DB) {
	policyGroup := router.Group("/api/communities/:communityID/policy")
	policyGroup.Use(middlewares.AuthMiddleware(env, db))
	{
		policyGroup.GET("", ctrl.GetPolicy)
		policyGroup.PUT("", ctrl.UpdatePolicy)
	}

	contributionGroup := router.Group("/api/communities/:communityID/contributions")
	{
		contributionGroup.GET("/leaderboard", ctrl.GetLeaderboard)
		contributionGroup.GET("/me", middlewares.AuthMiddleware(env, db), ctrl.GetMyContribution)
		contributionGroup.GET("/:userID", ctrl.GetUserContribution)
	}

	membersGroup := router.Group("/api/communities/:communityID/contributions/members")
	{
		membersGroup.GET("", ctrl.GetCommunityMembers)
	}

	challengeGroup := router.Group("/api/communities/:communityID/challenges")
	{
		challengeGroup.GET("", ctrl.GetActiveChallenges)
		challengeGroup.POST("", middlewares.AuthMiddleware(env, db), ctrl.CreateChallenge)
		challengeGroup.POST("/:challengeID/join", middlewares.AuthMiddleware(env, db), ctrl.JoinChallenge)
		challengeGroup.GET("/:challengeID/participants", ctrl.GetChallengeParticipants)
	}
}
