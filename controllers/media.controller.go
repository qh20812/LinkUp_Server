package controllers

import (
	"fmt"
	"linkup/dto"
	"linkup/services"
	"linkup/validations"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MediaController struct {
	service    services.MediaService
	validation *validations.MediaValidation
}

func NewMediaController(service services.MediaService) *MediaController {
	return &MediaController{
		service:    service,
		validation: validations.NewMediaValidation(),
	}
}

func (ctrl *MediaController) UploadMedia(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin đăng nhập"})
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	media, err := ctrl.service.UploadMedia(c.Request.Context(), userID, file)
	if err != nil {
		if err.Error() == validations.ErrFileTypeNotAllowed.Error() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Định dạng file không được hỗ trợ"})
			return
		}
		if err.Error() == validations.ErrFileTooLarge.Error() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("File quá lớn. Tối đa: Ảnh 50MB, Video 500MB"),
			})
			return
		}
		if err.Error() == validations.ErrInsufficientStorage.Error() {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error": "Dung lượng lưu trữ không đủ. Vui lòng mua thêm dung lượng",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	quota, used, available, _ := ctrl.service.GetUserStorageStatus(c.Request.Context(), userID)

	c.JSON(http.StatusCreated, gin.H{
		"data": dto.UploadMediaResponse{
			ID:               media.ID,
			FileURI:          media.FileURI,
			FileType:         media.FileType,
			FileSize:         media.FileSize,
			Status:           media.Status.String(),
			AvailableStorage: available,
		},
		"storage": gin.H{
			"quota_bytes":     quota,
			"used_bytes":      used,
			"available_bytes": available,
		},
	})
}

func (ctrl *MediaController) GetStorageStatus(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin đăng nhập"})
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)

	quota, used, available, err := ctrl.service.GetUserStorageStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	usagePercent := (used / quota) * 100

	c.JSON(http.StatusOK, gin.H{
		"data": dto.StorageStatusResponse{
			StorageQuotaBytes: quota,
			StorageUsedBytes:  used,
			AvailableBytes:    available,
			UsagePercentage:   usagePercent,
		},
	})
}

func (ctrl *MediaController) GetUserMedia(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin đăng nhập"})
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)

	medias, err := ctrl.service.GetUserMedia(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": medias,
	})
}
