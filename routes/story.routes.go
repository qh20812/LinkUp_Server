package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterStoryRoutes(router *gin.Engine, ctrl *controllers.StoryController, env config.Env, db *gorm.DB) {
	// Cấu hình router static giúp client tải và xem được file ảnh/video
	router.Static("/static/stories", "./uploads/stories")

	storyGroup := router.Group("/api/stories")
	{
		// 1. Đăng tải Story mới (Cần đăng nhập)
		storyGroup.POST("", middlewares.AuthMiddleware(env, db), ctrl.CreateStory)

		// 2. Lấy danh sách các story đầu trang chủ (Công khai)
		storyGroup.GET("/feed", ctrl.GetHomeFeed)

		// 3. Kiểm tra user có story active không (Công khai)
		storyGroup.GET("/user/:userID/active", ctrl.CheckUserStory)

		// 4. Lấy danh sách story của 1 user (Optional auth để xác định đã xem chưa)
		storyGroup.GET("/user/:userID", middlewares.OptionalAuth(env, db), ctrl.GetUserStories)

		// 5. Mở xem chi tiết một Story cụ thể (Cần đăng nhập để ghi nhận view)
		storyGroup.GET("/:id", middlewares.AuthMiddleware(env, db), ctrl.ViewStory)

		// 6. Tương tác với Story (React, reply, share - Cần đăng nhập)
		storyGroup.POST("/:id/interact", middlewares.AuthMiddleware(env, db), ctrl.Interact)

		// 7. Thống kê chi tiết Analytics dành cho chủ story (Cần đăng nhập)
		storyGroup.GET("/:id/analytics", middlewares.AuthMiddleware(env, db), ctrl.GetAnalytics)
	}
}
