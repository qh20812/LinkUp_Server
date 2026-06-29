package controllers

import (
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TagController struct {
	tagService *services.TagService
}

func NewTagController(tagService *services.TagService) *TagController {
	return &TagController{tagService: tagService}
}

func (ctrl *TagController) GetPostsByHashtag(c *gin.Context) {
	tagName := c.Param("name")
	if tagName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng cung cấp tên hashtag cần tìm"})
		return
	}

	postIDs, err := ctrl.tagService.GetPostIDsByHashtag(c.Request.Context(), tagName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi tìm kiếm bài viết: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hashtag":  tagName,
		"post_ids": postIDs,
	})
}
