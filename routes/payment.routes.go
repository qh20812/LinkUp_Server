package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"
	"linkup/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterPaymentRoutes(router *gin.Engine, ctrl *controllers.PaymentController, env config.Env, db *gorm.DB) {
	// Tạo link thanh toán cho partner đã đăng nhập.
	router.POST("/api/ads-management/payments/vnpay",
		middlewares.AuthMiddleware(env, db),
		middlewares.RequireRoles(db, models.RoleUser, models.RolePartner, models.RoleAdmin, models.RoleSuperAdmin),
		ctrl.CreateVNPayPayment,
	)

	// Return (trình duyệt) + IPN (server-to-server) — công khai, xác thực bằng chữ ký.
	router.GET("/api/payments/vnpay/return", ctrl.VNPayReturn)
	router.GET("/api/payments/vnpay/ipn", ctrl.VNPayIPN)
}
