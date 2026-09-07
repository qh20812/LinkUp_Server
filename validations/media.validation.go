package validations

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"path/filepath"
	"strings"

	errorsapp "linkup/errors"
	_ "golang.org/x/image/webp"
)

type DimensionConstraint struct {
	MinWidth    int
	MinHeight   int
	MaxWidth    int
	MaxHeight   int
	AspectRatio string
}

type MediaValidation struct {
	MaxImageSize      int64
	MaxVideoSize      int64
	MaxAudioSize      int64
	MaxAudioDuration  int
	AllowedImageTypes []string
	AllowedVideoTypes []string
	AllowedAudioTypes []string
}

func NewMediaValidation() *MediaValidation {
	return &MediaValidation{
		MaxImageSize:      10485760,  // 10MB
		MaxVideoSize:      104857600, // 100MB
		MaxAudioSize:      10485760,  // 10MB — tin nhắn thoại (voice notes)
		MaxAudioDuration:  300,       // 5 phút
		AllowedImageTypes: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
		AllowedVideoTypes: []string{".mp4", ".avi", ".mov", ".mkv", ".webm"},
		AllowedAudioTypes: []string{".mp3", ".wav", ".ogg", ".webm", ".m4a"},
	}
}

func (v *MediaValidation) ValidateFile(filename string, fileSize int64, contentType string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	isImage := contains(v.AllowedImageTypes, ext)
	isAudio := contains(v.AllowedAudioTypes, ext) || strings.HasPrefix(contentType, "audio/")
	isVideo := contains(v.AllowedVideoTypes, ext)

	if !isImage && !isAudio && !isVideo {
		return errorsapp.New(errorsapp.ErrCodeMediaFileTypeNotAllowed)
	}

	if isAudio {
		if fileSize > v.MaxAudioSize {
			return errorsapp.New(errorsapp.ErrCodeMediaFileTooLarge)
		}
		return nil
	}
	if isImage && fileSize > v.MaxImageSize {
		return errorsapp.New(errorsapp.ErrCodeMediaFileTooLarge)
	}
	if isVideo && fileSize > v.MaxVideoSize {
		return errorsapp.New(errorsapp.ErrCodeMediaFileTooLarge)
	}

	return nil
}

func (v *MediaValidation) ValidateAudioDuration(durationSeconds float64) error {
	if durationSeconds > float64(v.MaxAudioDuration) {
		return errorsapp.New(errorsapp.ErrCodeMediaDurationTooLong)
	}
	return nil
}

func (v *MediaValidation) ValidateStorageQuota(availableBytes float64, requestedBytes int64) error {
	if float64(requestedBytes) > availableBytes {
		return errorsapp.New(errorsapp.ErrCodeMediaInsufficientStorage)
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

func ValidateImageDimensions(r io.Reader, c DimensionConstraint) (width, height int, err error) {
	config, _, err := image.DecodeConfig(r)
	if err != nil {
		log.Printf("[media] image decode config failed (format may be unsupported): %v", err)
		return 0, 0, nil
	}

	w, h := config.Width, config.Height

	if w < c.MinWidth || h < c.MinHeight {
		return w, h, fmt.Errorf("%w: yêu cầu tối thiểu %dx%d, nhận được %dx%d", errorsapp.New(errorsapp.ErrCodeMediaImageTooSmall), c.MinWidth, c.MinHeight, w, h)
	}
	maxDim := math.Max(float64(w), float64(h))
	maxAllowed := math.Max(float64(c.MaxWidth), float64(c.MaxHeight))
	if maxDim > maxAllowed {
		return w, h, fmt.Errorf("%w: yêu cầu tối đa %dx%d, nhận được %dx%d", errorsapp.New(errorsapp.ErrCodeMediaImageTooLarge), c.MaxWidth, c.MaxHeight, w, h)
	}

	if c.AspectRatio == "1:1" {
		diff := math.Abs(float64(w - h))
		maxDim := math.Max(float64(w), float64(h))
		if maxDim > 0 && diff/maxDim > 0.05 {
			return w, h, errorsapp.New(errorsapp.ErrCodeMediaInvalidAspectRatio)
		}
	}

	return w, h, nil
}
