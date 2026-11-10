package validations

import "testing"

func TestValidateCreateReport_AllowsAdTarget(t *testing.T) {
	v := NewReportValidation()

	if err := v.ValidateCreateReport("ad", "ad-1", "spam", "quảng cáo sai sự thật", false); err != nil {
		t.Fatalf("target_type ad phải hợp lệ: %v", err)
	}
	if err := v.ValidateCreateReport("ad", "ad-1", "spam", "", true); err != nil {
		t.Fatalf("ad + violation rule không cần reason_detail: %v", err)
	}
}

func TestValidateCreateReport_RejectsUnknownTarget(t *testing.T) {
	v := NewReportValidation()

	if err := v.ValidateCreateReport("banner", "x", "spam", "lý do", false); err == nil {
		t.Fatalf("target_type lạ phải bị từ chối")
	}
}
