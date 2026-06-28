package validations

import (
	"errors"
	"strings"
)

// ValidateCreateComment kiểm tra dữ liệu đầu vào khi tạo bình luận
func ValidateCreateComment(content string) error {
	content = strings.TrimSpace(content)

	if content == "" {
		return errors.New("nội dung bình luận không được trống")
	}

	if len(content) > 1000 {
		return errors.New("nội dung bình luận không được vượt quá 1000 ký tự")
	}

	return nil
}
