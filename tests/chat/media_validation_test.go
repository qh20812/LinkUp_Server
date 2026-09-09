package chat_test

import (
	"testing"

	errorsapp "linkup/errors"
	"linkup/validations"
)

// TestMediaValidation_AudioRules đảm bảo validate file hỗ trợ tin nhắn thoại:
// chấp nhận định dạng audio (webm/ogg/mpeg/wav), giới hạn 10MB, và chặn file
// không phải media hợp lệ.
func TestMediaValidation_AudioRules(t *testing.T) {
	v := validations.NewMediaValidation()

	cases := []struct {
		name        string
		filename    string
		size        int64
		contentType string
		wantErr     bool
	}{
		{"webm audio voice note", "voice.webm", 1_000_000, "audio/webm", false},
		{"webm video (not voice)", "clip.webm", 1_000_000, "video/webm", false},
		{"ogg audio", "voice.ogg", 2_000_000, "audio/ogg", false},
		{"mp3 audio", "voice.mp3", 512_000, "audio/mpeg", false},
		{"audio identified by MIME only", "file.bin", 1_000_000, "audio/wav", false},
		{"audio beyond size limit", "big.webm", 11_000_000, "audio/webm", true},
		{"unsupported ext", "voice.xyz", 1000, "", true},
		{"video file too large", "movie.mp4", 105_000_000, "video/mp4", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateFile(tc.filename, tc.size, tc.contentType)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestMediaValidation_AudioDuration đảm bảo thời lượng voice note ≤ 5 phút.
func TestMediaValidation_AudioDuration(t *testing.T) {
	v := validations.NewMediaValidation()

	if err := v.ValidateAudioDuration(0); err != nil {
		t.Fatalf("0s should pass, got %v", err)
	}
	if err := v.ValidateAudioDuration(299.8); err != nil {
		t.Fatalf("299.8s should pass, got %v", err)
	}
	if err := v.ValidateAudioDuration(301); err == nil {
		t.Fatalf("301s should be rejected")
	}
	if err := v.ValidateAudioDuration(600); err == nil {
		t.Fatalf("600s should be rejected")
	}
	if err := v.ValidateAudioDuration(301); err == nil {
		t.Fatalf("expected error, got nil")
	} else if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.GetCode() != errorsapp.ErrCodeMediaDurationTooLong {
		t.Fatalf("expected DURATION_TOO_LONG error, got %v", err)
	}
}