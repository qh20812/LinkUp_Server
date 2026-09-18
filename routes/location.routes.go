package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterLocationRoutes(router *gin.Engine, locationController *controllers.LocationController, env config.Env, db *gorm.DB) {
	locations := router.Group("/api/locations")
	{
		locations.GET("/provinces", locationController.ListProvinces)
		locations.GET("/wards", locationController.ListWards)
		locations.POST("/reverse-geocode", middlewares.AuthMiddleware(env, db), locationController.ReverseGeocode)
	}
}