package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/services"
)

type PresenceController struct {
	presenceService *services.PresenceService
}

func NewPresenceController(presenceService *services.PresenceService) *PresenceController {
	return &PresenceController{presenceService: presenceService}
}

// GetPresence returns the presence data for a specific user.
func (h *PresenceController) GetPresence(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	targetUserID := c.Param("userID")
	if targetUserID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	presence, err := h.presenceService.GetPresence(userID, targetUserID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, presence)
}

// BatchGetPresence returns presence data for multiple users.
func (h *PresenceController) BatchGetPresence(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	var input dto.BatchPresenceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if len(input.UserIDs) == 0 {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	presenceMap, err := h.presenceService.BatchGetPresence(userID, input.UserIDs)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": presenceMap})
}

// GetOnlineUsers returns a list of online users.
func (h *PresenceController) GetOnlineUsers(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	onlineUsers := h.presenceService.GetOnlineUsers()
	c.JSON(http.StatusOK, gin.H{"data": onlineUsers})
}

// GetOnlineCount returns the number of online users.
func (h *PresenceController) GetOnlineCount(c *gin.Context) {
	count := h.presenceService.GetOnlineUserCount()
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// GetPresenceSettings returns the current user's presence settings.
func (h *PresenceController) GetPresenceSettings(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	activityStatusEnabled, lastSeenVisibility, err := h.presenceService.GetPresenceSettings(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activity_status_enabled": activityStatusEnabled,
		"last_seen_visibility":    lastSeenVisibility,
	})
}

// UpdatePresenceSettings updates the current user's presence settings.
func (h *PresenceController) UpdatePresenceSettings(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	var input dto.UpdatePresenceSettingsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	// Get current settings to fill in missing values
	currentEnabled, currentVisibility, err := h.presenceService.GetPresenceSettings(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	// Use provided values or fall back to current values
	enabled := currentEnabled
	visibility := currentVisibility

	if input.ActivityStatusEnabled != nil {
		enabled = *input.ActivityStatusEnabled
	}
	if input.LastSeenVisibility != nil {
		visibility = *input.LastSeenVisibility
	}

	if err := h.presenceService.UpdatePresenceSettings(c.Request.Context(), userID, enabled, visibility); err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activity_status_enabled": enabled,
		"last_seen_visibility":    visibility,
	})
}
