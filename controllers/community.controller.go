package controllers

import (
	"linkup/dto"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommunityController struct {
	communityService *services.CommunityService
}

func NewCommunityController(communityService *services.CommunityService) *CommunityController {
	return &CommunityController{communityService: communityService}
}

func (ctrl *CommunityController) CreateCommunity(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}

	var input dto.CreateCommunityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ hoặc thiếu trường bắt buộc"})
		return
	}

	community, err := ctrl.communityService.CreateCommunity(c.Request.Context(), userID.(string), input.Name, input.Description, input.AvatarURI)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Tạo cộng đồng thành công!",
		"community_id": community.ID,
	})
}
