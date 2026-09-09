package services

import (
	"context"
	"net/mail"
	"strconv"
	"strings"

	"linkup/config"
	"linkup/dto"
	"linkup/errors"
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
		return nil, errors.Wrap(errors.ErrCodeAdminSettingsLoadFailed, err)
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
			return errors.Newf(errors.ErrCodeAdminInvalidSettingKey, map[string]any{"key": key})
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
			return errors.Wrap(errors.ErrCodeAdminSessionsInvalidFailed, err)
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
		return errors.Wrap(errors.ErrCodeAdminNoAccess, err)
	}
	if roleName != string(models.RoleSuperAdmin) {
		return errors.New(errors.ErrCodeAdminNotSuperadmin)
	}
	return nil
}

func validateSettingValue(key, value, expectedType string) error {
	switch expectedType {
	case "int":
		num, err := strconv.Atoi(value)
		if err != nil {
			return errors.Newf(errors.ErrCodeAdminInvalidInt, map[string]any{"key": key})
		}
		switch key {
		case "password_min_length":
			if num < 8 {
				return errors.Newf(errors.ErrCodeAdminValueTooLow, map[string]any{"key": key, "min": 8})
			}
			if num > 50 {
				return errors.Newf(errors.ErrCodeAdminValueTooHigh, map[string]any{"key": key, "max": 50})
			}
		case "max_login_attempts":
			if num < 1 {
				return errors.Newf(errors.ErrCodeAdminValueTooLow, map[string]any{"key": key, "min": 1})
			}
			if num > 10 {
				return errors.Newf(errors.ErrCodeAdminValueTooHigh, map[string]any{"key": key, "max": 10})
			}
		case "jwt_expiry_minutes":
			if num < 1 {
				return errors.Newf(errors.ErrCodeAdminValueTooLow, map[string]any{"key": key, "min": 1})
			}
			if num > 60 {
				return errors.Newf(errors.ErrCodeAdminValueTooHigh, map[string]any{"key": key, "max": 60})
			}
		case "refresh_token_expiry_days":
			if num < 1 {
				return errors.Newf(errors.ErrCodeAdminValueTooLow, map[string]any{"key": key, "min": 1})
			}
			if num > 30 {
				return errors.Newf(errors.ErrCodeAdminValueTooHigh, map[string]any{"key": key, "max": 30})
			}
		}
	case "bool":
		if value != "true" && value != "false" {
			return errors.Newf(errors.ErrCodeAdminInvalidBool, map[string]any{"key": key})
		}
	case "email":
		if strings.TrimSpace(value) == "" {
			return nil
		}
		if _, err := mail.ParseAddress(value); err != nil {
			return errors.Newf(errors.ErrCodeAdminInvalidEmail, map[string]any{"key": key})
		}
	}
	return nil
}
