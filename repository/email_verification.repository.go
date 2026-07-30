package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

func (r *EmailVerificationRepository) DB() *gorm.DB {
	return r.db
}

func (r *EmailVerificationRepository) Create(ctx context.Context, token *models.EmailVerificationToken) (*models.EmailVerificationToken, error) {
	tx := r.db.WithContext(ctx).Create(token)
	if tx.Error != nil {
		return nil, fmt.Errorf("create email verification token: %w", tx.Error)
	}
	return token, nil
}

func (r *EmailVerificationRepository) FindByToken(ctx context.Context, token string) (*models.EmailVerificationToken, error) {
	hash := sha256.Sum256([]byte(token))
	hashed := hex.EncodeToString(hash[:])
	var t models.EmailVerificationToken
	err := r.db.WithContext(ctx).Where("token = ?", hashed).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrVerificationTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find email verification token: %w", err)
	}
	return &t, nil
}

func (r *EmailVerificationRepository) FindLatestByUserID(ctx context.Context, userID string) (*models.EmailVerificationToken, error) {
	var t models.EmailVerificationToken
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find latest by user id: %w", err)
	}
	return &t, nil
}

func (r *EmailVerificationRepository) MarkAsUsedTx(ctx context.Context, tx *gorm.DB, tokenID string) error {
	now := time.Now().UTC()
	result := tx.Model(&models.EmailVerificationToken{}).
		Where("id = ?", tokenID).
		Update("used_at", now)
	if result.Error != nil {
		return fmt.Errorf("mark email verification token as used: %w", result.Error)
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

func (r *EmailVerificationRepository) DeleteExpired(ctx context.Context) error {
	tx := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&models.EmailVerificationToken{})
	if tx.Error != nil {
		return fmt.Errorf("delete expired email verification tokens: %w", tx.Error)
	}
	return nil
}
