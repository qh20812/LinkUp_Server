package controllers

import (
	"fmt"
	"linkup/dto"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GroupChatController struct {
	groupService *services.GroupChatService
	chatService  *services.ChatService
}

func NewGroupChatController(groupService *services.GroupChatService, chatService *services.ChatService) *GroupChatController {
	return &GroupChatController{
		groupService: groupService,
		chatService:  chatService,
	}
}

func (ctrl *GroupChatController) CreateGroup(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)

	var input dto.CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ hoặc thiếu trường bắt buộc"})
		return
	}

	group, err := ctrl.groupService.CreateGroup(c.Request.Context(), userID, input.Name, input.AvatarURI, input.MemberIDs)
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
	requesterIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực"})
		return
	}
	requesterID := fmt.Sprintf("%v", requesterIDVal)

	chatID := c.Param("chatID")

	var input dto.AddMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng chọn thành viên hợp lệ cần thêm"})
		return
	}

	err := ctrl.groupService.AddMember(c.Request.Context(), chatID, requesterID, input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thêm thành viên vào nhóm thành công!"})
}

func (ctrl *GroupChatController) LeaveGroup(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực"})
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)

	chatID := c.Param("chatID")

	err := ctrl.groupService.LeaveGroup(c.Request.Context(), chatID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bạn đã rời khỏi nhóm chat thành công"})
}

func (ctrl *GroupChatController) BanMember(c *gin.Context) {
	adminIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực"})
		return
	}
	adminID := fmt.Sprintf("%v", adminIDVal)

	chatID := c.Param("chatID")

	var input dto.BanMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng cung cấp user_id cần chặn"})
		return
	}

	err := ctrl.groupService.BanMember(c.Request.Context(), chatID, adminID, input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã trục xuất và chặn người dùng này tham gia lại nhóm thành công!"})
}

func (ctrl *GroupChatController) SendGroupMessage(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực"})
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)
	chatID := c.Param("chatID")

	var input dto.SendGroupMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		// Chỉ báo lỗi cú pháp JSON đầu vào
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu yêu cầu không hợp lệ"})
		return
	}

	msg, err := ctrl.chatService.SendMessage(
		c.Request.Context(),
		userID,
		chatID,
		input.Content,
		input.EmojiID,
		input.MediaID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Gửi tin nhắn vào nhóm thành công",
		"data":    msg,
	})
}

func (ctrl *GroupChatController) GetSettings(c *gin.Context) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin đăng nhập"})
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)
	chatID := c.Param("chatID")

	settings, err := ctrl.groupService.GetSettings(c.Request.Context(), chatID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": settings})
}

func (ctrl *GroupChatController) UpdateSettings(c *gin.Context) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin đăng nhập"})
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)
	chatID := c.Param("chatID")

	var input dto.GroupChatSettingsDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSONP(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	settings, err := ctrl.groupService.UpdateSettings(c.Request.Context(), chatID, userID, &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"error": "Cập nhật cấu hình nhóm thành công", "data": settings})
}

func (ctrl *GroupChatController) TransferAdmin(c *gin.Context) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực"})
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)
	chatID := c.Param("chatID")

	var input dto.TransferAdminInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng cung cấp target_user_id"})
		return
	}

	if err := ctrl.groupService.TransferAdmin(c.Request.Context(), chatID, userID, input.TargetUserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã chuyển quyền admin thành công"})
}

func (ctrl *GroupChatController) MuteMember(c *gin.Context) {
	adminIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	adminID := fmt.Sprintf("%v", adminIDVal)
	chatID := c.Param("chatID")

	var input dto.MuteMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mute, err := ctrl.groupService.MuteMember(c.Request.Context(), chatID, adminID, input.UserID, input.Reason, input.DurationMins)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã tắt tiếng thành viên", "data": mute})
}

func (ctrl *GroupChatController) UnmuteMember(c *gin.Context) {
	adminIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	adminID := fmt.Sprintf("%v", adminIDVal)
	chatID := c.Param("chatID")

	var input dto.UnmuteMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.groupService.UnmuteMember(c.Request.Context(), chatID, adminID, input.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã mở lại quyền gửi tin nhắn cho thành viên"})
}
