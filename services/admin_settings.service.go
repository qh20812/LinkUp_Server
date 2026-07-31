package services

import (
	"context"
	"fmt"
	"net/mail"
	"strconv"
	"strings"

	"linkup/config"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
)

var allowedSettings = map[string]string{
	"site_name":            "string",
	"site_description":     "string",
	"contact_email":        "email",
	"maintenance_mode":     "bool",
	"allow_registration":   "bool",
	"require_email_verify": "bool",
	"password_min_length":       "int",
	"max_login_attempts":        "int",
	"jwt_expiry_minutes":        "int",
	"refresh_token_expiry_days": "int",
	"default_user_role":         "string",
}

type AdminSettingsService struct {
	repo     *repository.AdminSettingsRepository
	authRepo *repository.AuthRepository
	env      config.Env
}

func NewAdminSettingsService(repo *repository.AdminSettingsRepository, env config.Env) *AdminSettingsService {
	return &AdminSettingsService{repo: repo, env: env}
}

func (s *AdminSettingsService) SetAuthRepository(repo *repository.AuthRepository) {
	s.authRepo = repo
}

func (s *AdminSettingsService) GetSettings(ctx context.Context, adminID string) (*dto.AdminSettingsResponse, error) {
	if err := s.ensureSuperAdmin(ctx, adminID); err != nil {
		return nil, err
	}

	configs, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("không thể tải cài đặt: %w", err)
	}

	settings := make(map[string]string, len(configs))
	for _, c := range configs {
		settings[c.Key] = c.Value
	}

	return &dto.AdminSettingsResponse{Settings: settings}, nil
}

func (s *AdminSettingsService) UpdateSettings(ctx context.Context, adminID string, input *dto.AdminSettingsUpdateInput) error {
	if err := s.ensureSuperAdmin(ctx, adminID); err != nil {
		return err
	}

	for key, value := range input.Settings {
		expectedType, ok := allowedSettings[key]
		if !ok {
			return fmt.Errorf("cài đặt '%s' không hợp lệ", key)
		}
		if err := validateSettingValue(key, value, expectedType); err != nil {
			return err
		}
	}

	if err := s.repo.UpsertBatch(ctx, input.Settings); err != nil {
		return err
	}

	if val, ok := input.Settings["maintenance_mode"]; ok && val == "true" && s.authRepo != nil {
		if err := s.authRepo.IncrementAllTokenVersions(ctx); err != nil {
			return fmt.Errorf("vô hiệu hóa phiên đăng nhập thất bại: %w", err)
		}
	}

	return nil
}

func (s *AdminSettingsService) ensureSuperAdmin(ctx context.Context, adminID string) error {
	var roleName string
	err := s.repo.DB().WithContext(ctx).Table("roles").
		Select("roles.name").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", adminID).
		Where("user_roles.scope_id IS NULL").
		Scan(&roleName).Error
	if err != nil || roleName == "" {
		return fmt.Errorf("không thể xác thực quyền: %w", err)
	}
	if roleName != string(models.RoleSuperAdmin) {
		return fmt.Errorf("Bạn không có quyền truy cập trang này")
	}
	return nil
}

func validateSettingValue(key, value, expectedType string) error {
	switch expectedType {
	case "int":
		num, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("'%s' phải là số nguyên", key)
		}
		switch key {
		case "password_min_length":
			if num < 8 {
				return fmt.Errorf("'password_min_length' phải >= 8")
			}
			if num > 50 {
				return fmt.Errorf("'password_min_length' không được vượt quá 50")
			}
		case "max_login_attempts":
			if num < 1 {
				return fmt.Errorf("'max_login_attempts' phải >= 1")
			}
			if num > 10 {
				return fmt.Errorf("'max_login_attempts' không được vượt quá 10")
			}
		case "jwt_expiry_minutes":
			if num < 1 {
				return fmt.Errorf("'jwt_expiry_minutes' phải >= 1")
			}
			if num > 60 {
				return fmt.Errorf("'jwt_expiry_minutes' không được vượt quá 60")
			}
		case "refresh_token_expiry_days":
			if num < 1 {
				return fmt.Errorf("'refresh_token_expiry_days' phải >= 1")
			}
			if num > 30 {
				return fmt.Errorf("'refresh_token_expiry_days' không được vượt quá 30")
			}
		}
	case "bool":
		if value != "true" && value != "false" {
			return fmt.Errorf("'%s' phải là 'true' hoặc 'false'", key)
		}
	case "email":
		if strings.TrimSpace(value) == "" {
			return nil
		}
		if _, err := mail.ParseAddress(value); err != nil {
			return fmt.Errorf("'%s' không phải email hợp lệ", key)
		}
	}
	return nil
}