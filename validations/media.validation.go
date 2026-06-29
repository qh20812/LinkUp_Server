package validations

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	ErrFileRequired         = errors.New("file là bắt buộc")
	ErrFileTypeNotAllowed   = errors.New("định dạng file không được hỗ trợ")
	ErrFileTooLarge         = errors.New("file vượt quá giới hạn kích thước")
	ErrInsufficientStorage  = errors.New("dung lượng lưu trữ không đủ")
	ErrStorageQuotaExceeded = errors.New("dung lượng lưu trữ đã đầy, vui lòng mua thêm dung lượng")
)

type MediaValidation struct {
	MaxImageSize      int64
	MaxVideoSize      int64
	AllowedImageTypes []string
	AllowedVideoTypes []string
}

func NewMediaValidation() *MediaValidation {
	return &MediaValidation{
		MaxImageSize:      10485760,  // 10MB
		MaxVideoSize:      104857600, // 100MB
		AllowedImageTypes: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
		AllowedVideoTypes: []string{".mp4", ".avi", ".mov", ".mkv", ".webm"},
	}
}

func (v *MediaValidation) ValidateFile(filename string, fileSize int64, contentType string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	isImage := contains(v.AllowedImageTypes, ext)
	isVideo := contains(v.AllowedVideoTypes, ext)

	if !isImage && !isVideo {
		return ErrFileTypeNotAllowed
	}

	if isImage && fileSize > v.MaxImageSize {
		return ErrFileTooLarge
	}
	if isVideo && fileSize > v.MaxVideoSize {
		return ErrFileTooLarge
	}

	return nil
}

func (v *MediaValidation) ValidateStorageQuota(availableBytes float64, requestedBytes int64) error {
	if float64(requestedBytes) > availableBytes {
		return ErrInsufficientStorage
	}
	return nil
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
