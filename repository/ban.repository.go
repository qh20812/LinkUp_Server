package repository

import (
	"context"
	"errors"
	"fmt"
	"linkup/models"

	"gorm.io/gorm"
)

var ErrBanNotFound = errors.New("không tìm thấy bản ghi ban")

type BanRepository struct {
	db *gorm.DB
}

func NewBanRepository(db *gorm.DB) *BanRepository {
	return &BanRepository{db: db}
}

func (r *BanRepository) CreateBan(ctx context.Context, ban *models.Ban) error {
	if err := r.db.WithContext(ctx).Create(ban).Error; err != nil {
		return fmt.Errorf("create ban: %w", err)
	}
	return nil
}

func (r *BanRepository) GetLatestBanByUserID(ctx context.Context, userID string) (*models.Ban, error) {
	var ban models.Ban
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(1).
		First(&ban).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrBanNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get latest ban: %w", err)
	}
	return &ban, nil
}
