package services

import (
	"context"
	"errors"
	"strings"

	"linkup/config"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
)

type UserSettingsService struct {
	settingsRepo *repository.UserSettingsRepository
	sessionRepo  *repository.UserSessionRepository
	authRepo     *repository.AuthRepository
	env          config.Env
}

func NewUserSettingsService(
	settingsRepo *repository.UserSettingsRepository,
	sessionRepo *repository.UserSessionRepository,
	authRepo *repository.AuthRepository,
	env config.Env,
) *UserSettingsService {
	return &UserSettingsService{
		settingsRepo: settingsRepo,
		sessionRepo:  sessionRepo,
		authRepo:     authRepo,
		env:          env,
	}
}

func (s *UserSettingsService) GetPrivacy(ctx context.Context, userID string) (dto.PrivacySettingsResponse, error) {
	setting, err := s.settingsRepo.GetByUserID(ctx, userID)
	if err != nil {
		return dto.PrivacySettingsResponse{}, err
	}
	if setting == nil {
		def := models.DefaultUserSetting(userID)
		setting = &def
	}
	return dto.PrivacySettingsResponse{
		DiscoverableInSearch:  setting.DiscoverableInSearch,
		AllowStrangerMessages: setting.AllowStrangerMessages,
	}, nil
}

func (s *UserSettingsService) UpdatePrivacy(ctx context.Context, userID string, input dto.UpdatePrivacyInput) (dto.PrivacySettingsResponse, error) {
	setting, err := s.settingsRepo.GetByUserID(ctx, userID)
	if err != nil {
		return dto.PrivacySettingsResponse{}, err
	}
	if setting == nil {
		def := models.DefaultUserSetting(userID)
		setting = &def
	}

	if input.DiscoverableInSearch != nil {
		setting.DiscoverableInSearch = *input.DiscoverableInSearch
	}
	if input.AllowStrangerMessages != nil {
		setting.AllowStrangerMessages = *input.AllowStrangerMessages
	}

	if err := s.settingsRepo.Upsert(ctx, setting); err != nil {
		return dto.PrivacySettingsResponse{}, err
	}

	return dto.PrivacySettingsResponse{
		DiscoverableInSearch:  setting.DiscoverableInSearch,
		AllowStrangerMessages: setting.AllowStrangerMessages,
	}, nil
}

func (s *UserSettingsService) GetStorage(ctx context.Context, userID string) (dto.StorageInfoResponse, error) {
	user, err := s.authRepo.FindByID(ctx, userID)
	if err != nil {
		return dto.StorageInfoResponse{}, err
	}
	return dto.StorageInfoResponse{
		QuotaBytes: user.StorageQuotaBytes,
		UsedBytes:  user.StorageUsedBytes,
		AvailBytes: user.AvailableStorageBytes(),
	}, nil
}

// Deactivate suspends the account (restorable by logging in again) after
// verifying the current password.
func (s *UserSettingsService) Deactivate(ctx context.Context, userID, password string) error {
	user, err := s.authRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := utils.ComparePassword(user.PasswordHash, password); err != nil {
		return errors.New("mật khẩu hiện tại không đúng")
	}

	if err := s.authRepo.Deactivate(ctx, userID); err != nil {
		return err
	}
	if s.sessionRepo != nil {
		_ = s.sessionRepo.RevokeAllByUserID(ctx, userID)
	}
	return nil
}

func (s *UserSettingsService) GetAppearance(ctx context.Context, userID string) (dto.AppearanceSettingsResponse, error) {
	setting, err := s.settingsRepo.GetByUserID(ctx, userID)
	if err != nil {
		return dto.AppearanceSettingsResponse{}, err
	}
	if setting == nil {
		def := models.DefaultUserSetting(userID)
		setting = &def
	}
	return dto.AppearanceSettingsResponse{
		Theme:    setting.Theme,
		Language: setting.Language,
	}, nil
}

func (s *UserSettingsService) UpdateAppearance(ctx context.Context, userID string, input dto.UpdateAppearanceInput) (dto.AppearanceSettingsResponse, error) {
	setting, err := s.settingsRepo.GetByUserID(ctx, userID)
	if err != nil {
		return dto.AppearanceSettingsResponse{}, err
	}
	if setting == nil {
		def := models.DefaultUserSetting(userID)
		setting = &def
	}

	if input.Theme != nil {
		theme := strings.ToLower(strings.TrimSpace(*input.Theme))
		if theme != "light" && theme != "dark" {
			return dto.AppearanceSettingsResponse{}, errors.New("chủ đề không hợp lệ")
		}
		setting.Theme = theme
	}
	if input.Language != nil {
		language := strings.ToLower(strings.TrimSpace(*input.Language))
		if language != "vi" && language != "en" {
			return dto.AppearanceSettingsResponse{}, errors.New("ngôn ngữ không hợp lệ")
		}
		setting.Language = language
	}

	if err := s.settingsRepo.Upsert(ctx, setting); err != nil {
		return dto.AppearanceSettingsResponse{}, err
	}

	return dto.AppearanceSettingsResponse{
		Theme:    setting.Theme,
		Language: setting.Language,
	}, nil
}

func (s *UserSettingsService) ListSessions(ctx context.Context, userID, currentSessionID string) ([]dto.SessionDTO, error) {
	if s.sessionRepo == nil {
		return []dto.SessionDTO{}, nil
	}
	sessions, err := s.sessionRepo.ListActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.SessionDTO, 0, len(sessions))
	for _, sess := range sessions {
		result = append(result, dto.SessionDTO{
			ID:           sess.ID,
			DeviceName:   sess.DeviceName,
			IPAddress:    sess.IPAddress,
			UserAgent:    sess.UserAgent,
			CreatedAt:    sess.CreatedAt,
			ExpiresAt:    sess.ExpiresAt,
			LastActiveAt: sess.LastActiveAt,
			IsCurrent:    sess.ID == currentSessionID,
		})
	}
	return result, nil
}

func (s *UserSettingsService) RevokeSession(ctx context.Context, userID, sessionID string) error {
	if s.sessionRepo == nil {
		return nil
	}
	revoked, err := s.sessionRepo.Revoke(ctx, sessionID, userID)
	if err != nil {
		return err
	}
	if !revoked {
		return errors.New("phiên đăng nhập không tồn tại")
	}
	return nil
}

func (s *UserSettingsService) RevokeOtherSessions(ctx context.Context, userID, currentSessionID string) error {
	if s.sessionRepo == nil {
		return nil
	}
	return s.sessionRepo.RevokeAllExcept(ctx, userID, currentSessionID)
}
