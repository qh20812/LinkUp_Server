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
		postGroup.DELETE("/:id", middlewares.AuthMiddleware(env), ctrl.DeletePost)
		postGroup.POST("/:id/react", middlewares.AuthMiddleware(env), ctrl.ReactPost)

		postGroup.POST("/:id/comments", middlewares.AuthMiddleware(env), ctrl.CreateComment)
		postGroup.GET("/:id/comments", ctrl.GetComments)

		postGroup.POST("/:id/share", middlewares.AuthMiddleware(env), ctrl.SharePost)
		postGroup.POST("/:id/save", middlewares.AuthMiddleware(env), ctrl.SavePost)

		postGroup.GET("/hashtag/:name", ctrl.GetPostsByHashtag)
	}
}
