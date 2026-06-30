package controllers

import (
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "người dùng chưa xác thực"})
		return
	}

	var input dto.BlockUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}

	response, err := h.blockService.ToggleBlock(c.Request.Context(), userID.(string), input.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *BlockController) GetBlockedUsers(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "người dùng chưa xác thực"})
		return
	}

	blockedUsers, err := h.blockService.GetBlockedUsers(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": blockedUsers})
}
