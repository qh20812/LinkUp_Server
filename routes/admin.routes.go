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

	adminGroup.GET("/analytics", adminController.GetDashboardAnalytics)

	adminGroup.GET("/users", adminController.ListUsers)
	adminGroup.PUT("/users/:userID/status", adminController.UpdateUserStatus)
	adminGroup.POST("/users/:userID/ban", adminController.BanUser)

	adminGroup.GET("/posts", adminController.ListPosts)
	adminGroup.PUT("/posts/:postID/hide", adminController.HidePost)
	adminGroup.PUT("/posts/:postID/status", adminController.ChangePostStatus)

	adminGroup.GET("/reports", adminController.ListReports)
	adminGroup.GET("/reports/:reportID", adminController.GetReportDetail)
	adminGroup.PUT("/reports/:reportID/decision", adminController.ReviewReport)
}
