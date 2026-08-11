package controllers

import (
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SearchController struct {
	searchService *services.SearchService
}

func NewSearchController(searchService *services.SearchService) *SearchController {
	return &SearchController{
		searchService: searchService,
	}
}

func (h *SearchController) GetTrending(c *gin.Context) {
	hashtags, err := h.searchService.GetTrendingHashtags(c.Request.Context())
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": hashtags})
}

func (h *SearchController) Search(c *gin.Context) {
	var input dto.SearchInput
	if err := c.ShouldBindQuery(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	response, err := h.searchService.Search(c.Request.Context(), input)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}

	c.JSON(http.StatusOK, response)
}
