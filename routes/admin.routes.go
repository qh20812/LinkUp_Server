package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(router *gin.Engine, adminController *controllers.AdminController, env config.Env) {
	adminGroup := router.Group("/api/admin")
	adminGroup.Use(middlewares.AuthMiddleware(env))

	adminGroup.GET("/users", adminController.ListUsers)
	adminGroup.PUT("/users/:userID/status", adminController.UpdateUserStatus)
	adminGroup.POST("/users/:userID/ban", adminController.BanUser)

	adminGroup.GET("/posts", adminController.ListPosts)
	adminGroup.PUT("/posts/:postID/hide", adminController.HidePost)
	adminGroup.PUT("/posts/:postID/status", adminController.ChangePostStatus)

	adminGroup.GET("/reports", adminController.ListReports)
	adminGroup.GET("/reports/:reportID", adminController.GetReportDetail)
	adminGroup.PUT("/reports/:reportID/decision", adminController.ReviewReport)

	// ── Admin Groups ──
	adminGroup.GET("/groups", adminController.ListGroups)
	adminGroup.GET("/groups/:chatID", adminController.GetGroupDetail)
	adminGroup.GET("/groups/:chatID/members", adminController.ListGroupMembers)
	adminGroup.GET("/groups/:chatID/logs", adminController.GetGroupModerationLogs)
	adminGroup.POST("/groups/:chatID/hide", adminController.HideGroup)
	adminGroup.POST("/groups/:chatID/unhide", adminController.UnhideGroup)
	adminGroup.POST("/groups/:chatID/archive", adminController.ArchiveGroup)
	adminGroup.POST("/groups/:chatID/warn", adminController.WarnGroup)

	// ── Admin Communities ──
	adminGroup.GET("/communities", adminController.ListCommunities)
	adminGroup.GET("/communities/:id", adminController.GetCommunityDetail)
	adminGroup.GET("/communities/:id/members", adminController.ListCommunityMembers)
	adminGroup.GET("/communities/:id/logs", adminController.GetCommunityModerationLogs)
	adminGroup.POST("/communities/:id/hide", adminController.HideCommunity)
	adminGroup.POST("/communities/:id/unhide", adminController.UnhideCommunity)
	adminGroup.POST("/communities/:id/archive", adminController.ArchiveCommunity)
	adminGroup.POST("/communities/:id/warn", adminController.WarnCommunity)
	adminGroup.DELETE("/groups/:chatID", adminController.DeleteGroup)
	adminGroup.DELETE("/communities/:id", adminController.DeleteCommunity)
}
