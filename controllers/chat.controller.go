package controllers

import (
	"errors"
	"fmt"
	"linkup/config"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/repository"
	"linkup/services"
	"linkup/utils"
	"linkup/ws"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatController struct {
	hub         *ws.Hub
	chatService *services.ChatService
	postRepo    *repository.PostRepository
	env         config.Env
}

func NewChatController(hub *ws.Hub, chatService *services.ChatService, postRepo *repository.PostRepository, env config.Env) *ChatController {
	return &ChatController{
		hub:         hub,
		chatService: chatService,
		postRepo:    postRepo,
		env:         env,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ctrl *ChatController) HandleWebsocket(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}

	token, err := utils.ParseToken(ctrl.env.JWTSecret, tokenString)
	if err != nil || !token.Valid {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeInvalidToken))
		return
	}

	claims := token.Claims.(*utils.TokenClaims)
	if claims.TokenType != "access" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeInvalidToken))
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInternal))
		return
	}

	client := ws.NewClient(c.Request.Context(), conn, ctrl.hub, ctrl.chatService, nil, ctrl.postRepo, claims.UserID)
	ctrl.hub.RegisterClient(client)

	go client.WritePump()
	client.ReadPump()
}

func (ctrl *ChatController) ListChats(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	chats, err := ctrl.chatService.ListUserChats(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, dto.ChatListResponse{Data: chats})
}

func (ctrl *ChatController) CreateDirectChat(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	var input dto.DirectChatRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	chat, exists, err := ctrl.chatService.GetOrCreateDirectChat(c.Request.Context(), userID, input.TargetUserID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
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
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	invite, err := ctrl.chatService.RequestChatInvite(c.Request.Context(), userID, input.TargetUserID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invite_id": invite.ID,
		"message":   "Yêu cầu chat đã được gửi",
	})
}

func (ctrl *ChatController) ListChatInvites(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	invites, err := ctrl.chatService.ListReceivedInvites(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, dto.ChatInviteListResponse{Data: invites})
}

func (ctrl *ChatController) ResponseChatInvite(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	var input dto.ChatInviteResponseRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	chat, err := ctrl.chatService.ResponseChatInvite(c.Request.Context(), userID, input.InviteID, input.Accept)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
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
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	media, contentType, filename, data, err := ctrl.chatService.DownloadMessageMedia(c.Request.Context(), userID, messageID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusInternalServerError, err)
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
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	err := ctrl.chatService.DeleteChat(c.Request.Context(), userID, chatID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrChatNotFound):
			errorsapp.RespondError(c, http.StatusNotFound, errorsapp.New(errorsapp.ErrCodeNotFound))
		default:
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": "Xóa phòng chat thành công"})
}

type SharePostInput struct {
	TargetUserID string `json:"target_user_id" binding:"required"`
	SharedPostID string `json:"shared_post_id" binding:"required"`
}

func (ctrl *ChatController) SharePost(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	var input SharePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	chat, _, err := ctrl.chatService.GetOrCreateDirectChat(c.Request.Context(), userID, input.TargetUserID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	sharedPostID := input.SharedPostID
	msg, err := ctrl.chatService.SendMessage(c.Request.Context(), userID, chat.ID, "", 0, nil, nil, nil, nil, &sharedPostID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Chia sẻ bài viết thành công",
		"data":    msg,
	})
}
