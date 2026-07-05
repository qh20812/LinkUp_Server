package repository

import (
	"context"
	"linkup/models"

	"gorm.io/gorm"
)

type ModerationRepository struct {
	db *gorm.DB
}

func NewModerationRepository(db *gorm.DB) *ModerationRepository {
	return &ModerationRepository{db: db}
}

func (r *ModerationRepository) CreateLog(ctx context.Context, log *models.ModerationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
