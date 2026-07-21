package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"
	"linkup/models"
	"linkup/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterAdRoutes(
	router *gin.Engine,
	ctrl *controllers.AdController,
	adService services.AdService,
	env config.Env,
	db *gorm.DB,
) {
	adsManagement := router.Group("/ads-management")
	{
		adsManagement.POST("",
			middlewares.AuthMiddleware(env),
			middlewares.RequireRoles(db, models.RoleSuperAdmin, models.RoleAdmin, models.RolePartner),
			ctrl.CreateAd,
		)

		adsManagement.GET("",
			middlewares.AuthMiddleware(env),
			middlewares.RequireRoles(db, models.RoleSuperAdmin, models.RoleAdmin, models.RolePartner),
			ctrl.GetAdminList,
		)

		adsManagement.GET("/:id/analytics",
			middlewares.AuthMiddleware(env),
			middlewares.RequireRoles(db, models.RoleSuperAdmin, models.RoleAdmin, models.RolePartner),
			middlewares.CheckAdOwnership(db, adService),
			ctrl.GetAnalytics,
		)

		adsManagement.PATCH("/:id/status",
			middlewares.AuthMiddleware(env),
			middlewares.RequireRoles(db, models.RoleSuperAdmin, models.RoleAdmin, models.RolePartner),
			middlewares.CheckAdOwnership(db, adService),
			ctrl.UpdateStatus,
		)
	}

	customerGroup := router.Group("/customer")
	{
		customerGroup.GET("/feed",
			middlewares.AuthMiddleware(env),
			ctrl.GetUserFeed,
		)

		customerGroup.POST("/ads/:id/track",
			middlewares.AuthMiddleware(env),
			ctrl.TrackAction,
		)
	}
}
