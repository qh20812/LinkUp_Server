package repository

import (
	"context"
	"errors"
	"fmt"
	"linkup/models"
	"time"

	"gorm.io/gorm"
)

var ErrVerificationTokenNotFound = errors.New("không tìm thấy token xác thực email")

type EmailVerificationRepository struct {
	db *gorm.DB
}

func NewEmailVerificationRepository(db *gorm.DB) *EmailVerificationRepository {
	return &EmailVerificationRepository{db: db}
}

func (r *EmailVerificationRepository) Create(ctx context.Context, token *models.EmailVerificationToken) (*models.EmailVerificationToken, error) {
	tx := r.db.WithContext(ctx).Create(token)
	if tx.Error != nil {
		return nil, fmt.Errorf("create email verification token: %w", tx.Error)
	}
	return token, nil
}

func (r *EmailVerificationRepository) FindByToken(ctx context.Context, token string) (*models.EmailVerificationToken, error) {
	var t models.EmailVerificationToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrVerificationTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find email verification token: %w", err)
	}
	return &t, nil
}

func (r *EmailVerificationRepository) MarkAsUsed(ctx context.Context, tokenID string) error {
	now := time.Now().UTC()
	tx := r.db.WithContext(ctx).Model(&models.EmailVerificationToken{}).
		Where("id = ?", tokenID).
		Update("used_at", now)
	if tx.Error != nil {
		return fmt.Errorf("mark email verification token as used: %w", tx.Error)
	}
	return nil
}

func (r *EmailVerificationRepository) DeleteUserOldTokens(ctx context.Context, userID string) error {
	tx := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.EmailVerificationToken{})
	if tx.Error != nil {
		return fmt.Errorf("delete old email verification tokens: %w", tx.Error)
	}
	return nil
}
