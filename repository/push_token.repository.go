package repository

import (
	"context"
	"fmt"

	"linkup/models"

	"gorm.io/gorm"
)

type PushTokenRepository struct {
	db *gorm.DB
}

func NewPushTokenRepository(db *gorm.DB) *PushTokenRepository {
	return &PushTokenRepository{db: db}
}

func (r *PushTokenRepository) Upsert(ctx context.Context, token *models.PushToken) error {
	var existing models.PushToken
	err := r.db.WithContext(ctx).Where("push_token = ?", token.PushToken).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		tx := r.db.WithContext(ctx).Create(token)
		if tx.Error != nil {
			return fmt.Errorf("create push token: %w", tx.Error)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("find push token: %w", err)
	}
	existing.UserID = token.UserID
	existing.Platform = token.Platform
	existing.UpdatedAt = token.UpdatedAt
	tx := r.db.WithContext(ctx).Save(&existing)
	if tx.Error != nil {
		return fmt.Errorf("update push token: %w", tx.Error)
	}
	return nil
}

func (r *PushTokenRepository) FindByUserID(ctx context.Context, userID string) ([]models.PushToken, error) {
	var tokens []models.PushToken
	tx := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tokens)
	if tx.Error != nil {
		return nil, fmt.Errorf("find push tokens by user: %w", tx.Error)
	}
	return tokens, nil
}

func (r *PushTokenRepository) FindByUserIDs(ctx context.Context, userIDs []string) ([]models.PushToken, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	var tokens []models.PushToken
	tx := r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&tokens)
	if tx.Error != nil {
		return nil, fmt.Errorf("find push tokens by users: %w", tx.Error)
	}
	return tokens, nil
}

func (r *PushTokenRepository) DeleteByPushToken(ctx context.Context, pushToken string) error {
	tx := r.db.WithContext(ctx).Where("push_token = ?", pushToken).Delete(&models.PushToken{})
	if tx.Error != nil {
		return fmt.Errorf("delete push token: %w", tx.Error)
	}
	return nil
}

func (r *PushTokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	tx := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.PushToken{})
	if tx.Error != nil {
		return fmt.Errorf("delete push tokens by user: %w", tx.Error)
	}
	return nil
}
