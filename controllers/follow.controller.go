package controllers

import (
	"linkup/dto"
	"linkup/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FollowController struct {
	followService *services.FollowService
}

func NewFollowController(followService *services.FollowService) *FollowController {
	return &FollowController{
		followService: followService,
	}
}

func (crtl *FollowController) FollowToggle(c *gin.Context) {
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
	followerID := val.(string)

	action, err := crtl.followService.FollowToggle(c.Request.Context(), followerID, targetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stats, err := crtl.followService.GetFollowerStats(c.Request.Context(), targetUserID, &followerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lỗi khi lấy thống kê"})
		return
	}

	response := dto.FollowToggleResponse{
		Action:         action,
		IsFollowing:    action == "followed",
		FollowerCount:  stats["follower_count"].(int64),
		FollowingCount: stats["following_count"].(int64),
		Message:        "Cập nhật thành công",
	}

	c.JSON(http.StatusOK, response)
}

func (crtl *FollowController) GetFollowStats(c *gin.Context) {
	targetUserID := c.Param("userID")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID là bắt buộc"})
		return
	}

	var viewerID *string
	if val, exists := c.Get("userID"); exists {
		viewerIDStr := val.(string)
		viewerID = &viewerIDStr
	}

	stats, err := crtl.followService.GetFollowerStats(c.Request.Context(), targetUserID, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.FollowStatsResponse{
		FollowerCount:  stats["follower_count"].(int64),
		FollowingCount: stats["following_count"].(int64),
	}

	if isFollowing, ok := stats["is_following"]; ok {
		response.IsFollowing = isFollowing.(bool)
	}

	c.JSON(http.StatusOK, response)
}

func (crtl *FollowController) GetSuggestions(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bạn cần đăng nhập"})
		return
	}
	userID := val.(string)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "5"))

	response, err := crtl.followService.GetSuggestions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
