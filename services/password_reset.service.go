package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"linkup/config"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
)

type PasswordResetService struct {
	resetRepo  *repository.PasswordResetRepository
	authRepo   *repository.AuthRepository
	validation *validations.AuthValidation
	env        config.Env
}

func NewPasswordResetService(resetRepo *repository.PasswordResetRepository, authRepo *repository.AuthRepository, validation *validations.AuthValidation, env config.Env) *PasswordResetService {
	return &PasswordResetService{
		resetRepo:  resetRepo,
		authRepo:   authRepo,
		validation: validation,
		env:        env,
	}
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

	token := s.generateResetToken()
	resetToken := models.NewPasswordResetToken(user.ID, token, 10*time.Minute)

	if _, err := s.resetRepo.Create(ctx, &resetToken); err != nil {
		return dto.ForgotPasswordResponse{}, err
	}

	frontendURL := os.Getenv("FRONTEND_RESET_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	resetLink := fmt.Sprintf("%s?token=%s", frontendURL, token)

	if err := utils.SendResetPasswordEmail(user.Email, user.Username, resetLink); err != nil {
		fmt.Printf("Warning: Failed to send email: %v\n", err)
	}

	return dto.ForgotPasswordResponse{
		Message: "Nếu email tồn tại, bạn sẽ nhận được hướng dẫn đặt lại mật khẩu",
		Token:   token, // ⚠️ ONLY FOR TESTING
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
			return dto.ResetPasswordResponse{}, errors.New("token không hợp lệ")
		}
		return dto.ResetPasswordResponse{}, err
	}

	if resetToken.IsExpired() {
		return dto.ResetPasswordResponse{}, errors.New("token đã hết hạn")
	}

	if resetToken.IsUsed() {
		return dto.ResetPasswordResponse{}, errors.New("token đã được sử dụng")
	}

	if err := s.validation.ValidatePassword(input.NewPassword); err != nil {
		return dto.ResetPasswordResponse{}, err
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

// Helper: Tạo random token
func (s *PasswordResetService) generateResetToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
