package repository

import (
	"context"
	"errors"
	"fmt"

	"linkup/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type E2ERepository struct {
	db *gorm.DB
}

func NewE2ERepository(db *gorm.DB) *E2ERepository {
	return &E2ERepository{db: db}
}

func (r *E2ERepository) UpsertUserKey(ctx context.Context, key *models.UserE2EKey) error {
	tx := r.db.WithContext(ctx).Save(key)
	if tx.Error != nil {
		return fmt.Errorf("upsert user e2e key: %w", tx.Error)
	}
	return nil
}

func (r *E2ERepository) GetUserKey(ctx context.Context, userID string) (*models.UserE2EKey, error) {
	var key models.UserE2EKey
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user e2e key: %w", err)
	}
	return &key, nil
}

// UpsertChatKey persists a wrapped chat key entry. First-writer-wins: một khi
// (chat_id, user_id) đã tồn tại thì không ghi đè — bảo đảm hai client tạo khóa
// đồng thời không ghi đè khóa của nhau, cả hai cùng quy về khóa duy nhất.
func (r *E2ERepository) UpsertChatKey(ctx context.Context, key *models.ChatE2EKey) error {
	tx := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chat_id"}, {Name: "user_id"}},
		DoNothing: true,
	}).Create(key)
	if tx.Error != nil {
		return fmt.Errorf("insert chat e2e key: %w", tx.Error)
	}
	return nil
}

func (r *E2ERepository) UpsertChatKeys(ctx context.Context, keys []models.ChatE2EKey) error {
	if len(keys) == 0 {
		return nil
	}
	tx := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chat_id"}, {Name: "user_id"}},
		DoNothing: true,
	}).Create(&keys)
	if tx.Error != nil {
		return fmt.Errorf("insert chat e2e keys: %w", tx.Error)
	}
	return nil
}

func (r *E2ERepository) GetChatKey(ctx context.Context, chatID, userID string) (*models.ChatE2EKey, error) {
	var key models.ChatE2EKey
	err := r.db.WithContext(ctx).Where("chat_id = ? AND user_id = ?", chatID, userID).First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get chat e2e key: %w", err)
	}
	return &key, nil
}
