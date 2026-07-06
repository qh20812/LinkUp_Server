package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"linkup/config"
	"linkup/dto"
	"linkup/services"
	"linkup/ws"

	"github.com/gin-gonic/gin"
)

type VoiceCallController struct {
	hub         *ws.Hub
	callService *services.VoiceCallService
	env         config.Env
}

func NewVoiceCallController(hub *ws.Hub, callService *services.VoiceCallService, env config.Env) *VoiceCallController {
	return &VoiceCallController{
		hub:         hub,
		callService: callService,
		env:         env,
	}
}

func (ctrl *VoiceCallController) HandleWebsocket(c *gin.Context) {
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

	client := ws.NewClient(c.Request.Context(), conn, ctrl.hub, nil, ctrl.callService, userID)
	ctrl.hub.RegisterClient(client)

	go client.WritePump()
	client.ReadPump()
}

func (ctrl *VoiceCallController) GetCallHistory(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	calls, total, err := ctrl.callService.GetCallHistory(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  calls,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}

func (ctrl *VoiceCallController) GetCallDetail(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))
	callID := c.Param("callID")

	if callID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "call_id là bắt buộc"})
		return
	}

	call, err := ctrl.callService.GetCallDetail(c.Request.Context(), userID, callID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": call})
}

func (ctrl *VoiceCallController) InitiateCall(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))

	var req dto.CallInitiatePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "callee_id và call_type là bắt buộc"})
		return
	}

	call, err := ctrl.callService.InitiateCall(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "call_id là bắt buộc"})
		return
	}

	if err := ctrl.callService.AcceptCall(c.Request.Context(), userID, callID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cuộc gọi đã được chấp nhận"})
}

func (ctrl *VoiceCallController) RejectCall(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))
	callID := c.Param("callID")

	if callID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "call_id là bắt buộc"})
		return
	}

	if err := ctrl.callService.RejectCall(c.Request.Context(), userID, callID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cuộc gọi đã bị từ chối"})
}

func (ctrl *VoiceCallController) ToggleMute(c *gin.Context) {
	userID := fmt.Sprintf("%v", c.GetString("userID"))
	callID := c.Param("callID")

	if callID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "call_id là bắt buộc"})
		return
	}

	var req dto.ToggleMuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "muted là bắt buộc"})
		return
	}

	if err := ctrl.callService.ToggleMute(c.Request.Context(), userID, callID, req.Muted); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã cập nhật trạng thái tắt tiếng"})
}
