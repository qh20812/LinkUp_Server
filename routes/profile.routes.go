package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterProfileRoutes(router *gin.Engine, profileController *controllers.ProfileController, env config.Env) {
	profile := router.Group("/api/profile")
	{
		profile.GET("", middlewares.AuthMiddleware(env), profileController.ViewProfile)
	}
}
