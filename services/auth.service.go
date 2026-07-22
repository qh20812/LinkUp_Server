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
	banRepo     *repository.BanRepository
	env         config.Env
}

func NewAuthService(authRepo *repository.AuthRepository, profileRepo *repository.ProfileRepository, banRepo *repository.BanRepository, env config.Env) *AuthService {
	return &AuthService{
		authRepo:    authRepo,
		profileRepo: profileRepo,
		banRepo:     banRepo,
		env:         env,
	}
}

func (s *AuthService) Register(ctx context.Context, input dto.RegisterInput) (dto.AuthResponse, error) {
	email := normalizeEmail(input.Email)

	if _, err := s.authRepo.FindByEmail(ctx, email); err == nil {
		return dto.AuthResponse{}, errors.New("email đã tồn tại")
	} else if !errors.Is(err, repository.ErrUserNotFound) {
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
		ID:                utils.GenerateUUID(),
		Username:          username,
		Email:             email,
		PasswordHash:      hashedPassword,
		Status:            models.UserStatusActive,
		StorageQuotaBytes: models.DefaultStorageQuotaBytes,
		StorageUsedBytes:  0,
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

	accessToken, refreshToken, err := s.generateTokens(createdUser, "USER")
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

	if err := s.ensureBanStatus(ctx, user); err != nil {
		return dto.AuthResponse{}, err
	}

	if !user.IsActive() {
		return dto.AuthResponse{}, fmt.Errorf("tài khoản chưa được kích hoạt")
	}

	if err := utils.ComparePassword(user.PasswordHash, input.Password); err != nil {
		return dto.AuthResponse{}, errors.New("email hoặc mật khẩu không hợp lệ")
	}

	role, err := s.authRepo.GetUserRole(ctx, user.ID)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	accessToken, refreshToken, err := s.generateTokens(user, role)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	return buildAuthResponse(*user, accessToken, refreshToken, s.accessTTL(), s.refreshTTL()), nil
}

func (s *AuthService) generateTokens(user *models.User, role string) (string, string, error) {
	return utils.GenerateTokenPair(s.env.JWTSecret, user.ID, user.Email, role, s.accessTTL(), s.refreshTTL())
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
			UsedBytes:  user.StorageUsedBytes,
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

func (s *AuthService) RefreshToken(ctx context.Context, input dto.RefreshTokenInput) (dto.TokenResponse, error) {
	token, err := utils.ParseToken(s.env.JWTSecret, input.RefreshToken)
	if err != nil || !token.Valid {
		return dto.TokenResponse{}, errors.New("refresh token không hợp lệ hoặc đã hết hạn")
	}

	claims, ok := token.Claims.(*utils.TokenClaims)
	if !ok {
		return dto.TokenResponse{}, errors.New("refresh token không hợp lệ")
	}

	if claims.TokenType != "refresh" {
		return dto.TokenResponse{}, errors.New("token không phải refresh token")
	}

	user, err := s.authRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.TokenResponse{}, errors.New("người dùng không tồn tại")
		}
		return dto.TokenResponse{}, err
	}

	if !user.IsActive() {
		return dto.TokenResponse{}, errors.New("tài khoản chưa được kích hoạt")
	}

	role, err := s.authRepo.GetUserRole(ctx, user.ID)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	accessToken, refreshToken, err := s.generateTokens(user, role)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTTL().Seconds()),
		RefreshTTLIn: int64(s.refreshTTL().Seconds()),
	}, nil
}

func (s *AuthService) ensureBanStatus(ctx context.Context, user *models.User) error {
	if user.Status != models.UserStatusBanned {
		return nil
	}

	ban, err := s.banRepo.GetLatestBanByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrBanNotFound) {
			return fmt.Errorf("tài khoản đang bị ban")
		}
		return err
	}

	if ban.ExpiresAt != nil && ban.ExpiresAt.Before(time.Now().UTC()) {
		if err := s.authRepo.UpdateUserStatus(ctx, user.ID, models.UserStatusActive); err != nil {
			return err
		}
		user.Status = models.UserStatusActive
		return nil
	}

	if ban.ExpiresAt != nil {
		return fmt.Errorf("tài khoản đang bị ban đến %s", ban.ExpiresAt.Format("2006-01-02 15:04:05"))
	}

	return errors.New("tài khoản bị ban vô thời hạn")
}
