package validations

import (
	"errors"
	"strings"
)

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

	targetURL = strings.TrimSpace(targetURL)
	if targetURL == "" {
		return errors.New("link đích không được để trống")
	}

	if budget <= 0 {
		return errors.New("ngân sách phải lớn hơn 0")
	}

	if format != "image" && format != "video" && format != "carousel" {
		return errors.New("định dạng phải là image, video hoặc carousel")
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
