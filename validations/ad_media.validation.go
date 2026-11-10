package validations

import (
	"mime/multipart"
	"path/filepath"
	"strings"

	errorsapp "linkup/errors"
)

// Giới hạn media cho quảng cáo. Đây là nguồn số duy nhất phía server; web có
// bản tương ứng trong web/lib/ad-media.ts. Giữ khớp với trần Cloudinary Free
// (ảnh 10 MB, video 100 MB) để không upload xong mới bị từ chối.
const (
	MaxAdImageBytes       int64 = 10 << 20 // 10 MB
	MaxAdVideoBytes       int64 = 100 << 20
	MaxAdVideoSeconds           = 60
	MinCarouselFiles            = 2
	MaxCarouselFiles            = 6
	MaxCarouselTotalBytes int64 = 30 << 20
)

// ValidateAdMedia kiểm tra số lượng, loại và dung lượng media của một ad.
//
// `image`/`video` cho phép 0 hoặc 1 file (media vẫn là tuỳ chọn); `carousel`
// yêu cầu 2–6 ảnh, mỗi ảnh ≤ 10 MB và tổng ≤ 30 MB. Chạy trước khi service giữ
// slot để lỗi rẻ và không tạo giao dịch dở dang.
func ValidateAdMedia(files []*multipart.FileHeader, format string) error {
	format = strings.ToLower(strings.TrimSpace(format))

	switch format {
	case "carousel":
		if len(files) < MinCarouselFiles || len(files) > MaxCarouselFiles {
			return errorsapp.Newf(errorsapp.ErrCodeAdMediaCountInvalid, map[string]any{
				"min": MinCarouselFiles,
				"max": MaxCarouselFiles,
			})
		}
	case "image", "video":
		if len(files) > 1 {
			return errorsapp.Newf(errorsapp.ErrCodeAdMediaCountInvalid, map[string]any{
				"min": 0,
				"max": 1,
			})
		}
	default:
		return errorsapp.New(errorsapp.ErrCodeAdMediaTypeMismatch)
	}

	var total int64
	for _, file := range files {
		if file == nil {
			continue
		}

		kind := adMediaKind(file.Filename, file.Header.Get("Content-Type"))
		switch format {
		case "carousel", "image":
			if kind != "image" {
				return errorsapp.New(errorsapp.ErrCodeAdMediaTypeMismatch)
			}
			if file.Size > MaxAdImageBytes {
				return errorsapp.Newf(errorsapp.ErrCodeAdMediaTooLarge, map[string]any{
					"max": MaxAdImageBytes >> 20,
				})
			}
		case "video":
			if kind != "video" {
				return errorsapp.New(errorsapp.ErrCodeAdMediaTypeMismatch)
			}
			if file.Size > MaxAdVideoBytes {
				return errorsapp.Newf(errorsapp.ErrCodeAdMediaTooLarge, map[string]any{
					"max": MaxAdVideoBytes >> 20,
				})
			}
		}

		total += file.Size
	}

	if format == "carousel" && total > MaxCarouselTotalBytes {
		return errorsapp.Newf(errorsapp.ErrCodeAdCarouselTotalSize, map[string]any{
			"max": MaxCarouselTotalBytes >> 20,
		})
	}

	return nil
}

// adMediaKind suy ra loại media từ content-type, fallback theo phần mở rộng
// (một số trình duyệt/proxy gửi application/octet-stream).
func adMediaKind(filename, contentType string) string {
	ct := strings.ToLower(contentType)
	switch {
	case strings.HasPrefix(ct, "image/"):
		return "image"
	case strings.HasPrefix(ct, "video/"):
		return "video"
	}

	switch strings.ToLower(filepath.Ext(filename)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return "image"
	case ".mp4", ".mov", ".avi", ".mkv", ".webm":
		return "video"
	}
	return ""
}
