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
		partnerGroup.POST("/subscribe",
			middlewares.AuthMiddleware(env, db),
			middlewares.RequireRoles(db, models.RoleUser, models.RolePartner, models.RoleAdmin, models.RoleSuperAdmin),
			ctrl.Subscribe,
		)

		partnerGroup.GET("/subscription",
			middlewares.AuthMiddleware(env, db),
			middlewares.RequireRoles(db, models.RolePartner, models.RoleAdmin, models.RoleSuperAdmin),
			ctrl.GetMySubscription,
		)
	}
}
