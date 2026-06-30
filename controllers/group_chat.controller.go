package controllers

import (
	"linkup/dto"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GroupChatController struct {
	groupService *services.GroupChatService
}

func NewGroupChatController(groupService *services.GroupChatService) *GroupChatController {
	return &GroupChatController{groupService: groupService}
}

func (ctrl *GroupChatController) CreateGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}

	// 🌟 Sử dụng Struct từ file DTO của bạn
	var input dto.CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ hoặc thiếu trường bắt buộc"})
		return
	}

	group, err := ctrl.groupService.CreateGroup(c.Request.Context(), userID.(string), input.Name, input.AvatarURI, input.MemberIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Tạo nhóm chat thành công!",
		"group_id": group.ID,
	})
}

func (ctrl *GroupChatController) AddMember(c *gin.Context) {
	requesterID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực"})
		return
	}

	chatID := c.Param("chatID")

	// 🌟 Sử dụng Struct từ file DTO của bạn
	var input dto.AddMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng chọn thành viên hợp lệ cần thêm"})
		return
	}

	err := ctrl.groupService.AddMember(c.Request.Context(), chatID, requesterID.(string), input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thêm thành viên vào nhóm thành công!"})
}
