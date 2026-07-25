package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterPostRoutes(router *gin.Engine, ctrl *controllers.PostController, env config.Env, db *gorm.DB) {
	postGroup := router.Group("/posts")
	{
		postGroup.POST("", middlewares.AuthMiddleware(env, db), ctrl.CreatePost)
		postGroup.GET("", ctrl.GetPosts)
		postGroup.GET("/:id", ctrl.ViewPostDetail)
		postGroup.DELETE("/:id", middlewares.AuthMiddleware(env, db), ctrl.DeletePost)
		postGroup.POST("/:id/react", middlewares.AuthMiddleware(env, db), ctrl.ReactPost)

		postGroup.POST("/:id/comments", middlewares.AuthMiddleware(env, db), ctrl.CreateComment)
		postGroup.GET("/:id/comments", ctrl.GetComments)

		postGroup.POST("/:id/share", middlewares.AuthMiddleware(env, db), ctrl.SharePost)
		postGroup.POST("/:id/save", middlewares.AuthMiddleware(env, db), ctrl.SavePost)

		postGroup.GET("/hashtag/:name", ctrl.GetPostsByHashtag)
	}
}
