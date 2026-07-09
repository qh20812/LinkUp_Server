package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterStoryRoutes(router *gin.Engine, ctrl *controllers.StoryController, env config.Env) {
	// Cấu hình router static giúp client tải và xem được file ảnh/video
	router.Static("/static/stories", "./uploads/stories")

	storyGroup := router.Group("/stories")
	{
		// 1. Đăng tải Story mới (Cần đăng nhập)
		storyGroup.POST("", middlewares.AuthMiddleware(env), ctrl.CreateStory)

		// 2. Lấy danh sách các story đầu trang chủ (Công khai)
		storyGroup.GET("/feed", ctrl.GetHomeFeed)

		// 3. Mở xem chi tiết một Story cụ thể (Cần đăng nhập để ghi nhận view)
		storyGroup.GET("/:id", middlewares.AuthMiddleware(env), ctrl.ViewStory)

		// 4. Tương tác với Story (React, reply, share - Cần đăng nhập)
		storyGroup.POST("/:id/interact", middlewares.AuthMiddleware(env), ctrl.Interact)

		// 5. Thống kê chi tiết Analytics dành cho chủ story (Cần đăng nhập)
		storyGroup.GET("/:id/analytics", middlewares.AuthMiddleware(env), ctrl.GetAnalytics)
	}
}
