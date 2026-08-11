package controllers

import (
	"fmt"
	"linkup/dto"
	errorsapp "linkup/errors"
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
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)

	file, err := c.FormFile("file")
	if err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	media, err := ctrl.service.UploadMedia(c.Request.Context(), userID, file)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			status := errorsapp.StatusCode(appErr.Code)
			if appErr.Code == errorsapp.ErrCodeMediaInsufficientStorage {
				status = http.StatusPaymentRequired
			}
			errorsapp.Respond(c, status, appErr)
		} else {
			errorsapp.Respond(c, http.StatusInternalServerError, err)
		}
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

func (crtl *MediaController) DeleteMedia(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)

	mediaID := c.Param("id")
	if mediaID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	err := crtl.service.DeleteMedia(c.Request.Context(), userID, mediaID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusInternalServerError, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "Xóa media thành công",
		"media_id": mediaID,
	})
}

func (ctrl *MediaController) GetStorageStatus(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)

	quota, used, available, err := ctrl.service.GetUserStorageStatus(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
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
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", userIDVal)

	medias, err := ctrl.service.GetUserMedia(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": medias,
	})
}
