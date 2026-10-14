package routes

import (
	"linkup/config"
	"linkup/controllers"
	"linkup/middlewares"
	"linkup/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterViolationRuleRoutes đăng ký endpoint public cho user thường:
// GET /api/violation-rules?target_type=post&keyword= — đổ dropdown khi report.
func RegisterViolationRuleRoutes(router *gin.Engine, violationRuleController *controllers.ViolationRuleController, env config.Env, db *gorm.DB) {
	rules := router.Group("/api/violation-rules")
	rules.Use(middlewares.AuthMiddleware(env, db))
	{
		rules.GET("", violationRuleController.ListRules)
		rules.GET("/:id", violationRuleController.GetRule)
	}
}

// RegisterAdminViolationRuleRoutes đăng ký CRUD cho super-admin:
// GET/POST /api/admin/violation-rules, PUT/PATCH /api/admin/violation-rules/:id
func RegisterAdminViolationRuleRoutes(router *gin.Engine, violationRuleController *controllers.ViolationRuleController, env config.Env, db *gorm.DB) {
	admin := router.Group("/api/admin/violation-rules")
	admin.Use(middlewares.AuthMiddleware(env, db))
	admin.Use(middlewares.RequireRoles(db, models.RoleSuperAdmin, models.RoleAdmin))
	{
		admin.GET("", violationRuleController.ListAllRules)
		admin.POST("", violationRuleController.CreateRule)
		admin.PUT("/:id", violationRuleController.UpdateRule)
		admin.PATCH("/:id/active", violationRuleController.SetRuleActive)
	}
}
