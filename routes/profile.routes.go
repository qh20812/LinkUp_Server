package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterProfileRoutes(router *gin.Engine, profileController *controllers.ProfileController, env config.Env, db *gorm.DB) {
	profile := router.Group("/api/profile")
	{
		profile.GET("", middlewares.AuthMiddleware(env, db), profileController.ViewProfile)
		profile.PATCH("", middlewares.AuthMiddleware(env, db), profileController.EditProfile)
		profile.GET("/:userID", profileController.ViewProfileByID)
	}

	
}
