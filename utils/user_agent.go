package utils

import (
	"strings"
)

// FriendlyDeviceName trích xuất tên thiết bị thân thiện (trình duyệt + hệ điều hành)
// từ chuỗi User-Agent. Trả về chuỗi rỗng nếu không nhận diện được gì.
func FriendlyDeviceName(userAgent string) string {
	ua := strings.ToLower(userAgent)

	browser := ""
	switch {
	case strings.Contains(ua, "edg/"):
		browser = "Edge"
	case strings.Contains(ua, "coccoc/"):
		browser = "Cốc Cốc"
	case strings.Contains(ua, "opr/"):
		browser = "Opera"
	case strings.Contains(ua, "firefox/"):
		browser = "Firefox"
	case strings.Contains(ua, "chrome/"):
		browser = "Chrome"
	case strings.Contains(ua, "safari/"):
		browser = "Safari"
	}

	if strings.Contains(ua, "mobile") && browser == "Chrome" {
		browser = "Chrome Mobile"
	}

	os := ""
	switch {
	case strings.Contains(ua, "windows nt 6.1"):
		os = "Windows 7"
	case strings.Contains(ua, "windows nt 6.3"):
		os = "Windows 8.1"
	case strings.Contains(ua, "windows nt"):
		os = "Windows"
	case strings.Contains(ua, "android"):
		os = "Android"
	case strings.Contains(ua, "iphone"), strings.Contains(ua, "ipad"):
		os = "iOS"
	case strings.Contains(ua, "cros"):
		os = "Chrome OS"
	case strings.Contains(ua, "mac os x"), strings.Contains(ua, "macintosh"):
		os = "macOS"
	case strings.Contains(ua, "linux"):
		os = "Linux"
	}

	switch {
	case browser != "" && os != "":
		return browser + " trên " + os
	case browser != "":
		return browser
	case os != "":
		return os
	}
	return ""
}