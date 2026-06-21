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

func (r *ProfileRepository) FindByUserID(ctx context.Context, userID string) (*models.Profile, error) {
	var profile models.Profile
	tx := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("profile not found")
		}
		return nil, fmt.Errorf("find profile: %w", tx.Error)
	}
	return &profile, nil
}
