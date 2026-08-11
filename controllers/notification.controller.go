package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/services"
)

type NotificationController struct {
	service *services.NotificationService
}

func NewNotificationController(service *services.NotificationService) *NotificationController {
	return &NotificationController{service: service}
}

type UpdatePreferencesInput struct {
	LikeEnabled          *bool `json:"like_enabled"`
	CommentEnabled       *bool `json:"comment_enabled"`
	FollowEnabled        *bool `json:"follow_enabled"`
	MessageEnabled       *bool `json:"message_enabled"`
	FriendRequestEnabled *bool `json:"friend_request_enabled"`
	CommunityEnabled     *bool `json:"community_enabled"`
	VoiceCallEnabled     *bool `json:"voice_call_enabled"`
}

func (ctrl *NotificationController) GetNotifications(c *gin.Context) {
	userID, _ := c.Get("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	unreadOnly, _ := strconv.ParseBool(c.DefaultQuery("unreadOnly", "false"))

	notifications, total, err := ctrl.service.GetList(c.Request.Context(), userID.(string), page, pageSize, unreadOnly)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  notifications,
		"total": total,
		"page":  page,
	})
}

func (ctrl *NotificationController) MarkAsRead(c *gin.Context) {
	userID, _ := c.Get("userID")
	notifID := c.Param("id")

	if err := ctrl.service.MarkAsRead(c.Request.Context(), userID.(string), notifID); err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *NotificationController) MarkAllAsRead(c *gin.Context) {
	userID, _ := c.Get("userID")

	if err := ctrl.service.MarkAllAsRead(c.Request.Context(), userID.(string)); err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *NotificationController) GetUnreadCount(c *gin.Context) {
	userID, _ := c.Get("userID")

	count, err := ctrl.service.GetUnreadCount(c.Request.Context(), userID.(string))
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (ctrl *NotificationController) GetPreferences(c *gin.Context) {
	userID, _ := c.Get("userID")

	pref, err := ctrl.service.GetPreferences(c.Request.Context(), userID.(string))
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pref})
}

func (ctrl *NotificationController) UpdatePreferences(c *gin.Context) {
	userID, _ := c.Get("userID")

	var input UpdatePreferencesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if input.LikeEnabled == nil && input.CommentEnabled == nil && input.FollowEnabled == nil && input.MessageEnabled == nil && input.FriendRequestEnabled == nil && input.CommunityEnabled == nil && input.VoiceCallEnabled == nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	pref, err := ctrl.service.GetPreferences(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if pref == nil {
		// No row yet = all types enabled (matches Create's nil-pref behavior).
		pref = &models.NotificationPreference{
			UserID:               userID.(string),
			LikeEnabled:          true,
			CommentEnabled:       true,
			FollowEnabled:        true,
			MessageEnabled:       true,
			FriendRequestEnabled: true,
			CommunityEnabled:     true,
			VoiceCallEnabled:     true,
		}
	}
	if input.LikeEnabled != nil {
		pref.LikeEnabled = *input.LikeEnabled
	}
	if input.CommentEnabled != nil {
		pref.CommentEnabled = *input.CommentEnabled
	}
	if input.FollowEnabled != nil {
		pref.FollowEnabled = *input.FollowEnabled
	}
	if input.MessageEnabled != nil {
		pref.MessageEnabled = *input.MessageEnabled
	}
	if input.FriendRequestEnabled != nil {
		pref.FriendRequestEnabled = *input.FriendRequestEnabled
	}
	if input.CommunityEnabled != nil {
		pref.CommunityEnabled = *input.CommunityEnabled
	}
	if input.VoiceCallEnabled != nil {
		pref.VoiceCallEnabled = *input.VoiceCallEnabled
	}

	if err := ctrl.service.UpdatePreferences(c.Request.Context(), pref); err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
