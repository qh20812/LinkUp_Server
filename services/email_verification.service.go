package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"linkup/config"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
)

type EmailVerificationService struct {
	verifyRepo       *repository.EmailVerificationRepository
	authRepo         *repository.AuthRepository
	adminSettingsRepo *repository.AdminSettingsRepository
	env              config.Env
}

func NewEmailVerificationService(
	verifyRepo *repository.EmailVerificationRepository,
	authRepo *repository.AuthRepository,
	adminSettingsRepo *repository.AdminSettingsRepository,
	env config.Env,
) *EmailVerificationService {
	return &EmailVerificationService{
		verifyRepo:        verifyRepo,
		authRepo:          authRepo,
		adminSettingsRepo: adminSettingsRepo,
		env:               env,
	}
}

func (s *EmailVerificationService) SendVerificationEmail(ctx context.Context, userID, email, userName string) error {
	if err := s.verifyRepo.DeleteUserOldTokens(ctx, userID); err != nil {
		return err
	}

	token := s.generateToken()
	vt := models.NewEmailVerificationToken(userID, token, 1*time.Hour)
	vt.ID = utils.GenerateUUID()

	if _, err := s.verifyRepo.Create(ctx, &vt); err != nil {
		return err
	}

	frontendURL := os.Getenv("FRONTEND_RESET_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	verifyLink := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, token)

	if err := utils.SendVerificationEmail(email, userName, verifyLink); err != nil {
		fmt.Printf("Warning: Failed to send verification email: %v\n", err)
	}

	return nil
}

func (s *EmailVerificationService) VerifyEmail(ctx context.Context, tokenStr string) (*dto.VerifyEmailResponse, error) {
	vt, err := s.verifyRepo.FindByToken(ctx, tokenStr)
	if err != nil {
		if errors.Is(err, repository.ErrVerificationTokenNotFound) {
			return &dto.VerifyEmailResponse{
				Message:  "Token xác thực không hợp lệ",
				Verified: false,
			}, nil
		}
		return nil, err
	}

	if vt.IsExpired() {
		return &dto.VerifyEmailResponse{
			Message:  "Token xác thực đã hết hạn",
			Verified: false,
		}, nil
	}

	if vt.IsUsed() {
		return &dto.VerifyEmailResponse{
			Message:  "Token xác thực đã được sử dụng",
			Verified: false,
		}, nil
	}

	now := time.Now().UTC()
	if err := s.authRepo.UpdateEmailVerifiedAt(ctx, vt.UserID, now); err != nil {
		return nil, err
	}

	if err := s.verifyRepo.MarkAsUsed(ctx, vt.ID); err != nil {
		return nil, err
	}

	user, err := s.authRepo.FindByID(ctx, vt.UserID)
	if err != nil {
		return nil, err
	}

	role, err := s.authRepo.GetUserRole(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	accessTTL := s.getJWTExpiryMinutes(ctx)
	refreshDays := s.getRefreshTokenExpiryDays(ctx)
	refreshTTL := time.Duration(refreshDays) * 24 * time.Hour

	accessToken, refreshToken, err := utils.GenerateTokenPair(s.env.JWTSecret, user.ID, user.Email, role, user.TokenVersion, time.Duration(accessTTL)*time.Minute, refreshTTL)
	if err != nil {
		return nil, err
	}

	return &dto.VerifyEmailResponse{
		Message:      "Xác thực email thành công",
		Verified:     true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Role:         role,
	}, nil
}

func (s *EmailVerificationService) ResendVerification(ctx context.Context, email string) (*dto.ResendVerificationResponse, error) {
	user, err := s.authRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return &dto.ResendVerificationResponse{
				Message: "Nếu email tồn tại, bạn sẽ nhận được hướng dẫn xác thực",
			}, nil
		}
		return nil, err
	}

	if user.IsEmailVerified() {
		return &dto.ResendVerificationResponse{
			Message: "Email đã được xác thực trước đó",
		}, nil
	}

	if err := s.SendVerificationEmail(ctx, user.ID, user.Email, user.Username); err != nil {
		return nil, err
	}

	return &dto.ResendVerificationResponse{
		Message: "Nếu email tồn tại, bạn sẽ nhận được hướng dẫn xác thực",
	}, nil
}

func (s *EmailVerificationService) generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *EmailVerificationService) getJWTExpiryMinutes(ctx context.Context) int {
	cfg, err := s.adminSettingsRepo.GetByKey(ctx, "jwt_expiry_minutes")
	if err != nil || cfg == nil {
		return clampEmail(s.env.JWTExpiresIn, 1, 60)
	}
	n, err := strconv.Atoi(cfg.Value)
	if err != nil || n < 1 {
		return clampEmail(s.env.JWTExpiresIn, 1, 60)
	}
	if n > 60 {
		return 60
	}
	return n
}

func (s *EmailVerificationService) getRefreshTokenExpiryDays(ctx context.Context) int {
	cfg, err := s.adminSettingsRepo.GetByKey(ctx, "refresh_token_expiry_days")
	if err != nil || cfg == nil {
		return 7
	}
	n, err := strconv.Atoi(cfg.Value)
	if err != nil || n < 1 {
		return 7
	}
	if n > 30 {
		return 30
	}
	return n
}

func clampEmail(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
