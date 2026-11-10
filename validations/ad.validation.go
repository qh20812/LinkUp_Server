package validations

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const maxAdTargetURLLength = 2048

// validateAdTargetURL chỉ nhận link http/https tuyệt đối; chặn
// javascript:, data:, ftp: … để link quảng cáo không chạy script.
func validateAdTargetURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return errors.New("link đích không được để trống")
	}
	if len(raw) > maxAdTargetURLLength {
		return fmt.Errorf("link đích không được vượt quá %d ký tự", maxAdTargetURLLength)
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return errors.New("link đích không hợp lệ")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("link đích phải bắt đầu bằng http:// hoặc https://")
	}
	return nil
}

func ValidateCreateAd(title, content, targetURL, format string, budget float64) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.New("tiêu đề không được để trống")
	}
	if len(title) > 100 {
		return errors.New("tiêu đề không được vượt quá 100 ký tự")
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("nội dung không được để trống")
	}
	if len(content) > 2000 {
		return errors.New("nội dung không được vượt quá 2000 ký tự")
	}

	if err := validateAdTargetURL(targetURL); err != nil {
		return err
	}

	if budget <= 0 {
		return errors.New("ngân sách phải lớn hơn 0")
	}

	if format != "image" && format != "video" && format != "carousel" {
		return errors.New("định dạng phải là image, video hoặc carousel")
	}

	return nil
}

// ValidateUpdateAd chỉ kiểm tra những trường client thực sự gửi lên (PATCH bán
// phần), để trường vắng mặt không bị coi là "để trống".
func ValidateUpdateAd(title, content, targetURL string, budget float64) error {
	if title != "" {
		trimmed := strings.TrimSpace(title)
		if trimmed == "" {
			return errors.New("tiêu đề không được để trống")
		}
		if len(trimmed) > 100 {
			return errors.New("tiêu đề không được vượt quá 100 ký tự")
		}
	}

	if content != "" {
		trimmed := strings.TrimSpace(content)
		if trimmed == "" {
			return errors.New("nội dung không được để trống")
		}
		if len(trimmed) > 2000 {
			return errors.New("nội dung không được vượt quá 2000 ký tự")
		}
	}

	if targetURL != "" {
		if strings.TrimSpace(targetURL) == "" {
			return errors.New("link đích không được để trống")
		}
		if err := validateAdTargetURL(targetURL); err != nil {
			return err
		}
	}

	if budget < 0 {
		return errors.New("ngân sách phải lớn hơn 0")
	}

	return nil
}

func ValidateTargeting(gender string, ageMin, ageMax int) error {
	if gender != "" && gender != "all" && gender != "male" && gender != "female" {
		return errors.New("target_gender phải là all, male hoặc female")
	}
	if ageMin < 0 || ageMin > 100 {
		return errors.New("target_age_min phải từ 0 đến 100")
	}
	if ageMax < 0 || ageMax > 100 {
		return errors.New("target_age_max phải từ 0 đến 100")
	}
	if ageMin > ageMax && ageMax != 0 {
		return errors.New("target_age_min phải nhỏ hơn hoặc bằng target_age_max")
	}
	return nil
}
