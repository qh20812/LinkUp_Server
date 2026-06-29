package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterCommunityRuleRoutes(router *gin.Engine, ctrl *controllers.CommunityRuleController, env config.Env) {
	rules := router.Group("/api/communities/:communityID/rules")
	{
		rules.GET("", ctrl.GetRules)
	}

	rulesAuth := router.Group("/api/communities/:communityID/rules")
	rulesAuth.Use(middlewares.AuthMiddleware(env))
	{
		rulesAuth.POST("", ctrl.CreateRule)
		rulesAuth.PUT("/:ruleID", ctrl.UpdateRule)
		rulesAuth.DELETE("/:ruleID", ctrl.DeleteRule)
	}
}
