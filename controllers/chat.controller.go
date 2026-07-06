package controllers

import (
	"errors"
	"fmt"
	"linkup/config"
	"linkup/dto"
	"linkup/repository"
	"linkup/services"
	"linkup/validations"
	"linkup/ws"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatController struct {
	hub         *ws.Hub
	chatService *services.ChatService
	env         config.Env
}

func NewChatController(hub *ws.Hub, chatService *services.ChatService, env config.Env) *ChatController {
	return &ChatController{
		hub:         hub,
		chatService: chatService,
		env:         env,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ctrl *ChatController) HandleWebsocket(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "không có quyền truy cập"})
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "không thể nâng cấp kết nối websocket"})
		return
	}

	client := ws.NewClient(c.Request.Context(), conn, ctrl.hub, ctrl.chatService, nil, userID)
	ctrl.hub.RegisterClient(client)

	go client.WritePump()
	client.ReadPump()
}

func (ctrl *ChatController) CreateDirectChat(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	var input dto.DirectChatRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_user_id là bắt buộc"})
		return
	}

	chat, exists, err := ctrl.chatService.GetOrCreateDirectChat(c.Request.Context(), userID, input.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := "Tạo chat trực tiếp thành công"
	if exists {
		message = "Đã có chat trực tiếp giữa 2 người"
	}

	c.JSON(http.StatusOK, dto.DirectChatResponse{
		ChatID:  chat.ID,
		Message: message,
	})
}

func (ctrl *ChatController) CreateChatInvite(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	var input dto.ChatInviteRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_user_id là bắt buộc"})
		return
	}

	invite, err := ctrl.chatService.RequestChatInvite(c.Request.Context(), userID, input.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invite_id": invite.ID,
		"message":   "Yêu cầu chat đã được gửi",
	})
}

func (ctrl *ChatController) ResponseChatInvite(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	var input dto.ChatInviteResponseRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invite_id và accept là bắt buộc"})
		return
	}

	chat, err := ctrl.chatService.ResponseChatInvite(c.Request.Context(), userID, input.InviteID, input.Accept)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !input.Accept {
		c.JSON(http.StatusOK, dto.ChatInviteResponse{
			InviteID: input.InviteID,
			Message:  "Người dùng không chấp nhận chat",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ChatInviteResponse{
		InviteID: input.InviteID,
		ChatID:   &chat.ID,
		Message:  "Chat đã được kích hoạt",
	})
}

func (ctrl *ChatController) DownloadMessageMedia(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))
	messageID := c.Param("messageID")

	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message_id là bắt buộc"})
		return
	}

	media, contentType, filename, data, err := ctrl.chatService.DownloadMessageMedia(c.Request.Context(), userID, messageID)
	if err != nil {
		switch {
		case errors.Is(err, validations.ErrMessageNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Tin nhắn không tồn tại"})
		case errors.Is(err, validations.ErrMessageAccessDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền truy cập tin nhắn này"})
		case errors.Is(err, validations.ErrMediaNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Media không tồn tại"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	c.Data(http.StatusOK, contentType, data)

	_ = media
}

func (ctrl *ChatController) DeleteChat(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))
	chatID := c.Param("chatID")

	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chat_id là bắt buộc"})
		return
	}

	err := ctrl.chatService.DeleteChat(c.Request.Context(), userID, chatID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrChatNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "không tìm thấy phòng chat"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": "Xóa phòng chat thành công"})
}
