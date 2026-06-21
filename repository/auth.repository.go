package repository

import (
	"context"
	"errors"
	"fmt"

	"linkup/models"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	tx := r.db.WithContext(ctx).Create(user)
	if tx.Error != nil {
		return nil, fmt.Errorf("insert user: %w", tx.Error)
	}
	return user, nil
}

func (r *AuthRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &user, nil
}

func (r *AuthRepository) IsUsernameTaken(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check username: %w", err)
	}
	return count > 0, nil
}

func (r *AuthRepository) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	tx := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("password_hash", hashedPassword)
	if tx.Error != nil {
		return fmt.Errorf("update password: %w", tx.Error)
	}
	return nil
}

func (r *AuthRepository) FindByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}
