package community_test

import (
	"testing"
	"time"

	"linkup/models"
	"linkup/validations"
)

func TestValidateCreateCommunity(t *testing.T) {
	v := validations.NewCommunityValidation()

	tests := []struct {
		name         string
		community    string
		description  string
		avatarURI    string
		backgroundURI string
		privacy      models.CommunityPrivacy
		wantErr      string
	}{
		{"valid minimal", "Go Developers", "", "", "", "", ""},
		{"valid full", "Cloud Architects", "A community for cloud architecture", "https://example.com/avatar.png", "https://example.com/bg.png", models.PrivacyPublic, ""},
		{"valid invitation only", "Secret Group", "", "", "", models.PrivacyInvitationOnly, ""},
		{"empty name", "", "Some description", "", "", "", "Tên cộng đồng không được để trống"},
		{"name too short", "AB", "Some description", "", "", "", "Tên cộng đồng phải có ít nhất 3 ký tự"},
		{"name too long", string(make([]rune, 101)), "Some description", "", "", "", "Tên cộng đồng không được vượt quá 100 ký tự"},
		{"description too long", "Valid Name", string(make([]rune, 501)), "", "", "", "Mô tả cộng đồng không được vượt quá 500 ký tự"},
		{"invalid avatar", "Valid Name", "", "not-a-url", "", "", "Avatar URI không hợp lệ"},
		{"avatar without scheme", "Valid Name", "", "example.com/avatar.png", "", "", "Avatar URI không hợp lệ"},
		{"invalid background", "Valid Name", "", "", "not-a-url", "", "Background URI không hợp lệ"},
		{"invalid privacy typo", "Valid Name", "", "", "", "publik", "Quyền riêng tư không hợp lệ. Chỉ chấp nhận: public, invitation_only"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateCreateCommunity(tt.community, tt.description, tt.avatarURI, tt.backgroundURI, tt.privacy)
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
		{"whitespace only", "   ", "Tên cộng đồng không được để trống"},
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
		{"501 chars", string(make([]rune, 501)), "Mô tả cộng đồng không được vượt quá 500 ký tự"},
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

func TestValidateInviteCode(t *testing.T) {
	v := validations.NewCommunityValidation()
	now := time.Now().UTC()

	tests := []struct {
		name    string
		code    *models.CommunityInviteCode
		wantErr string
	}{
		{"nil code", nil, "Mã mời không tồn tại"},
		{"inactive", &models.CommunityInviteCode{IsActive: false}, "Mã mời đã bị vô hiệu hóa"},
		{"expired", &models.CommunityInviteCode{IsActive: true, ExpiresAt: ptrTime(now.Add(-1 * time.Hour))}, "Mã mời đã hết hạn"},
		{"max uses reached", &models.CommunityInviteCode{IsActive: true, MaxUses: 5, UsedCount: 5}, "Mã mời đã đạt số lần sử dụng tối đa"},
		{"under max uses", &models.CommunityInviteCode{IsActive: true, MaxUses: 5, UsedCount: 3}, ""},
		{"unlimited uses", &models.CommunityInviteCode{IsActive: true, MaxUses: 0, UsedCount: 100}, ""},
		{"not expired yet", &models.CommunityInviteCode{IsActive: true, ExpiresAt: ptrTime(now.Add(1 * time.Hour))}, ""},
		{"no expiry", &models.CommunityInviteCode{IsActive: true}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateInviteCode(tt.code)
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

func ptrTime(t time.Time) *time.Time {
	return &t
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
		{"no scheme", "ftp://example.com/avatar.png", "Avatar URI không hợp lệ"},
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
