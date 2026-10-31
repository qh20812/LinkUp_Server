package services

import (
	"strings"
	"testing"

	"linkup/models"
)

func strPtr(s string) *string { return &s }

func TestMessagePushBody(t *testing.T) {
	cases := []struct {
		name          string
		plaintext     string
		e2eVersion    int
		emojiID       *string
		mediaID       *string
		sharedPostID  *string
		forwardedFrom *string
		mediaFileType string
		want          string
	}{
		{"legacy text", "xin chào", 0, nil, nil, nil, nil, "", "xin chào"},
		{"e2e text", "ciphertext...", 1, nil, nil, nil, nil, "", ""},
		{"e2e image", "", 1, nil, strPtr("m1"), nil, nil, "image/jpeg", "đã gửi một ảnh"},
		{"e2e gif", "", 1, nil, strPtr("m1"), nil, nil, "image/gif", "đã gửi một ảnh"},
		{"e2e video", "", 1, nil, strPtr("m1"), nil, nil, "video/mp4", "đã gửi một video"},
		{"e2e file", "", 1, nil, strPtr("m1"), nil, nil, "application/pdf", "đã gửi một tệp"},
		{"e2e emoji", "", 1, strPtr("e1"), nil, nil, nil, "", "đã gửi một nhãn dán"},
		{"e2e shared post", "", 1, nil, nil, strPtr("p1"), nil, "", "đã gửi một bài viết"},
		{"e2e forward", "", 1, nil, nil, nil, strPtr("orig"), "", "đã chuyển tiếp một tin nhắn"},
		{"legacy media không caption", "", 0, nil, strPtr("m1"), nil, nil, "image/png", "đã gửi một ảnh"},
		{"legacy media có caption", "caption", 0, nil, strPtr("m1"), nil, nil, "image/png", "caption"},
		{"không có gì", "", 1, nil, nil, nil, nil, "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := messagePushBody(tc.plaintext, tc.e2eVersion, tc.emojiID, tc.mediaID, tc.sharedPostID, tc.forwardedFrom, tc.mediaFileType)
			if got != tc.want {
				t.Errorf("messagePushBody() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTruncateRunes(t *testing.T) {
	if got := truncateRunes("abc", 10); got != "abc" {
		t.Errorf("truncateRunes ngắn = %q, want %q", got, "abc")
	}

	long := strings.Repeat("á", 200)
	got := truncateRunes(long, 160)
	if n := len([]rune(got)); n != 161 {
		t.Errorf("độ dài sau truncate = %d runes, want 161 (160 + …)", n)
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("thiếu dấu … ở cuối: %q", got[len(got)-3:])
	}
}

func TestApplyPushOptions(t *testing.T) {
	if cfg := applyPushOptions(nil); cfg.body != "" {
		t.Errorf("cfg.body = %q, want empty", cfg.body)
	}
	cfg := applyPushOptions([]PushOption{WithPushBody("nội dung chi tiết")})
	if cfg.body != "nội dung chi tiết" {
		t.Errorf("cfg.body = %q, want %q", cfg.body, "nội dung chi tiết")
	}
}

func TestPushThreadID(t *testing.T) {
	chatID := "chat-1"
	if got := pushThreadID(models.NotificationTypeMessage, &chatID); got != "chat-1" {
		t.Errorf("pushThreadID(message) = %q, want chat-1", got)
	}
	if got := pushThreadID(models.NotificationTypeLike, &chatID); got != "" {
		t.Errorf("pushThreadID(like) = %q, want empty", got)
	}
	if got := pushThreadID(models.NotificationTypeMessage, nil); got != "" {
		t.Errorf("pushThreadID(nil) = %q, want empty", got)
	}
}
