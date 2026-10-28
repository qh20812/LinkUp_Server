package seeddata

import "testing"

// TestValidate đảm bảo toàn bộ seeddata không có tham chiếu gãy. Test này chạy
// không cần DB nên CI bắt được lỗi ngay khi ai đó sửa seeddata.
func TestValidate(t *testing.T) {
	if err := Validate(); err != nil {
		t.Fatalf("seeddata không hợp lệ: %v", err)
	}
}
