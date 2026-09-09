package controllers

import (
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/services"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StoryController struct {
	service services.StoryService
}

func NewStoryController(service services.StoryService) *StoryController {
	return &StoryController{service: service}
}

// CreateStory tiếp nhận dữ liệu đăng tải một hoặc nhiều file dạng Form-Data
// - `file`: lặp lại nhiều lần cho mỗi ảnh/video (có thể không có nếu là story text)
// - `captions`: mảng caption, khớp theo thứ tự với `file`
// - `caption`: caption đơn (fallback khi chỉ có 1 story / story text)
func (ctrl *StoryController) CreateStory(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid := userID.(string)

	var files []*multipart.FileHeader
	if form, err := c.MultipartForm(); err == nil {
		files = form.File["file"]
	}
	if len(files) == 0 {
		if f, err := c.FormFile("file"); err == nil {
			files = append(files, f)
		}
	}

	captions := c.PostFormArray("captions")
	if len(captions) == 0 {
		if single := c.PostForm("caption"); single != "" {
			captions = []string{single}
		}
	}

	if len(files) == 0 && len(captions) == 0 {
		errorsapp.Respond(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeStoryContentRequired))
		return
	}

	if len(files) == 0 {
		// Story dạng text (không có media)
		res, err := ctrl.service.CreateStory(c.Request.Context(), uid, nil, captions[0])
		if err != nil {
			errorsapp.Respond(c, http.StatusBadRequest, err)
			return
		}
		c.JSON(http.StatusCreated, []dto.CreateStoryResponse{*res})
		return
	}

	// Nhiều story: tạo một story row cho từng file
	results := make([]dto.CreateStoryResponse, 0, len(files))
	for i, file := range files {
		capText := ""
		if i < len(captions) {
			capText = captions[i]
		}
		res, err := ctrl.service.CreateStory(c.Request.Context(), uid, file, capText)
		if err != nil {
			errorsapp.Respond(c, http.StatusBadRequest, err)
			return
		}
		results = append(results, *res)
	}

	c.JSON(http.StatusCreated, results)
}

// GetHomeFeed lấy danh sách story trang chủ (lọc theo viewer: muted/blocked/following)
func (ctrl *StoryController) GetHomeFeed(c *gin.Context) {
	viewerID, _ := c.Get("userID")
	viewerStr, _ := viewerID.(string)

	scope := c.DefaultQuery("scope", "all")
	followingOnly := scope == "following"

	res, err := ctrl.service.GetHomeStories(viewerStr, followingOnly)
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

// DeleteStory xóa story của chính mình (cascade story_views / story_interacts)
func (ctrl *StoryController) DeleteStory(c *gin.Context) {
	userID, _ := c.Get("userID")
	storyID := c.Param("id")

	err := ctrl.service.DeleteStory(c.Request.Context(), storyID, userID.(string))
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
			return
		}
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xóa story thành công"})
}

// ToggleMute bật/tắt ẩn tất cả story của 1 user
func (ctrl *StoryController) ToggleMute(c *gin.Context) {
	userID, _ := c.Get("userID")
	targetUserID := c.Param("userID")

	muted, err := ctrl.service.ToggleStoryMute(userID.(string), targetUserID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"muted": muted})
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
