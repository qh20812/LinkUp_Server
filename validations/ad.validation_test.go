package validations

import "testing"

func TestValidateUpdateAd_OnlyChecksProvidedFields(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		content   string
		targetURL string
		budget    float64
		wantErr   bool
	}{
		{name: "payload rỗng hợp lệ (PATCH bán phần)", wantErr: false},
		{name: "chỉ đổi ngân sách", budget: 500, wantErr: false},
		{name: "tiêu đề hợp lệ", title: "Khuyến mãi", wantErr: false},
		{name: "tiêu đề toàn khoảng trắng", title: "   ", wantErr: true},
		{name: "tiêu đề quá 100 ký tự", title: string(make([]rune, 101)), wantErr: true},
		{name: "nội dung toàn khoảng trắng", content: "  ", wantErr: true},
		{name: "link đích toàn khoảng trắng", targetURL: "   ", wantErr: true},
		{name: "ngân sách âm", budget: -1, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateUpdateAd(tc.title, tc.content, tc.targetURL, tc.budget)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateCreateAd_StillRequiresCoreFields(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		content   string
		targetURL string
		format    string
		budget    float64
		wantErr   bool
	}{
		{name: "đầy đủ", title: "A", content: "B", targetURL: "https://x.vn", format: "image", budget: 1},
		{name: "thiếu tiêu đề", content: "B", targetURL: "https://x.vn", format: "image", budget: 1, wantErr: true},
		{name: "thiếu nội dung", title: "A", targetURL: "https://x.vn", format: "image", budget: 1, wantErr: true},
		{name: "thiếu link", title: "A", content: "B", format: "image", budget: 1, wantErr: true},
		{name: "ngân sách 0", title: "A", content: "B", targetURL: "https://x.vn", format: "image", wantErr: true},
		{name: "format lạ", title: "A", content: "B", targetURL: "https://x.vn", format: "gif", budget: 1, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCreateAd(tc.title, tc.content, tc.targetURL, tc.format, tc.budget)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAdTargetURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "https hợp lệ", url: "https://linkup.vn/sale"},
		{name: "http hợp lệ", url: "http://linkup.vn"},
		{name: "tự trim khoảng trắng", url: "  https://linkup.vn  "},
		{name: "rỗng", url: "   ", wantErr: true},
		{name: "javascript scheme", url: "javascript:alert(1)", wantErr: true},
		{name: "data scheme", url: "data:text/html,<script>alert(1)</script>", wantErr: true},
		{name: "ftp scheme", url: "ftp://linkup.vn", wantErr: true},
		{name: "thiếu host", url: "https://", wantErr: true},
		{name: "không có scheme", url: "linkup.vn/sale", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAdTargetURL(tc.url)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

	tooLong := "https://linkup.vn/" + string(make([]rune, 2100))
	if err := validateAdTargetURL(tooLong); err == nil {
		t.Fatalf("URL quá dài phải bị từ chối")
	}
}
