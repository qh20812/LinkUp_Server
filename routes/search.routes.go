package routes

import (
	"linkup/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterSearchRoutes(router *gin.Engine, searchController *controllers.SearchController) {
	router.GET("/api/search", searchController.Search)
	router.GET("/api/trending", searchController.GetTrending)
}
