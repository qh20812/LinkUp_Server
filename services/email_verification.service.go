package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"linkup/config"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"

	"gorm.io/gorm"
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
	if err := s.verifyRepo.DeleteExpired(ctx); err != nil {
		return err
	}
	if err := s.verifyRepo.DeleteUserOldTokens(ctx, userID); err != nil {
		return err
	}

	rawToken, hashedToken := s.generateToken()
	vt := models.NewEmailVerificationToken(userID, hashedToken, 1*time.Hour)
	vt.ID = utils.GenerateUUID()

	if _, err := s.verifyRepo.Create(ctx, &vt); err != nil {
		return err
	}

	verifyLink := fmt.Sprintf("%s/verify-email?token=%s", s.env.FrontendURL, rawToken)

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
				Message:  errorsapp.New(errorsapp.ErrCodeVerifyTokenNotFound).Message,
				Verified: false,
			}, nil
		}
		return nil, err
	}

	if vt.IsExpired() {
		return &dto.VerifyEmailResponse{
			Message:  errorsapp.New(errorsapp.ErrCodeVerifyTokenExpired).Message,
			Verified: false,
		}, nil
	}

	if vt.IsUsed() {
		return &dto.VerifyEmailResponse{
			Message:  errorsapp.New(errorsapp.ErrCodeVerifyTokenUsed).Message,
			Verified: false,
		}, nil
	}

	now := time.Now().UTC()
	if err := s.verifyRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.authRepo.UpdateEmailVerifiedAtTx(ctx, tx, vt.UserID, now); err != nil {
			return err
		}
		if err := s.verifyRepo.MarkAsUsedTx(ctx, tx, vt.ID); err != nil {
			return err
		}
		return nil
	}); err != nil {
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
			Message: errorsapp.New(errorsapp.ErrCodeVerifyAlreadyDone).Message,
		}, nil
	}

	last, err := s.verifyRepo.FindLatestByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if last != nil && time.Since(last.CreatedAt) < 60*time.Second {
		remaining := 60 - int(time.Since(last.CreatedAt).Seconds())
		return &dto.ResendVerificationResponse{
			Message: errorsapp.Newf(errorsapp.ErrCodeVerifyRateLimited, map[string]any{"seconds": remaining}).Message,
		}, nil
	}

	if err := s.SendVerificationEmail(ctx, user.ID, user.Email, user.Username); err != nil {
		return nil, err
	}

	return &dto.ResendVerificationResponse{
		Message: "Nếu email tồn tại, bạn sẽ nhận được hướng dẫn xác thực",
	}, nil
}

func (s *EmailVerificationService) generateToken() (raw string, hashed string) {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	raw = hex.EncodeToString(bytes)
	sum := sha256.Sum256([]byte(raw))
	hashed = hex.EncodeToString(sum[:])
	return
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
