package controllers

import (
	"fmt"
	"linkup/dto"
	"linkup/services"
	"linkup/validations"
	"mime/multipart"
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
	// Sử dụng ShouldBind để xử lý đồng thời text fields và file upload dạng form-data
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Định dạng dữ liệu form-data gửi lên không hợp lệ"})
		return
	}

	// Lấy danh sách tệp đính kèm gửi lên qua form-data với key đặt tên là "media"
		var files []*multipart.FileHeader
		form, err := c.MultipartForm()
		if err == nil && form != nil {
			files = form.File["media"]
		}

		if err := validations.ValidateCreatePost(input.Title, input.Content, input.Status, len(files) > 0); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		val, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin đăng nhập (Unauthorized)"})
			return
		}
		userID := fmt.Sprintf("%v", val)

	post, err := ctrl.service.CreatePost(c.Request.Context(), userID, input.Title, input.Content, input.Status, input.CommunityID, files)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": post})
}

func (ctrl *PostController) GetPosts(c *gin.Context) {
	cursor := c.Query("cursor")
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	filter := c.Query("filter")

	if filter == "following" && pageSize > 5 {
		pageSize = 5
	}

	userID, _ := c.Get("userID")
	userIDStr, _ := userID.(string)

	posts, nextCursor, err := ctrl.service.GetPostList(c.Request.Context(), cursor, pageSize, userIDStr, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var nc *string
	if nextCursor != "" {
		nc = &nextCursor
	}

	c.JSON(http.StatusOK, gin.H{
		"page_size":   pageSize,
		"next_cursor": nc,
		"data":        posts,
	})
}

// Lấy danh sách bài viết đã lưu (Bookmark) của người dùng hiện tại
func (ctrl *PostController) GetSavedPosts(c *gin.Context) {
	cursor := c.Query("cursor")
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bạn cần đăng nhập để xem bài viết đã lưu"})
		return
	}
	userID := fmt.Sprintf("%v", val)

	posts, nextCursor, err := ctrl.service.GetSavedPosts(c.Request.Context(), userID, cursor, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var nc *string
	if nextCursor != "" {
		nc = &nextCursor
	}

	c.JSON(http.StatusOK, gin.H{
		"page_size":   pageSize,
		"next_cursor": nc,
		"data":        posts,
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

	comments, total, err := ctrl.service.GetCommentList(c.Request.Context(), postID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"data":      comments,
	})
}

func (ctrl *PostController) SharePost(c *gin.Context) {
	postID := c.Param("id")
	val, _ := c.Get("userID")
	userID := fmt.Sprintf("%v", val)

	var input dto.SharePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		input.Content = "" // Nếu không viết Caption thì mặc định rỗng
	}

	err := ctrl.service.SharePost(c.Request.Context(), userID, postID, input.Content)
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

	action, err := ctrl.service.SavePost(c.Request.Context(), userID, postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if action == "removed" {
		c.JSON(http.StatusOK, gin.H{"action": action, "message": "Đã bỏ lưu bài viết khỏi mục Bookmark!"})
	} else {
		c.JSON(http.StatusOK, gin.H{"action": action, "message": "Đã lưu bài viết vào mục Bookmark thành công!"})
	}
}

// Xóa bài viết
func (ctrl *PostController) DeletePost(c *gin.Context) {
	postID := c.Param("id")
	val, _ := c.Get("userID")
	userID := fmt.Sprintf("%v", val)

	if err := ctrl.service.DeletePost(c.Request.Context(), userID, postID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa bài viết và cập nhật dữ liệu liên kết thành công!"})
}

// Lấy danh sách bài viết theo thẻ Hashtag
func (ctrl *PostController) GetPostsByHashtag(c *gin.Context) {
	name := c.Param("name")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	posts, err := ctrl.service.GetPostsByHashtag(c.Request.Context(), name, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hashtag":   name,
		"page":      page,
		"page_size": pageSize,
		"data":      posts,
	})
}

func (ctrl *PostController) GetEmojis(c *gin.Context) {
	emojis, err := ctrl.service.ListEmojis(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": emojis})
}
