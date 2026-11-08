package validations

import (
	"mime/multipart"
	"net/textproto"
	"testing"

	errorsapp "linkup/errors"
)

func adMediaFile(name string, size int64, contentType string) *multipart.FileHeader {
	header := make(textproto.MIMEHeader)
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	return &multipart.FileHeader{Filename: name, Size: size, Header: header}
}

func assertAdMediaErr(t *testing.T, err error, wantCode string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %s, got nil", wantCode)
	}
	appErr, ok := errorsapp.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != wantCode {
		t.Fatalf("code = %s, want %s (message: %s)", appErr.Code, wantCode, appErr.Message)
	}
}

func TestValidateAdMedia(t *testing.T) {
	const mb = int64(1 << 20)

	tests := []struct {
		name    string
		files   []*multipart.FileHeader
		format  string
		wantErr string
	}{
		{
			name:   "ảnh đơn 5MB hợp lệ",
			files:  []*multipart.FileHeader{adMediaFile("a.jpg", 5*mb, "image/jpeg")},
			format: "image",
		},
		{
			name:   "không có media vẫn hợp lệ (media tuỳ chọn)",
			files:  nil,
			format: "image",
		},
		{
			name:    "ảnh vượt 10MB bị chặn",
			files:   []*multipart.FileHeader{adMediaFile("a.jpg", 10*mb+1, "image/jpeg")},
			format:  "image",
			wantErr: errorsapp.ErrCodeAdMediaTooLarge,
		},
		{
			name:   "ảnh đúng 10MB vẫn hợp lệ",
			files:  []*multipart.FileHeader{adMediaFile("a.png", 10*mb, "image/png")},
			format: "image",
		},
		{
			name:    "format ảnh nhưng file video",
			files:   []*multipart.FileHeader{adMediaFile("v.mp4", mb, "video/mp4")},
			format:  "image",
			wantErr: errorsapp.ErrCodeAdMediaTypeMismatch,
		},
		{
			name:    "ảnh đơn gửi 2 file",
			files:   []*multipart.FileHeader{adMediaFile("a.jpg", mb, "image/jpeg"), adMediaFile("b.jpg", mb, "image/jpeg")},
			format:  "image",
			wantErr: errorsapp.ErrCodeAdMediaCountInvalid,
		},
		{
			name:   "video 50MB hợp lệ",
			files:  []*multipart.FileHeader{adMediaFile("v.mp4", 50*mb, "video/mp4")},
			format: "video",
		},
		{
			name:    "video vượt 100MB bị chặn",
			files:   []*multipart.FileHeader{adMediaFile("v.mp4", 100*mb+1, "video/mp4")},
			format:  "video",
			wantErr: errorsapp.ErrCodeAdMediaTooLarge,
		},
		{
			name:    "format video nhưng file ảnh",
			files:   []*multipart.FileHeader{adMediaFile("a.jpg", mb, "image/jpeg")},
			format:  "video",
			wantErr: errorsapp.ErrCodeAdMediaTypeMismatch,
		},
		{
			name:    "carousel 1 ảnh là không đủ",
			files:   []*multipart.FileHeader{adMediaFile("a.jpg", mb, "image/jpeg")},
			format:  "carousel",
			wantErr: errorsapp.ErrCodeAdMediaCountInvalid,
		},
		{
			name: "carousel 7 ảnh là quá nhiều",
			files: []*multipart.FileHeader{
				adMediaFile("1.jpg", mb, "image/jpeg"),
				adMediaFile("2.jpg", mb, "image/jpeg"),
				adMediaFile("3.jpg", mb, "image/jpeg"),
				adMediaFile("4.jpg", mb, "image/jpeg"),
				adMediaFile("5.jpg", mb, "image/jpeg"),
				adMediaFile("6.jpg", mb, "image/jpeg"),
				adMediaFile("7.jpg", mb, "image/jpeg"),
			},
			format:  "carousel",
			wantErr: errorsapp.ErrCodeAdMediaCountInvalid,
		},
		{
			name: "carousel 6 ảnh tổng đúng 30MB hợp lệ",
			files: []*multipart.FileHeader{
				adMediaFile("1.jpg", 5*mb, "image/jpeg"),
				adMediaFile("2.jpg", 5*mb, "image/jpeg"),
				adMediaFile("3.jpg", 5*mb, "image/jpeg"),
				adMediaFile("4.jpg", 5*mb, "image/jpeg"),
				adMediaFile("5.jpg", 5*mb, "image/jpeg"),
				adMediaFile("6.jpg", 5*mb, "image/jpeg"),
			},
			format: "carousel",
		},
		{
			name: "carousel 6 ảnh tổng vượt 30MB bị chặn",
			files: []*multipart.FileHeader{
				adMediaFile("1.jpg", 6*mb, "image/jpeg"),
				adMediaFile("2.jpg", 6*mb, "image/jpeg"),
				adMediaFile("3.jpg", 6*mb, "image/jpeg"),
				adMediaFile("4.jpg", 6*mb, "image/jpeg"),
				adMediaFile("5.jpg", 6*mb, "image/jpeg"),
				adMediaFile("6.jpg", 6*mb, "image/jpeg"),
			},
			format:  "carousel",
			wantErr: errorsapp.ErrCodeAdCarouselTotalSize,
		},
		{
			name:    "carousel có video",
			files:   []*multipart.FileHeader{adMediaFile("1.jpg", mb, "image/jpeg"), adMediaFile("2.mp4", mb, "video/mp4")},
			format:  "carousel",
			wantErr: errorsapp.ErrCodeAdMediaTypeMismatch,
		},
		{
			name:   "content-type rỗng thì fallback theo phần mở rộng",
			files:  []*multipart.FileHeader{adMediaFile("a.JPG", mb, "")},
			format: "image",
		},
		{
			name:    "định dạng lạ bị chặn",
			files:   []*multipart.FileHeader{adMediaFile("a.jpg", mb, "image/jpeg")},
			format:  "banner",
			wantErr: errorsapp.ErrCodeAdMediaTypeMismatch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAdMedia(tc.files, tc.format)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}
				return
			}
			assertAdMediaErr(t, err, tc.wantErr)
		})
	}
}
