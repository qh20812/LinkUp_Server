package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"linkup/config"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/repository"
	"linkup/services"
	"linkup/utils"
	"linkup/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type VoiceCallController struct {
	hub         *ws.Hub
	callService *services.VoiceCallService
	env         config.Env
}

var callUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewVoiceCallController(hub *ws.Hub, callService *services.VoiceCallService, env config.Env) *VoiceCallController {
	return &VoiceCallController{
		hub:         hub,
		callService: callService,
		env:         env,
	}
}

func (ctrl *VoiceCallController) GetIceServers(c *gin.Context) {
	servers := make([]dto.IceServer, 0)

	// Parse STUN server URLs từ env (comma-separated)
	if ctrl.env.IceServerUrls != "" {
		urls := strings.Split(ctrl.env.IceServerUrls, ",")
		for _, url := range urls {
			url = strings.TrimSpace(url)
			if url != "" {
				servers = append(servers, dto.IceServer{URLs: url})
			}
		}
	}

	// Nếu có TURN server config, append vào danh sách
	if ctrl.env.TurnServerUrl != "" {
		servers = append(servers, dto.IceServer{
			URLs:       ctrl.env.TurnServerUrl,
			Username:   ctrl.env.TurnUsername,
			Credential: ctrl.env.TurnCredential,
		})
	}

	c.JSON(http.StatusOK, dto.IceServersResponse{IceServers: servers})
}

func (ctrl *VoiceCallController) HandleWebsocket(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	var userID string
	if !exists {
		tokenString := c.Query("token")
		if tokenString == "" {
			errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
			return
		}
		token, err := utils.ParseToken(ctrl.env.JWTSecret, tokenString)
		if err != nil || !token.Valid {
			errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeInvalidInput))
			return
		}
		claims := token.Claims.(*utils.TokenClaims)
		if claims.TokenType != "access" {
			errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeInvalidInput))
			return
		}
		userID = claims.UserID
	} else {
		userID = fmt.Sprintf("%v", userIDVal)
	}

	conn, err := callUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInternal))
		return
	}

	client := ws.NewClient(c.Request.Context(), conn, ctrl.hub, nil, ctrl.callService, nil, userID)
	ctrl.hub.RegisterClient(client)

	go client.WritePump()
	client.ReadPump()
}

// queryToFilter converts the Gin-bound query DTO into a repository filter.
// Whitelists and validates values here so the service/repo never see raw user input.
func queryToFilter(q dto.CallHistoryQuery) repository.CallHistoryFilter {
	f := repository.CallHistoryFilter{
		Limit:  q.Limit,
		Offset: q.Offset,
	}
	// Whitelist call type.
	if q.Type != nil {
		t := strings.ToLower(strings.TrimSpace(*q.Type))
		if t == "voice" || t == "video" {
			f.CallType = &t
		}
	}
	// Whitelist call status.
	if q.Status != nil {
		s := strings.ToLower(strings.TrimSpace(*q.Status))
		switch s {
		case "missed", "ended", "rejected", "connected", "calling", "ringing":
			f.Status = &s
		}
	}
	// Whitelist sort column (repository also validates, defense-in-depth).
	sort := strings.ToLower(strings.TrimSpace(q.Sort))
	switch sort {
	case "created_at", "duration", "call_type", "status":
		f.Sort = sort
	default:
		f.Sort = "created_at"
	}
	// Whitelist order direction.
	order := strings.ToLower(strings.TrimSpace(q.Order))
	if order == "asc" || order == "desc" {
		f.Order = order
	} else {
		f.Order = "desc"
	}
	return f
}

func (ctrl *VoiceCallController) GetCallHistory(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	var query dto.CallHistoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}
	// Enforce safe defaults after binding.
	if query.Limit < 1 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	f := queryToFilter(query)

	items, total, err := ctrl.callService.GetCallHistoryFiltered(c.Request.Context(), userID, f)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   items,
		"total":  total,
		"limit":  f.Limit,
		"offset": f.Offset,
	})
}

func (ctrl *VoiceCallController) GetCallDetail(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))
	callID := c.Param("callID")

	if callID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	call, err := ctrl.callService.GetCallDetail(c.Request.Context(), userID, callID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": call})
}

func (ctrl *VoiceCallController) InitiateCall(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	var req dto.CallInitiatePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	call, err := ctrl.callService.InitiateCall(c.Request.Context(), userID, req)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}
	if call == nil {
		c.JSON(http.StatusOK, gin.H{"message": "người dùng đang bận"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": call})
}

func (ctrl *VoiceCallController) AcceptCall(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))
	callID := c.Param("callID")

	if callID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := ctrl.callService.AcceptCall(c.Request.Context(), userID, callID); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cuộc gọi đã được chấp nhận"})
}

func (ctrl *VoiceCallController) RejectCall(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))
	callID := c.Param("callID")

	if callID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := ctrl.callService.RejectCall(c.Request.Context(), userID, callID); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cuộc gọi đã bị từ chối"})
}

func (ctrl *VoiceCallController) ToggleVideo(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))
	callID := c.Param("callID")

	if callID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	var req dto.ToggleVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := ctrl.callService.ToggleVideo(c.Request.Context(), userID, callID, req.VideoEnabled); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã cập nhật trạng thái video"})
}

func (ctrl *VoiceCallController) ToggleMute(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))
	callID := c.Param("callID")

	if callID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	var req dto.ToggleMuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := ctrl.callService.ToggleMute(c.Request.Context(), userID, callID, req.Muted); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã cập nhật trạng thái tắt tiếng"})
}

// HideCall removes a call from the user's history view (soft-delete per user).
func (ctrl *VoiceCallController) HideCall(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))
	callID := c.Param("callID")

	if callID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := ctrl.callService.HideCallFromHistory(c.Request.Context(), userID, callID); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cuộc gọi đã được ẩn khỏi lịch sử"})
}

// GetMissedCallCount returns the number of unread missed calls.
func (ctrl *VoiceCallController) GetMissedCallCount(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	count, err := ctrl.callService.GetMissedCallCount(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// MarkMissedRead marks all current missed calls as read for the user.
func (ctrl *VoiceCallController) MarkMissedRead(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	if err := ctrl.callService.MarkMissedAsRead(c.Request.Context(), userID); err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã đánh dấu đã đọc"})
}
