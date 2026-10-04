package report_test

import (
	"strings"
	"testing"

	"linkup/validations"
)

func TestValidateCreateReport_DetailOptionalWithRule(t *testing.T) {
	v := validations.NewReportValidation()

	tests := []struct {
		name     string
		hasRule  bool
		detail   string
		wantErr  string
	}{
		{"rule chosen, empty detail", true, "", ""},
		{"rule chosen, whitespace detail", true, "   ", ""},
		{"rule chosen, with detail", true, "Nội dung xúc phạm", ""},
		{"no rule, empty detail", false, "", "reason_detail là bắt buộc"},
		{"no rule, whitespace detail", false, "   ", "reason_detail là bắt buộc"},
		{"no rule, with detail", false, "Lý do khác...", ""},
		{"detail too long", true, strings.Repeat("a", 1001), "reason_detail không được vượt quá 1000 ký tự"},
		{"detail too long without rule", false, strings.Repeat("a", 1001), "reason_detail không được vượt quá 1000 ký tự"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateCreateReport("post", "post-1", "violation", tt.detail, tt.hasRule)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestValidateUpdateReport_DetailOptionalWithRule(t *testing.T) {
	v := validations.NewReportValidation()

	if err := v.ValidateUpdateReport("violation", "", true); err != nil {
		t.Errorf("rule chosen with empty detail should pass, got: %v", err)
	}
	if err := v.ValidateUpdateReport("other", "", false); err == nil {
		t.Errorf("no rule with empty detail should fail, got nil")
	}
}
