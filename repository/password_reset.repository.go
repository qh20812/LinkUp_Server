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

var (
	ErrResetTokenNotFound = errors.New("không tìm thấy token đặt lại mật khẩu")
	ErrResetTokenExpired  = errors.New("token đặt lại mật khẩu đã hết hạn")
	ErrResetTokenUsed     = errors.New("token đặt lại mật khẩu đã được sử dụng")
)

type PasswordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(ctx context.Context, token *models.PasswordResetToken) (*models.PasswordResetToken, error) {
	tx := r.db.WithContext(ctx).Create(token)
	if tx.Error != nil {
		return nil, fmt.Errorf("create reset token: %w", tx.Error)
	}
	return token, nil
}

func (r *PasswordResetRepository) FindByToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	hash := sha256.Sum256([]byte(token))
	hashed := hex.EncodeToString(hash[:])
	var resetToken models.PasswordResetToken
	err := r.db.WithContext(ctx).Where("token = ?", hashed).First(&resetToken).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrResetTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find reset token: %w", err)
	}
	return &resetToken, nil
}

func (r *PasswordResetRepository) MarkAsUsed(ctx context.Context, tokenID string) error {
	now := time.Now().UTC()
	tx := r.db.WithContext(ctx).Model(&models.PasswordResetToken{}).
		Where("id = ?", tokenID).
		Update("used_at", now)
	if tx.Error != nil {
		return fmt.Errorf("mark token as used: %w", tx.Error)
	}
	return nil
}

func (r *PasswordResetRepository) DeleteUserOldToken(ctx context.Context, userID string) error {
	tx := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.PasswordResetToken{})
	if tx.Error != nil {
		return fmt.Errorf("delete old token: %w", tx.Error)
	}
	return nil
}

func (r *PasswordResetRepository) DeleteExpired(ctx context.Context) error {
	tx := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&models.PasswordResetToken{})
	if tx.Error != nil {
		return fmt.Errorf("delete expired reset tokens: %w", tx.Error)
	}
	return nil
}
