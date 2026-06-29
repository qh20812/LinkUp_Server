package routes

import (
	"linkup/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterTagRoutes(r *gin.Engine, tagCtrl *controllers.TagController) {
	api := r.Group("/api")
	{
		api.GET("/tags/:name/posts", tagCtrl.GetPostsByHashtag)
	}
}
