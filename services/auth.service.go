package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"linkup/config"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
)

type AuthService struct {
	authRepo    *repository.AuthRepository
	profileRepo *repository.ProfileRepository
	env         config.Env
}

func NewAuthService(authRepo *repository.AuthRepository, profileRepo *repository.ProfileRepository, env config.Env) *AuthService {
	return &AuthService{
		authRepo:    authRepo,
		profileRepo: profileRepo,
		env:         env,
	}
}

func (s *AuthService) Register(ctx context.Context, input dto.RegisterInput) (dto.AuthResponse, error) {
	email := normalizeEmail(input.Email)

	if _, err := s.authRepo.FindByEmail(ctx, email); err == nil {
		return dto.AuthResponse{}, errors.New("email đã tồn tại")
	} else if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return dto.AuthResponse{}, err
	}

	username, err := utils.GenerateUsername(email, func(u string) (bool, error) {
		return s.authRepo.IsUsernameTaken(ctx, u)
	})
	if err != nil {
		return dto.AuthResponse{}, err
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	createdUser, err := s.authRepo.Create(ctx, &models.User{
		ID:           utils.GenerateUUID(),
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		Status:       models.UserStatusActive,
		StorageQuotaBytes: models.DefaultStorageQuotaBytes,
		StorageUsedBytes: 0,
	})
	if err != nil {
		return dto.AuthResponse{}, err
	}

	if err := s.authRepo.SavePasswordHistory(ctx, createdUser.ID, hashedPassword); err != nil {
		return dto.AuthResponse{}, err
	}

	if _, err := s.profileRepo.Create(ctx, &models.Profile{
		ID:          utils.GenerateUUID(),
		UserID:      createdUser.ID,
		DisplayName: input.DisplayName,
	}); err != nil {
		return dto.AuthResponse{}, err
	}

	if err := s.authRepo.AssignUserRole(ctx, createdUser.ID, models.RoleUser, nil, nil); err != nil {
		return dto.AuthResponse{}, err
	}

	accessToken, refreshToken, err := s.generateTokens(createdUser)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	return buildAuthResponse(*createdUser, accessToken, refreshToken, s.accessTTL(), s.refreshTTL()), nil
}

func (s *AuthService) Login(ctx context.Context, input dto.LoginInput) (dto.AuthResponse, error) {
	email := normalizeEmail(input.Email)
	user, err := s.authRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.AuthResponse{}, errors.New("email hoặc mật khẩu không hợp lệ")
		}
		return dto.AuthResponse{}, err
	}

	if !user.IsActive() {
		return dto.AuthResponse{}, fmt.Errorf("tài khoản chưa được kích hoạt")
	}

	if err := utils.ComparePassword(user.PasswordHash, input.Password); err != nil {
		return dto.AuthResponse{}, errors.New("email hoặc mật khẩu không hợp lệ")
	}

	accessToken, refreshToken, err := s.generateTokens(user)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	return buildAuthResponse(*user, accessToken, refreshToken, s.accessTTL(), s.refreshTTL()), nil
}

func (s *AuthService) generateTokens(user *models.User) (string, string, error) {
	return utils.GenerateTokenPair(s.env.JWTSecret, user.ID, user.Email, s.accessTTL(), s.refreshTTL())
}

func (s *AuthService) accessTTL() time.Duration {
	if s.env.JWTExpiresIn <= 0 {
		return 15 * time.Minute
	}
	return time.Duration(s.env.JWTExpiresIn) * time.Minute
}

func (s *AuthService) refreshTTL() time.Duration {
	return 7 * 24 * time.Hour
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func buildAuthResponse(user models.User, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration) dto.AuthResponse {
	return dto.AuthResponse{
		User: dto.AuthUserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Status:    string(user.Status),
			CreatedAt: user.CreatedAt,
		},
		Tokens: dto.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(accessTTL.Seconds()),
			RefreshTTLIn: int64(refreshTTL.Seconds()),
		},
		Storage: dto.StorageInfo{
			QuotaBytes: user.StorageQuotaBytes,
			UsedBytes: user.StorageUsedBytes,
			AvailBytes: user.AvailableStorageBytes(),
		},
	}
}

func (s *AuthService) ChangePassword(ctx context.Context, userID string, input dto.ChangePasswordInput) error {
	if err := validations.NewAuthValidation().ValidatePassword(input.NewPassword); err != nil {
		return err
	}

	user, err := s.authRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return errors.New("không tìm thấy người dùng")
		}
		return err
	}

	if err := utils.ComparePassword(user.PasswordHash, input.OldPassword); err != nil {
		return errors.New("mật khẩu hiện tại không đúng")
	}

	if input.OldPassword == input.NewPassword {
		return validations.ErrPasswordSameAsOld
	}

	histories, err := s.authRepo.GetPasswordHistoryByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, history := range histories {
		if err := utils.ComparePassword(history.PasswordHash, input.NewPassword); err != nil {
			return errors.New("không thể sử dụng lại mật khẩu trước đó")
		}
	}

	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	if err := s.authRepo.SavePasswordHistory(ctx, userID, user.PasswordHash); err != nil {
		return err
	}

	return s.authRepo.UpdatePassword(ctx, userID, hashedPassword)
}
