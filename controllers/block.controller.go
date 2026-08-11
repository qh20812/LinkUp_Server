package controllers

import (
	errorsapp "linkup/errors"
	"linkup/dto"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BlockController struct {
	blockService *services.BlockService
}

func NewBlockController(blockService *services.BlockService) *BlockController {
	return &BlockController{
		blockService: blockService,
	}
}

func (h *BlockController) ToggleBlock(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	var input dto.BlockUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	response, err := h.blockService.ToggleBlock(c.Request.Context(), userID.(string), input.TargetUserID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *BlockController) GetBlockedUsers(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	blockedUsers, err := h.blockService.GetBlockedUsers(c.Request.Context(), userID.(string))
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": blockedUsers})
}
