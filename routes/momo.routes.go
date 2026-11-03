package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"
	"linkup/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterMoMoRoutes(router *gin.Engine, ctrl *controllers.MoMoController, env config.Env, db *gorm.DB) {
	router.POST("/api/ads-management/payments/momo",
		middlewares.AuthMiddleware(env, db),
		middlewares.RequireRoles(db, models.RoleUser, models.RolePartner, models.RoleAdmin, models.RoleSuperAdmin),
		ctrl.CreateMoMoPayment,
	)

	// Return (trình duyệt, GET) + IPN (server-to-server, POST JSON).
	router.GET("/api/payments/momo/return", ctrl.MoMoReturn)
	router.POST("/api/payments/momo/ipn", ctrl.MoMoIPN)
}
