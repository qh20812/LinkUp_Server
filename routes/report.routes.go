package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterReportRoutes(router *gin.Engine, reportController *controllers.ReportController, env config.Env) {
	report := router.Group("/api/reports")
	report.Use(middlewares.AuthMiddleware(env))
	{
		report.POST("", reportController.CreateReport)
		report.PUT("/:id", reportController.UpdateReport)
	}
}
