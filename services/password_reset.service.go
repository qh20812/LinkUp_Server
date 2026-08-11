package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"linkup/config"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
)

type PasswordResetService struct {
	resetRepo          *repository.PasswordResetRepository
	authRepo           *repository.AuthRepository
	adminSettingsRepo  *repository.AdminSettingsRepository
	validation         *validations.AuthValidation
	env                config.Env
}

func NewPasswordResetService(resetRepo *repository.PasswordResetRepository, authRepo *repository.AuthRepository, adminSettingsRepo *repository.AdminSettingsRepository, validation *validations.AuthValidation, env config.Env) *PasswordResetService {
	return &PasswordResetService{
		resetRepo:          resetRepo,
		authRepo:           authRepo,
		adminSettingsRepo:  adminSettingsRepo,
		validation:         validation,
		env:                env,
	}
}

func (s *PasswordResetService) getMinPasswordLength(ctx context.Context) int {
	cfg, err := s.adminSettingsRepo.GetByKey(ctx, "password_min_length")
	if err != nil || cfg == nil {
		return 8
	}
	n, err := strconv.Atoi(cfg.Value)
	if err != nil || n < 8 {
		return 8
	}
	if n > 50 {
		return 50
	}
	return n
}

// Gửi email quên mật khẩu
func (s *PasswordResetService) ForgotPassword(ctx context.Context, input dto.ForgotPasswordInput) (dto.ForgotPasswordResponse, error) {
	user, err := s.authRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.ForgotPasswordResponse{
				Message: "Nếu email tồn tại, bạn sẽ nhận được hướng dẫn đặt lại mật khẩu",
			}, nil
		}
		return dto.ForgotPasswordResponse{}, err
	}

	if err := s.resetRepo.DeleteExpired(ctx); err != nil {
		return dto.ForgotPasswordResponse{}, err
	}
	if err := s.resetRepo.DeleteUserOldToken(ctx, user.ID); err != nil {
		return dto.ForgotPasswordResponse{}, err
	}

	rawToken, hashedToken := s.generateResetToken()
	resetToken := models.NewPasswordResetToken(user.ID, hashedToken, 10*time.Minute)
	resetToken.ID = utils.GenerateUUID()

	if _, err := s.resetRepo.Create(ctx, &resetToken); err != nil {
		return dto.ForgotPasswordResponse{}, err
	}

	frontendURL := os.Getenv("FRONTEND_RESET_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, rawToken)

	if err := utils.SendResetPasswordEmail(user.Email, user.Username, resetLink); err != nil {
		fmt.Printf("Warning: Failed to send email: %v\n", err)
	}

	return dto.ForgotPasswordResponse{
		Message: "Nếu email tồn tại, bạn sẽ nhận được hướng dẫn đặt lại mật khẩu",
	}, nil
}

// Xác minh token hợp lệ
func (s *PasswordResetService) VerifyResetToken(ctx context.Context, input dto.VerifyResetTokenInput) (dto.VerifyResetTokenResponse, error) {
	resetToken, err := s.resetRepo.FindByToken(ctx, input.Token)
	if err != nil {
		if errors.Is(err, repository.ErrResetTokenNotFound) {
			return dto.VerifyResetTokenResponse{
				Message: "Token không hợp lệ",
				Valid:   false,
			}, nil
		}
		return dto.VerifyResetTokenResponse{}, err
	}

	if resetToken.IsExpired() {
		return dto.VerifyResetTokenResponse{
			Message: "Token đã hết hạn",
			Valid:   false,
		}, nil
	}

	if resetToken.IsUsed() {
		return dto.VerifyResetTokenResponse{
			Message: "Token đã được sử dụng",
			Valid:   false,
		}, nil
	}

	return dto.VerifyResetTokenResponse{
		Message: "Token hợp lệ",
		Valid:   true,
	}, nil
}

// Đặt lại mật khẩu
func (s *PasswordResetService) ResetPassword(ctx context.Context, input dto.ResetPasswordInput) (dto.ResetPasswordResponse, error) {
	resetToken, err := s.resetRepo.FindByToken(ctx, input.Token)
	if err != nil {
		if errors.Is(err, repository.ErrResetTokenNotFound) {
			return dto.ResetPasswordResponse{}, errorsapp.New(errorsapp.ErrCodeResetTokenNotFound)
		}
		return dto.ResetPasswordResponse{}, err
	}

	if resetToken.IsExpired() {
		return dto.ResetPasswordResponse{}, errorsapp.New(errorsapp.ErrCodeResetTokenExpired)
	}

	if resetToken.IsUsed() {
		return dto.ResetPasswordResponse{}, errorsapp.New(errorsapp.ErrCodeResetTokenUsed)
	}

	if err := s.validation.ValidatePassword(input.NewPassword); err != nil {
		return dto.ResetPasswordResponse{}, err
	}

	minLen := s.getMinPasswordLength(ctx)
	if len(input.NewPassword) < minLen {
		return dto.ResetPasswordResponse{}, errorsapp.Newf(errorsapp.ErrCodeResetPasswordShort, map[string]any{"min": minLen})
	}
	if len(input.NewPassword) > 50 {
		return dto.ResetPasswordResponse{}, errorsapp.New(errorsapp.ErrCodeResetPasswordLong)
	}

	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return dto.ResetPasswordResponse{}, err
	}

	if err := s.authRepo.UpdatePassword(ctx, resetToken.UserID, hashedPassword); err != nil {
		return dto.ResetPasswordResponse{}, err
	}

	if err := s.resetRepo.MarkAsUsed(ctx, resetToken.ID); err != nil {
		return dto.ResetPasswordResponse{}, err
	}

	return dto.ResetPasswordResponse{
		Message: "Mật khẩu đã được đặt lại thành công",
	}, nil
}

// Helper: Tạo random token (raw dùng cho link email, hashed lưu DB)
func (s *PasswordResetService) generateResetToken() (raw string, hashed string) {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	raw = hex.EncodeToString(bytes)
	sum := sha256.Sum256([]byte(raw))
	hashed = hex.EncodeToString(sum[:])
	return
}
