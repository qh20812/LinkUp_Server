package community_test

import (
	"testing"

	"linkup/validations"
)

func TestValidateCreateCommunity(t *testing.T) {
	v := validations.NewCommunityValidation()

	tests := []struct {
		name        string
		community   string
		description string
		avatarURI   string
		wantErr     string
	}{
		{"valid minimal", "Go Developers", "", "", ""},
		{"valid full", "Cloud Architects", "A community for cloud architecture", "https://example.com/avatar.png", ""},
		{"empty name", "", "Some description", "", "tên cộng đồng không được để trống"},
		{"name too short", "AB", "Some description", "", "tên cộng đồng phải có ít nhất 3 ký tự"},
		{"name too long", string(make([]rune, 101)), "Some description", "", "tên cộng đồng không được vượt quá 100 ký tự"},
		{"description too long", "Valid Name", string(make([]rune, 501)), "", "mô tả cộng đồng không được vượt quá 500 ký tự"},
		{"invalid avatar", "Valid Name", "", "not-a-url", "avatar URI không hợp lệ"},
		{"avatar without scheme", "Valid Name", "", "example.com/avatar.png", "avatar URI không hợp lệ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateCreateCommunity(tt.community, tt.description, tt.avatarURI)
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

func TestValidateName_EdgeCases(t *testing.T) {
	v := validations.NewCommunityValidation()

	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{"exactly 3 chars", "ABC", ""},
		{"exactly 100 chars", string(make([]rune, 100)), ""},
		{"whitespace only", "   ", "tên cộng đồng không được để trống"},
		{"unicode name", "Cộng Đồng Việt Nam", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateName(tt.input)
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

func TestValidateDescription_Boundaries(t *testing.T) {
	v := validations.NewCommunityValidation()

	tests := []struct {
		name        string
		description string
		wantErr     string
	}{
		{"empty", "", ""},
		{"exactly 500 chars", string(make([]rune, 500)), ""},
		{"501 chars", string(make([]rune, 501)), "mô tả cộng đồng không được vượt quá 500 ký tự"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateDescription(tt.description)
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

func TestValidateAvatarURI_ValidURLs(t *testing.T) {
	v := validations.NewCommunityValidation()

	tests := []struct {
		name      string
		avatarURI string
		wantErr   string
	}{
		{"empty", "", ""},
		{"https", "https://res.cloudinary.com/demo/image/upload/v1/avatar.jpg", ""},
		{"http", "http://example.com/avatar.png", ""},
		{"no scheme", "ftp://example.com/avatar.png", "avatar URI không hợp lệ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateAvatarURI(tt.avatarURI)
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
