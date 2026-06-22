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

func (r *ProfileRepository) FindByPhoneNumber(ctx context.Context, phoneNumber string, excludeUserID string) (*models.Profile, error) {
    var profile models.Profile
    tx := r.db.WithContext(ctx).
        Where("phone_number = ? AND user_id != ?", phoneNumber, excludeUserID).
        First(&profile)
    if tx.Error != nil {
        if tx.Error == gorm.ErrRecordNotFound {
            return nil, nil // Không tìm thấy = OK
        }
        return nil, fmt.Errorf("find by phone: %w", tx.Error)
    }
    return &profile, nil
}

func (r *ProfileRepository) Update(ctx context.Context, userID string, profile *models.Profile) (*models.Profile, error) {
    tx := r.db.WithContext(ctx).Where("user_id = ?", userID).Updates(profile)
    if tx.Error != nil {
        return nil, fmt.Errorf("update profile: %w", tx.Error)
    }
    if tx.RowsAffected == 0 {
        return nil, fmt.Errorf("profile not found")
    }
    return profile, nil
}