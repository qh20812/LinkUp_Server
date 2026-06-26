package controllers

import (
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FriendController struct {
	friendService *services.FriendService
}

func NewFriendController(friendService *services.FriendService) *FriendController {
	return &FriendController{
		friendService: friendService,
	}
}

func (ctrl *FriendController) ToggleFriendRequest(c *gin.Context) {
	targetUserID := c.Param("userID")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID là bắt buộc"})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bạn cần đăng nhập"})
		return
	}
	userID := val.(string)

	response, err := ctrl.friendService.ToggleFriendRequest(c.Request.Context(), userID, targetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *FriendController) GetFriendRequests(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bạn cần đăng nhập"})
		return
	}
	userID := val.(string)

	response, err := ctrl.friendService.GetFriendRequests(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *FriendController) AcceptFriendRequest(c *gin.Context) {
	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requestID là bắt buộc"})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bạn cần đăng nhập"})
		return
	}
	userID := val.(string)

	response, err := ctrl.friendService.AcceptFriendRequest(c.Request.Context(), userID, requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *FriendController) RejectFriendRequest(c *gin.Context) {
	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requestID là bắt buộc"})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bạn cần đăng nhập"})
		return
	}
	userID := val.(string)

	response, err := ctrl.friendService.RejectFriendRequest(c.Request.Context(), userID, requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
