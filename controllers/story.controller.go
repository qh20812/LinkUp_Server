package controllers

import (
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StoryController struct {
	service services.StoryService
}

func NewStoryController(service services.StoryService) *StoryController {
	return &StoryController{service: service}
}

// CreateStory tiếp nhận dữ liệu đăng tải file dạng Form-Data
func (ctrl *StoryController) CreateStory(c *gin.Context) {
	userID, _ := c.Get("userID")

	// Lấy file, nếu không có (err != nil) thì file sẽ là nil, ta không return lỗi nữa
	file, _ := c.FormFile("file")
	caption := c.PostForm("caption")

	res, err := ctrl.service.CreateStory(c.Request.Context(), userID.(string), file, caption)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, res)
}

// GetHomeFeed lấy danh sách story trang chủ
func (ctrl *StoryController) GetHomeFeed(c *gin.Context) {
	res, err := ctrl.service.GetHomeStories()
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// ViewStory ghi nhận xem và hiển thị thông tin chi tiết
func (ctrl *StoryController) ViewStory(c *gin.Context) {
	viewerID, _ := c.Get("userID")
	storyID := c.Param("id")

	story, err := ctrl.service.ViewStory(storyID, viewerID.(string))
	if err != nil {
		errorsapp.Respond(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, story)
}

// Interact xử lý gửi cảm xúc, tin nhắn hoặc chia sẻ
func (ctrl *StoryController) Interact(c *gin.Context) {
	userID, _ := c.Get("userID")
	storyID := c.Param("id")

	var req dto.InteractStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	err := ctrl.service.InteractWithStory(storyID, userID.(string), req)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gửi tương tác thành công"})
}

// GetAnalytics xem dữ liệu thống kê lượt xem/tương tác
func (ctrl *StoryController) GetAnalytics(c *gin.Context) {
	userID, _ := c.Get("userID")
	storyID := c.Param("id")

	analytics, err := ctrl.service.GetAnalytics(storyID, userID.(string))
	if err != nil {
		errorsapp.Respond(c, http.StatusForbidden, err)
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// CheckUserStory kiểm tra user có story active không
func (ctrl *StoryController) CheckUserStory(c *gin.Context) {
	targetUserID := c.Param("userID")
	if targetUserID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	has, err := ctrl.service.HasActiveStory(targetUserID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"has_story": has})
}

// GetUserStories lấy danh sách story active của 1 user
func (ctrl *StoryController) GetUserStories(c *gin.Context) {
	targetUserID := c.Param("userID")
	if targetUserID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	viewerID, _ := c.Get("userID")
	viewerStr, _ := viewerID.(string)

	stories, err := ctrl.service.GetUserActiveStories(targetUserID, viewerStr)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"stories": stories})
}
