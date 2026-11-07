package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"
	"linkup/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterPackageRoutes(router *gin.Engine, ctrl *controllers.PackageController, env config.Env, db *gorm.DB) {
	// API Công khai danh sách gói
	router.GET("/api/ads/packages", ctrl.GetPackages)

	partnerGroup := router.Group("/api/ads-management")
	{
		// Kích hoạt gói trực tiếp (không qua thanh toán) chỉ dành cho admin —
		// trước đây route này mở cho mọi user nên có thể bỏ qua VNPay. Luồng
		// mua gói của người dùng đi qua POST /api/ads-management/payments/vnpay.
		partnerGroup.POST("/subscribe",
			middlewares.AuthMiddleware(env, db),
			middlewares.RequireRoles(db, models.RoleAdmin, models.RoleSuperAdmin),
			ctrl.Subscribe,
		)

		partnerGroup.GET("/subscription",
			middlewares.AuthMiddleware(env, db),
			middlewares.RequireRoles(db, models.RoleUser, models.RolePartner, models.RoleAdmin, models.RoleSuperAdmin),
			ctrl.GetMySubscription,
		)
	}

	// ── Admin quản trị gói quảng cáo ──
	adminPackages := router.Group("/api/admin/packages")
	{
		adminPackages.GET("",
			middlewares.AuthMiddleware(env, db),
			middlewares.RequireRoles(db, models.RoleAdmin, models.RoleSuperAdmin),
			ctrl.AdminListPackages,
		)
		adminPackages.POST("",
			middlewares.AuthMiddleware(env, db),
			middlewares.RequireRoles(db, models.RoleAdmin, models.RoleSuperAdmin),
			ctrl.AdminCreatePackage,
		)
		adminPackages.PATCH("/:id",
			middlewares.AuthMiddleware(env, db),
			middlewares.RequireRoles(db, models.RoleAdmin, models.RoleSuperAdmin),
			ctrl.AdminUpdatePackage,
		)
	}
}
