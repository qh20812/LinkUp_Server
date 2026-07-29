package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"
	"linkup/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterAdminRoutes(router *gin.Engine, adminController *controllers.AdminController, env config.Env, db *gorm.DB) {
	adminGroup := router.Group("/api/admin")
	adminGroup.Use(middlewares.AuthMiddleware(env, db))
	adminGroup.Use(middlewares.RequireRoles(db, models.RoleSuperAdmin, models.RoleAdmin))

	adminGroup.GET("/analytics", adminController.GetDashboardAnalytics)

	adminGroup.GET("/users", adminController.ListUsers)
	adminGroup.PUT("/users/:userID/status", adminController.UpdateUserStatus)
	adminGroup.POST("/users/:userID/ban", adminController.BanUser)

	adminGroup.GET("/posts", adminController.ListPosts)
	adminGroup.PUT("/posts/:postID/hide", adminController.HidePost)
	adminGroup.PUT("/posts/:postID/status", adminController.ChangePostStatus)

	adminGroup.GET("/comments", adminController.ListComments)
	adminGroup.PUT("/comments/:commentID/hide", adminController.HideComment)
	adminGroup.PUT("/comments/:commentID/reveal", adminController.RevealComment)

	adminGroup.GET("/reports", adminController.ListReports)
	adminGroup.GET("/reports/:reportID", adminController.GetReportDetail)
	adminGroup.POST("/reports/:reportID/decision", adminController.ReviewReport)

	// ── Admin Media ──
	adminGroup.GET("/media/grouped", adminController.ListMediaGroupedByUser)
	adminGroup.GET("/media/flagged", adminController.ListFlaggedMedia)
	adminGroup.POST("/media/:id/review", adminController.ReviewMedia)
	adminGroup.POST("/media/cleanup-rejected", adminController.CleanupRejectedMedia)

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
	adminGroup.POST("/communities/:id/unarchive", adminController.UnarchiveCommunity)
	adminGroup.POST("/communities/:id/archive", adminController.ArchiveCommunity)
	adminGroup.POST("/communities/:id/warn", adminController.WarnCommunity)
	adminGroup.DELETE("/groups/:chatID", adminController.DeleteGroup)
	adminGroup.DELETE("/communities/:id", adminController.DeleteCommunity)

	// ── Admin Ads ──
	adminGroup.GET("/ads", adminController.ListAds)
	adminGroup.PATCH("/ads/:id/status", adminController.UpdateAdStatus)
	adminGroup.DELETE("/ads/:id", adminController.DeleteAd)
}
