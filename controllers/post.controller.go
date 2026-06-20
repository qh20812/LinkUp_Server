package controllers

import (
	"fmt"
	"linkup/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Đổi từ PostHandler thành PostController
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

func (ctrl *PostController) CreatePost(c *gin.Context) {
    var input CreatePostInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 🌟 SỬA ĐOẠN NÀY: Lấy userID thật do Middleware JWT lưu lại sau khi đăng nhập
    // Tùy thuộc vào cách bạn đặt tên key trong AuthMiddleware (thường là "userId" hoặc "userID")
    val, exists := c.Get("userId") 
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin đăng nhập (Unauthorized)"})
        return
    }

    // Ép kiểu val về dạng chuỗi hoặc số tùy thuộc vào Service của bạn nhận kiểu gì.
    // Nếu hàm ctrl.service.CreatePost nhận tham số userID là chuỗi (string):
    userID := fmt.Sprintf("%v", val) 

    // Nếu hàm service nhận userID là số (int), bạn dùng đoạn dưới này:
    // userID := val.(int)

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
