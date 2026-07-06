package controllers

import (
	"fmt"
	"linkup/dto"
	"linkup/services"
	"linkup/validations"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	service services.PostService
}

func NewPostController(service services.PostService) *PostController {
	return &PostController{service: service}
}

func (ctrl *PostController) CreatePost(c *gin.Context) {
	var input dto.CreatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Định dạng JSON gửi lên không hợp lệ"})
		return
	}

	// Gọi Validation kiểm tra độ dài ký tự và trạng thái hợp lệ
	if err := validations.ValidateCreatePost(input.Title, input.Content, input.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin đăng nhập (Unauthorized)"})
		return
	}
	userID := fmt.Sprintf("%v", val)

	post, err := ctrl.service.CreatePost(c.Request.Context(), userID, input.Title, input.Content, input.Status, input.CommunityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": post})
}

func (ctrl *PostController) GetPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	posts, err := ctrl.service.GetPostList(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":      page,
		"page_size": pageSize,
		"data":      posts,
	})
}

func (ctrl *PostController) ViewPostDetail(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID bài viết không được trống"})
		return
	}

	post, err := ctrl.service.GetPostDetail(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": post})
}

func (ctrl *PostController) ReactPost(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID bài viết không được trống"})
		return
	}

	var input dto.ReactPostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu hoặc sai định dạng emoji_id"})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bạn cần đăng nhập để thực hiện tính năng này"})
		return
	}
	userID := fmt.Sprintf("%v", val)

	action, emojiCode, err := ctrl.service.ReactPost(c.Request.Context(), userID, postID, input.EmojiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi xử lý hệ thống: " + err.Error()})
		return
	}

	if action == "removed" {
		c.JSON(http.StatusOK, gin.H{
			"action":  action,
			"message": fmt.Sprintf("Đã gỡ bỏ cảm xúc %s khỏi bài viết thành công!", emojiCode),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"action":  action,
			"message": fmt.Sprintf("Đã thả cảm xúc %s vào bài viết thành công!", emojiCode),
		})
	}
}

func (ctrl *PostController) CreateComment(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID bài viết không hợp lệ"})
		return
	}

	var input dto.CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Định dạng JSON gửi lên không hợp lệ"})
		return
	}

	// Gọi Validation kiểm tra độ dài ký tự của bình luận
	if err := validations.ValidateCreateComment(input.Content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bạn cần đăng nhập để bình luận"})
		return
	}
	userID := fmt.Sprintf("%v", val)

	comments, err := ctrl.service.CreateComment(c.Request.Context(), userID, postID, input.ParentID, input.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Đăng bình luận thành công!",
		"data":    comments,
	})
}

func (ctrl *PostController) GetComments(c *gin.Context) {
	postID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	comments, err := ctrl.service.GetCommentList(c.Request.Context(), postID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":      page,
		"page_size": pageSize,
		"data":      comments,
	})
}

func (ctrl *PostController) SharePost(c *gin.Context) {
	postID := c.Param("id")
	val, _ := c.Get("userID")
	userID := fmt.Sprintf("%v", val)

	err := ctrl.service.SharePost(c.Request.Context(), userID, postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chia sẻ bài viết thành công!"})
}

func (ctrl *PostController) SavePost(c *gin.Context) {
	postID := c.Param("id")
	val, _ := c.Get("userID")
	userID := fmt.Sprintf("%v", val)

	err := ctrl.service.SavePost(c.Request.Context(), userID, postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã lưu bài viết!"})
}
