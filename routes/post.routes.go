package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterPostRoutes(router *gin.Engine, ctrl *controllers.PostController, env config.Env, db *gorm.DB) {
	apiGroup := router.Group("/api")
	{
		apiGroup.GET("/posts", middlewares.AuthMiddleware(env, db), ctrl.GetPosts)
		apiGroup.GET("/posts/saved", middlewares.AuthMiddleware(env, db), ctrl.GetSavedPosts)
		apiGroup.GET("/posts/:id", ctrl.ViewPostDetail)
		apiGroup.POST("/posts", middlewares.AuthMiddleware(env, db), ctrl.CreatePost)
		apiGroup.DELETE("/posts/:id", middlewares.AuthMiddleware(env, db), ctrl.DeletePost)
		apiGroup.POST("/posts/:id/react", middlewares.AuthMiddleware(env, db), ctrl.ReactPost)
		apiGroup.POST("/posts/:id/comments", middlewares.AuthMiddleware(env, db), ctrl.CreateComment)
		apiGroup.GET("/posts/:id/comments", ctrl.GetComments)
		apiGroup.POST("/posts/:id/share", middlewares.AuthMiddleware(env, db), ctrl.SharePost)
		apiGroup.POST("/posts/:id/save", middlewares.AuthMiddleware(env, db), ctrl.SavePost)
		apiGroup.GET("/posts/hashtag/:name", ctrl.GetPostsByHashtag)
		apiGroup.GET("/emojis", ctrl.GetEmojis)
	}
}
