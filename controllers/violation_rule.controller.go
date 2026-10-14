package controllers

import (
	"net/http"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/services"

	"github.com/gin-gonic/gin"
)

type ViolationRuleController struct {
	ruleService *services.ViolationRuleService
}

func NewViolationRuleController(ruleService *services.ViolationRuleService) *ViolationRuleController {
	return &ViolationRuleController{ruleService: ruleService}
}

// ListRules: user đã đăng nhập xem rule đang bật để chọn khi report.
// Query: ?target_type=post|comment|user (lọc phạm vi) &keyword=...
func (h *ViolationRuleController) ListRules(c *gin.Context) {
	targetType := c.Query("target_type")
	keyword := c.Query("keyword")

	rules, err := h.ruleService.ListRules(c.Request.Context(), targetType, keyword, false)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"rules": rules, "total": len(rules)})
}

func (h *ViolationRuleController) GetRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	rule, err := h.ruleService.GetRule(c.Request.Context(), ruleID)
	if err != nil {
		errorsapp.Respond(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"rule": rule})
}

// ── Admin (super-admin, check trong service) ──

// ListAllRules: admin xem tất cả kể cả rule đã tắt. Query: ?target_type=&keyword=&include_inactive=true
func (h *ViolationRuleController) ListAllRules(c *gin.Context) {
	adminID := c.GetString("userID")
	_ = adminID
	targetType := c.Query("target_type")
	keyword := c.Query("keyword")
	includeInactive := c.Query("include_inactive") != "false"
	if c.Query("include_inactive") == "" {
		includeInactive = true
	}

	rules, err := h.ruleService.ListRules(c.Request.Context(), targetType, keyword, includeInactive)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"rules": rules, "total": len(rules)})
}

func (h *ViolationRuleController) CreateRule(c *gin.Context) {
	adminID, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	var input dto.CreateViolationRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	rule, err := h.ruleService.CreateRule(c.Request.Context(), adminID.(string), input.Title, input.Description, input.ApplicableTo, input.Severity, input.SortOrder)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Tạo quy tắc vi phạm thành công", "rule": rule})
}

func (h *ViolationRuleController) UpdateRule(c *gin.Context) {
	adminID, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	ruleID := c.Param("id")
	if ruleID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	var input dto.UpdateViolationRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	rule, err := h.ruleService.UpdateRule(c.Request.Context(), adminID.(string), ruleID, input.Title, input.Description, input.ApplicableTo, input.Severity, input.SortOrder)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật quy tắc vi phạm thành công", "rule": rule})
}

func (h *ViolationRuleController) SetRuleActive(c *gin.Context) {
	adminID, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	ruleID := c.Param("id")
	if ruleID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	var input struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := h.ruleService.SetActive(c.Request.Context(), adminID.(string), ruleID, input.IsActive); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật trạng thái quy tắc thành công"})
}
