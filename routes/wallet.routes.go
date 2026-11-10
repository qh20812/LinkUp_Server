package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"
	"linkup/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterWalletRoutes đăng ký API ví trả trước + nạp tiền bằng chuyển khoản.
//
// Người dùng đã đăng nhập mới xem được thông tin tài khoản nhận tiền (QR) nên
// toàn bộ nhóm /api/wallet đều có AuthMiddleware. Nhóm admin dùng thêm RBAC.
func RegisterWalletRoutes(
	router *gin.Engine,
	ctrl *controllers.WalletController,
	env config.Env,
	db *gorm.DB,
) {
	wallet := router.Group("/api/wallet")
	wallet.Use(middlewares.AuthMiddleware(env, db))
	{
		wallet.GET("", ctrl.GetWallet)
		wallet.GET("/transactions", ctrl.ListTransactions)
		wallet.GET("/topups", ctrl.ListTopups)
		wallet.POST("/topups", ctrl.CreateTopup)
		wallet.POST("/topups/:id/report", ctrl.ReportTopup)
		wallet.POST("/topups/:id/cancel", ctrl.CancelTopup)
	}

	adminTopups := router.Group("/api/admin/wallet/topups")
	adminTopups.Use(
		middlewares.AuthMiddleware(env, db),
		middlewares.RequireRoles(db, models.RoleSuperAdmin, models.RoleAdmin),
	)
	{
		adminTopups.GET("", ctrl.AdminListTopups)
		adminTopups.POST("/:id/approve", ctrl.AdminApproveTopup)
		adminTopups.POST("/:id/reject", ctrl.AdminRejectTopup)
	}

	adminWallets := router.Group("/api/admin/wallets")
	adminWallets.Use(
		middlewares.AuthMiddleware(env, db),
		middlewares.RequireRoles(db, models.RoleSuperAdmin, models.RoleAdmin),
	)
	{
		adminWallets.GET("", ctrl.AdminListWallets)
	}
}
