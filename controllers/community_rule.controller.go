package controllers

import (
	"linkup/dto"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommunityRuleController struct {
	ruleService *services.CommunityRuleService
}

func NewCommunityRuleController(ruleService *services.CommunityRuleService) *CommunityRuleController {
	return &CommunityRuleController{ruleService: ruleService}
}

func (ctrl *CommunityRuleController) CreateRule(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	uid := userID.(string)

	communityID := c.Param("communityID")

	var input dto.CreateCommunityRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ hoặc thiếu trường bắt buộc"})
		return
	}

	rule, err := ctrl.ruleService.CreateRule(c.Request.Context(), uid, communityID, input.Category, input.Title, input.Content, &input.Position)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thêm nội quy thành công!",
		"rule_id": rule.ID,
	})
}

func (ctrl *CommunityRuleController) UpdateRule(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	uid := userID.(string)

	ruleID := c.Param("ruleID")

	var input dto.UpdateCommunityRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ"})
		return
	}

	rule, err := ctrl.ruleService.UpdateRule(c.Request.Context(), uid, ruleID, input.Title, input.Content, &input.Category, input.Position)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật nội quy thành công!",
		"rule_id": rule.ID,
	})
}

func (ctrl *CommunityRuleController) DeleteRule(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	uid := userID.(string)

	ruleID := c.Param("ruleID")

	if err := ctrl.ruleService.DeleteRule(c.Request.Context(), uid, ruleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xoá nội quy thành công!"})
}

func (ctrl *CommunityRuleController) GetRules(c *gin.Context) {
	communityID := c.Param("communityID")

	rules, err := ctrl.ruleService.GetRulesByCommunity(c.Request.Context(), communityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rules": rules})
}