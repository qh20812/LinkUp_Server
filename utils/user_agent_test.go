package utils

import (
	"testing"
)

func TestFriendlyDeviceName(t *testing.T) {
	tests := []struct {
		name string
		ua   string
		want string
	}{
		{
			name: "chrome windows",
			ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
			want: "Chrome trên Windows",
		},
		{
			name: "edge contains chrome",
			ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
			want: "Edge trên Windows",
		},
		{
			name: "coccoc on android",
			ua:   "Mozilla/5.0 (Linux; Android 14; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36 CocCoc/8.5",
			want: "Cốc Cốc trên Android",
		},
		{
			name: "firefox android",
			ua:   "Mozilla/5.0 (Android 14; Mobile; rv:130.0) Gecko/130.0 Firefox/130.0",
			want: "Firefox trên Android",
		},
		{
			name: "safari ios",
			ua:   "Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1",
			want: "Safari trên iOS",
		},
		{
			name: "chrome mobile",
			ua:   "Mozilla/5.0 (Linux; Android 13; SM-A236B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36",
			want: "Chrome Mobile trên Android",
		},
		{
			name: "windows 8.1",
			ua:   "Mozilla/5.0 (Windows NT 6.3; Trident/7.0; rv:11.0) like Gecko",
			want: "Windows 8.1",
		},
		{
			name: "empty",
			ua:   "",
			want: "",
		},
		{
			name: "unknown",
			ua:   "CustomApp/1.0",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FriendlyDeviceName(tt.ua); got != tt.want {
				t.Errorf("FriendlyDeviceName(%q) = %q, want %q", tt.ua, got, tt.want)
			}
		})
	}
}