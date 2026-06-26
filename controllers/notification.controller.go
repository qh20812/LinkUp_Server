package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"linkup/dto"
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
}

func (ctrl *NotificationController) GetNotifications(c *gin.Context) {
	userID, _ := c.Get("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	unreadOnly, _ := strconv.ParseBool(c.DefaultQuery("unreadOnly", "false"))

	notifications, total, err := ctrl.service.GetList(c.Request.Context(), userID.(string), page, pageSize, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  dto.ToNotificationResponseList(notifications),
		"total": total,
		"page":  page,
	})
}

func (ctrl *NotificationController) MarkAsRead(c *gin.Context) {
	userID, _ := c.Get("userID")
	notifID := c.Param("id")

	if err := ctrl.service.MarkAsRead(c.Request.Context(), userID.(string), notifID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *NotificationController) MarkAllAsRead(c *gin.Context) {
	userID, _ := c.Get("userID")

	if err := ctrl.service.MarkAllAsRead(c.Request.Context(), userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *NotificationController) GetUnreadCount(c *gin.Context) {
	userID, _ := c.Get("userID")

	count, err := ctrl.service.GetUnreadCount(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (ctrl *NotificationController) GetPreferences(c *gin.Context) {
	userID, _ := c.Get("userID")

	pref, err := ctrl.service.GetPreferences(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pref})
}

func (ctrl *NotificationController) UpdatePreferences(c *gin.Context) {
	userID, _ := c.Get("userID")

	var input UpdatePreferencesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if input.LikeEnabled == nil && input.CommentEnabled == nil && input.FollowEnabled == nil && input.MessageEnabled == nil && input.FriendRequestEnabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	pref := &models.NotificationPreference{
		UserID: userID.(string),
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

	if err := ctrl.service.UpdatePreferences(c.Request.Context(), pref); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
