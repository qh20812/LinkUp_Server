package repository

import (
	"context"
	"errors"
	"fmt"

	"linkup/models"

	"gorm.io/gorm"
)

type BlockRepository struct {
	db *gorm.DB
}

func NewBlockRepository(db *gorm.DB) *BlockRepository {
	return &BlockRepository{db: db}
}

func (r *BlockRepository) FindByUserAndTarget(ctx context.Context, userID, targetUserID string) (*models.Block, error) {
	var block models.Block
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND blocked_user_id = ?", userID, targetUserID).
		First(&block).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find block: %w", err)
	}
	return &block, nil
}

func (r *BlockRepository) Create(ctx context.Context, block *models.Block) error {
	tx := r.db.WithContext(ctx).Create(block)
	if tx.Error != nil {
		return fmt.Errorf("insert block: %w", tx.Error)
	}
	return nil
}

func (r *BlockRepository) Delete(ctx context.Context, id string) error {
	tx := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Block{})
	if tx.Error != nil {
		return fmt.Errorf("delete block: %w", tx.Error)
	}
	return nil
}

func (r *BlockRepository) FindByUserID(ctx context.Context, userID string) ([]models.Block, error) {
	var blocks []models.Block
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&blocks).Error
	if err != nil {
		return nil, fmt.Errorf("find blocks: %w", err)
	}
	return blocks, nil
}
