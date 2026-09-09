package controllers

import (
	"fmt"
	errorsapp "linkup/errors"
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
	if err := c.ShouldBind(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodePostInvalidFormat))
		return
	}

	var files []*multipart.FileHeader
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		files = form.File["media"]
	}

	if err := validations.ValidateCreatePost(input.Title, input.Content, input.Status, len(files) > 0 || input.GifURL != ""); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := fmt.Sprintf("%v", val)

	post, err := ctrl.service.CreatePost(c.Request.Context(), userID, input.Title, input.Content, input.Status, input.CommunityID, files, input.GifURL)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
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
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	var nc *string
	if nextCursor != "" {
		nc = &nextCursor
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"page_size":   pageSize,
		"next_cursor": nc,
		"data":        posts,
	})
}

func (ctrl *PostController) GetSavedPosts(c *gin.Context) {
	cursor := c.Query("cursor")
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := fmt.Sprintf("%v", val)

	posts, nextCursor, err := ctrl.service.GetSavedPosts(c.Request.Context(), userID, cursor, pageSize)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
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

func (ctrl *PostController) GetUserPosts(c *gin.Context) {
	targetUserID := c.Param("userID")
	if targetUserID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	cursor := c.Query("cursor")
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var viewerID string
	if val, exists := c.Get("userID"); exists {
		viewerID = fmt.Sprintf("%v", val)
	}

	posts, nextCursor, err := ctrl.service.GetUserPosts(c.Request.Context(), targetUserID, viewerID, cursor, pageSize)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
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
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodePostIDRequired))
		return
	}

	post, err := ctrl.service.GetPostDetail(c.Request.Context(), postID)
	if err != nil {
		errorsapp.Respond(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": post})
}

func (ctrl *PostController) ReactPost(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodePostIDRequired))
		return
	}

	var input dto.ReactPostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeEmojiRequired))
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := fmt.Sprintf("%v", val)

	action, emojiCode, err := ctrl.service.ReactPost(c.Request.Context(), userID, postID, input.EmojiID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
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
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodePostIDRequired))
		return
	}

	var input dto.CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodePostInvalidFormat))
		return
	}

	if err := validations.ValidateCreateComment(input.Content); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := fmt.Sprintf("%v", val)

	comments, err := ctrl.service.CreateComment(c.Request.Context(), userID, postID, input.ParentID, input.Content)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
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
	sort := c.DefaultQuery("sort", "newest")

	var userIDPtr *string
	if val, exists := c.Get("userID"); exists {
		uid := fmt.Sprintf("%v", val)
		userIDPtr = &uid
	}

	comments, total, err := ctrl.service.GetCommentList(c.Request.Context(), postID, page, pageSize, sort, userIDPtr)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"data":      comments,
	})
}

func (ctrl *PostController) ToggleCommentReaction(c *gin.Context) {
	commentID := c.Param("commentID")
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := fmt.Sprintf("%v", val)

	var input struct {
		EmojiID string `json:"emoji_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	action, err := ctrl.service.ToggleCommentReaction(c.Request.Context(), userID, commentID, input.EmojiID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"action": action})
}

func (ctrl *PostController) SharePost(c *gin.Context) {
	postID := c.Param("id")
	val, _ := c.Get("userID")
	userID := fmt.Sprintf("%v", val)

	var input dto.SharePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		input.Content = ""
	}

	err := ctrl.service.SharePost(c.Request.Context(), userID, postID, input.Content)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
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
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	if action == "removed" {
		c.JSON(http.StatusOK, gin.H{"action": action, "message": "Đã bỏ lưu bài viết khỏi mục Bookmark!"})
	} else {
		c.JSON(http.StatusOK, gin.H{"action": action, "message": "Đã lưu bài viết vào mục Bookmark thành công!"})
	}
}

func (ctrl *PostController) DeletePost(c *gin.Context) {
	postID := c.Param("id")
	val, _ := c.Get("userID")
	userID := fmt.Sprintf("%v", val)

	if err := ctrl.service.DeletePost(c.Request.Context(), userID, postID); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa bài viết và cập nhật dữ liệu liên kết thành công!"})
}

func (ctrl *PostController) GetPostsByHashtag(c *gin.Context) {
	name := c.Param("name")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	posts, err := ctrl.service.GetPostsByHashtag(c.Request.Context(), name, page, pageSize)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
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
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": emojis})
}

func (ctrl *PostController) PinPost(c *gin.Context) {
	postID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	err := ctrl.service.PinPost(c.Request.Context(), userID.(string), postID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã ghim bài viết"})
}

func (ctrl *PostController) UnpinPost(c *gin.Context) {
	postID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	err := ctrl.service.UnpinPost(c.Request.Context(), userID.(string), postID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã bỏ ghim bài viết"})
}

func (ctrl *PostController) GetUserMedia(c *gin.Context) {
	targetUserID := c.Param("userID")
	if targetUserID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	media, total, err := ctrl.service.GetUserMedia(c.Request.Context(), targetUserID, page, pageSize)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      media,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"has_more":  int64(page*pageSize) < total,
	})
}
