package controllers

import (
	errorsapp "linkup/errors"
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
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	postIDs, err := ctrl.tagService.GetPostIDsByHashtag(c.Request.Context(), tagName)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hashtag":  tagName,
		"post_ids": postIDs,
	})
}
