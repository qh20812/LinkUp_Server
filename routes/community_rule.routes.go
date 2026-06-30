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
		// xem quy tắc cộng đồng (không cần xác thực)
		rules.GET("", ctrl.GetRules)
	}

	rulesAuth := router.Group("/api/communities/:communityID/rules")
	rulesAuth.Use(middlewares.AuthMiddleware(env))
	{
		// tạo quy tắc cộng đồng
		rulesAuth.POST("", ctrl.CreateRule)
		// cập nhật quy tắc cộng đồng
		rulesAuth.PUT("/:ruleID", ctrl.UpdateRule)
		// xóa quy tắc cộng đồng
		rulesAuth.DELETE("/:ruleID", ctrl.DeleteRule)
	}
}
