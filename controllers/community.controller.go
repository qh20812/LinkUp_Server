package controllers

import (
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommunityController struct {
	communityService *services.CommunityService
	mediaService     services.MediaService
}

func NewCommunityController(communityService *services.CommunityService, mediaService services.MediaService) *CommunityController {
	return &CommunityController{communityService: communityService, mediaService: mediaService}
}

func (ctrl *CommunityController) CreateCommunity(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ"})
		return
	}

	avatarURI := ""
	file, err := c.FormFile("avatar")
	if err == nil && file != nil {
		media, err := ctrl.mediaService.UploadMedia(c.Request.Context(), userID.(string), file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tải ảnh đại diện thất bại"})
			return
		}
		avatarURI = media.FileURI
	}

	community, err := ctrl.communityService.CreateCommunity(c.Request.Context(), userID.(string), name, description, avatarURI)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Tạo cộng đồng thành công!",
		"community_id": community.ID,
	})
}

func (ctrl *CommunityController) SetCommunityBackground(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}

	communityID := c.Param("communityID")

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ"})
		return
	}

	file, err := c.FormFile("background")
	if err != nil || file == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng chọn ảnh background"})
		return
	}

	if err := ctrl.communityService.SetCommunityBackground(c.Request.Context(), userID.(string), communityID, file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật background cộng đồng thành công!",
	})
}
