package repository

import (
	"context"
	"fmt"

	"linkup/models"

	"gorm.io/gorm"
)

type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) Create(ctx context.Context, profile *models.Profile) (*models.Profile, error) {
	tx := r.db.WithContext(ctx).Create(profile)
	if tx.Error != nil {
		return nil, fmt.Errorf("insert profile: %w", tx.Error)
	}
	return profile, nil
}
