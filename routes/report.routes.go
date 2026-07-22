package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterReportRoutes(router *gin.Engine, reportController *controllers.ReportController, env config.Env, db *gorm.DB) {
	report := router.Group("/api/reports")
	report.Use(middlewares.AuthMiddleware(env, db))
	{
		report.POST("", reportController.CreateReport)
		report.PUT("/:id", reportController.UpdateReport)
	}
}
