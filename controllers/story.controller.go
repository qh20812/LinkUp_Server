package controllers

import (
	"linkup/dto"
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

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng chọn file hình ảnh hoặc video"})
		return
	}

	caption := c.PostForm("caption")

	res, err := ctrl.service.CreateStory(userID.(string), file, caption)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// GetHomeFeed lấy danh sách story trang chủ
func (ctrl *StoryController) GetHomeFeed(c *gin.Context) {
	res, err := ctrl.service.GetHomeStories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu tương tác không hợp lệ"})
		return
	}

	err := ctrl.service.InteractWithStory(storyID, userID.(string), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}
