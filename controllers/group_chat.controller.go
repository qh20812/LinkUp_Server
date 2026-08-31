package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/groupws"
	"linkup/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type GroupChatController struct {
	groupService *services.GroupChatService
	chatService  *services.ChatService
	groupHub     *groupws.Hub
}

func NewGroupChatController(groupService *services.GroupChatService, chatService *services.ChatService, groupHub *groupws.Hub) *GroupChatController {
	return &GroupChatController{
		groupService: groupService,
		chatService:  chatService,
		groupHub:     groupHub,
	}
}

func (ctrl *GroupChatController) CreateGroup(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)

	var input dto.CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	group, err := ctrl.groupService.CreateGroup(c.Request.Context(), userID, input.Name, input.AvatarURI, input.MemberIDs)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
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
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	requesterID := fmt.Sprintf("%v", requesterIDVal)
	chatID := c.Param("chatID")

	var input dto.AddMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	requestID, err := ctrl.groupService.AddMemberWithRequestID(c.Request.Context(), chatID, requesterID, input.UserID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	ctrl.broadcastToChat(chatID, dto.WsEvent{
		Type: "group:member:added",
		Payload: mustMarshalCtrl(map[string]any{
			"chat_id": chatID,
			"user_id": input.UserID,
			"by":      requesterID,
		}),
	})

	c.JSON(http.StatusOK, gin.H{
		"message":    "Đã gửi lời mời tham gia nhóm. Chờ người dùng xác nhận.",
		"request_id": requestID,
	})
}

func (ctrl *GroupChatController) BanMember(c *gin.Context) {
	adminIDVal, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	adminID := fmt.Sprintf("%v", adminIDVal)

	chatID := c.Param("chatID")

	var input dto.BanMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	err := ctrl.groupService.BanMember(c.Request.Context(), chatID, adminID, input.UserID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã trục xuất và chặn người dùng này tham gia lại nhóm thành công!"})
}

func (ctrl *GroupChatController) SendGroupMessage(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)
	chatID := c.Param("chatID")

	var input dto.SendGroupMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	msg, err := ctrl.chatService.SendMessage(
		c.Request.Context(),
		userID,
		chatID,
		input.Content,
		0,
		input.EmojiID,
		input.MediaID,
		nil,
		input.ReplyToMessageID,
		input.SharedPostID,
		nil,
	)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
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
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)
	chatID := c.Param("chatID")

	settings, err := ctrl.groupService.GetSettings(c.Request.Context(), chatID, userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": settings})
}

func (ctrl *GroupChatController) UpdateSettings(c *gin.Context) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)
	chatID := c.Param("chatID")

	var input dto.GroupChatSettingsDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	settings, err := ctrl.groupService.UpdateSettings(c.Request.Context(), chatID, userID, &input)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	if input.Name != nil || input.AvatarURI != nil {
		actorName := ctrl.getActorName(c.Request.Context(), userID)
		detail := ""
		text := ""

		if input.Name != nil {
			detail = "name_changed"
			text = fmt.Sprintf("%s đã đổi tên nhóm thành \"%s\"", actorName, *input.Name)
		}
		if input.AvatarURI != nil {
			detail = "avatar_changed"
			text = fmt.Sprintf("%s đã đổi ảnh nhóm", actorName)
		}

		ctrl.broadcastToChat(chatID, dto.WsEvent{
			Type: "group:settings:updated",
			Payload: mustMarshalCtrl(map[string]any{
				"chat_id":   chatID,
				"name":      settings.Name,
				"avatar_uri": settings.AvatarURI,
				"by":        userID,
				"actor_name": actorName,
				"detail":    detail,
				"text":      text,
			}),
		})
	}

	c.JSON(http.StatusOK, gin.H{"error": "Cập nhật cấu hình nhóm thành công", "data": settings})
}

func (ctrl *GroupChatController) TransferAdmin(c *gin.Context) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)
	chatID := c.Param("chatID")

	var input dto.TransferAdminInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := ctrl.groupService.TransferAdmin(c.Request.Context(), chatID, userID, input.TargetUserID); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	targetName := ctrl.getActorName(c.Request.Context(), input.TargetUserID)
	ctrl.broadcastToChat(chatID, dto.WsEvent{
		Type: "group:admin:transferred",
		Payload: mustMarshalCtrl(map[string]any{
			"chat_id":        chatID,
			"target_user_id": input.TargetUserID,
			"by":             userID,
			"actor_name":     targetName,
		}),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Đã chuyển quyền admin thành công"})
}

func (ctrl *GroupChatController) TransferOwnership(c *gin.Context) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)
	chatID := c.Param("chatID")

	var input dto.GroupChatTransferOwnershipInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := ctrl.groupService.TransferOwnership(c.Request.Context(), chatID, userID, input.TargetUserID, input.KeepAdmin); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã chuyển quyền sở hữu nhóm thành công"})
}

func (ctrl *GroupChatController) MuteMember(c *gin.Context) {
	adminIDVal, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	adminID := fmt.Sprintf("%v", adminIDVal)
	chatID := c.Param("chatID")

	var input dto.MuteMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	mute, err := ctrl.groupService.MuteMember(c.Request.Context(), chatID, adminID, input.UserID, input.Reason, input.DurationMins)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã tắt tiếng thành viên", "data": mute})
}

func (ctrl *GroupChatController) UnmuteMember(c *gin.Context) {
	adminIDVal, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	adminID := fmt.Sprintf("%v", adminIDVal)
	chatID := c.Param("chatID")

	var input dto.UnmuteMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	if err := ctrl.groupService.UnmuteMember(c.Request.Context(), chatID, adminID, input.UserID); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã mở lại quyền gửi tin nhắn cho thành viên"})
}

func (ctrl *GroupChatController) ApproveMemberRequest(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)
	chatID := c.Param("chatID")
	requestID := c.Param("requestID")

	if err := ctrl.groupService.ApproveMemberRequest(c.Request.Context(), chatID, userID, requestID); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	memberName := ctrl.getActorName(c.Request.Context(), userID)
	ctrl.broadcastToChat(chatID, dto.WsEvent{
		Type: "group:member:added",
		Payload: mustMarshalCtrl(map[string]any{
			"chat_id":     chatID,
			"user_id":     userID,
			"user_name":   memberName,
			"by":          userID,
			"member_name": memberName,
		}),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Bạn đã tham gia nhóm thành công"})
}

func (ctrl *GroupChatController) RejectMemberRequest(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)
	chatID := c.Param("chatID")
	requestID := c.Param("requestID")

	if err := ctrl.groupService.RejectMemberRequest(c.Request.Context(), chatID, userID, requestID); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bạn đã từ chối lời mời tham gia nhóm"})
}

func (ctrl *GroupChatController) ListGroups(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)

	chats, err := ctrl.groupService.ListGroupChatsForUser(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, dto.GroupChatListResponse{Data: chats})
}

func (ctrl *GroupChatController) LeaveGroup(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)
	chatID := c.Param("chatID")

	var input struct {
		LeaveMode   string `json:"leave_mode"`
		HistoryMode string `json:"history_mode"`
	}
	_ = c.ShouldBindJSON(&input)

	if err := ctrl.groupService.LeaveGroup(c.Request.Context(), chatID, userID, input.LeaveMode, input.HistoryMode); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	if strings.EqualFold(strings.TrimSpace(input.LeaveMode), "public") {
		ctrl.broadcastToChat(chatID, dto.WsEvent{
			Type: "group:member:left",
			Payload: mustMarshalCtrl(map[string]any{
				"chat_id":    chatID,
				"user_id":    userID,
				"leave_mode": input.LeaveMode,
			}),
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã rời khỏi nhóm thành công"})
}

func (ctrl *GroupChatController) broadcastToChat(chatID string, event dto.WsEvent) {
	if ctrl.groupHub == nil {
		return
	}
	ctrl.groupHub.Broadcast(chatID, event)
}

func (ctrl *GroupChatController) getActorName(ctx context.Context, userID string) string {
	return ctrl.groupService.GetDisplayName(ctx, userID)
}

func mustMarshalCtrl(v any) []byte {
	out, _ := json.Marshal(v)
	return out
}
