package controllers

import (
	"fmt"
	"linkup/services"
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

type CreatePostInput struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type ReactPostInput struct {
	EmojiID string `json:"emoji_id" binding:"required"`
}

func (ctrl *PostController) CreatePost(c *gin.Context) {
	var input CreatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	val, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin đăng nhập (Unauthorized)"})
		return
	}
	userID := fmt.Sprintf("%v", val)

	post, err := ctrl.service.CreatePost(c.Request.Context(), userID, input.Title, input.Content)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy bài viết hoặc bài viết đã bị ẩn"})
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

	var input ReactPostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu hoặc sai định dạng emoji_id"})
		return
	}

	val, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bạn cần đăng nhập để thực hiện tính năng này"})
		return
	}
	userID := fmt.Sprintf("%v", val)

	// Gọi service xử lý logic DB
	action, err := ctrl.service.ReactPost(c.Request.Context(), userID, postID, input.EmojiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi xử lý hệ thống: " + err.Error()})
		return
	}

	// Bản đồ tra cứu từ UUID sang Tên hiển thị sinh động
	emojiMap := map[string]string{
		"2de88c4e-c8e7-4547-a1f7-dffee3ee65f2": ":rocket:",
		"5cc37cf0-0689-44a2-bfd0-a46bf2c667fe": ":heart:",
		"73b0c7db-e304-4603-a18f-320bfc30b4d4": ":sad:",
		"87f4024d-9903-445a-b087-e04a7dc1741d": ":haha:",
		"9445f716-43c6-4d6e-89da-eb80c1cd5827": ":clap:",
		"aab3f1e7-f6b6-430b-b09e-9230d0b42a57": ":wow:",
		"ade30341-0678-42e7-82cf-1664bddf57ee": ":love:",
		"b49ad59d-1364-40fa-b354-8e37946ca2fe": ":fire:",
		"bd3d43db-7e48-4471-a548-da9884723124": ":angry:",
		"ed740b65-d22b-4536-9278-2d0ef72df739": ":like:",
	}

	// Lấy tên emoji tương ứng, nếu không tìm thấy thì để mặc định là "cảm xúc"
	emojiName, found := emojiMap[input.EmojiID]
	if !found {
		emojiName = "cảm xúc"
	}

	// Phản hồi kết quả rõ ràng lên Postman dựa trên hành động
	if action == "removed" {
		c.JSON(http.StatusOK, gin.H{
			"action":  action,
			"message": fmt.Sprintf("Đã gỡ bỏ cảm xúc %s khỏi bài viết thành công!", emojiName),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"action":  action,
			"message": fmt.Sprintf("Đã thả cảm xúc %s vào bài viết thành công!", emojiName),
		})
	}
}