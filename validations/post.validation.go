package validations

import (
	"errors"
	"fmt"
	"linkup/models"
	"strings"
)

// ValidateCreatePost kiểm tra dữ liệu đầu vào khi tạo bài viết
func ValidateCreatePost(title, content, status string) error {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	status = strings.ToLower(strings.TrimSpace(status))

	// Kiểm tra tiêu đề (Từ 5 đến 150 ký tự)
	if len(title) < 5 || len(title) > 150 {
		return errors.New("tiêu đề bài viết phải từ 5 đến 150 ký tự")
	}

	// Kiểm tra nội dung (Tối đa 5000 ký tự)
	if content == "" {
		return errors.New("nội dung bài viết không được bỏ trống")
	}
	if len(content) > 5000 {
		return errors.New("nội dung bài viết không được vượt quá 5000 ký tự")
	}

	// Kiểm tra trạng thái bài viết hợp lệ
	validStatuses := map[models.PostStatus]bool{
		models.PostStatusActive:  true,
		models.PostStatusPublic:  true,
		models.PostStatusPrivate: true,
		models.PostStatusHidden:  true,
		models.PostStatusFriend:  true,
	}

	if status != "" && !validStatuses[models.PostStatus(status)] {
		return fmt.Errorf("trạng thái '%s' không hợp lệ. Chỉ chấp nhận: public, private, friend, hidden", status)
	}

	return nil
}
