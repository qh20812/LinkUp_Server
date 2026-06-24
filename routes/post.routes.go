package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterPostRoutes(router *gin.Engine, ctrl *controllers.PostController, env config.Env) {
	postGroup := router.Group("/posts")
	{
		postGroup.POST("", middlewares.AuthMiddleware(env), ctrl.CreatePost)
		postGroup.GET("", ctrl.GetPosts)
		postGroup.GET("/:id", ctrl.ViewPostDetail)
		postGroup.POST("/:id/react", middlewares.AuthMiddleware(env), ctrl.ReactPost)
		postGroup.POST("/:id/comments", middlewares.AuthMiddleware(env), ctrl.CreateComment)
	}
}
