package repository

import (
	"context"
	"fmt"
	"linkup/models"

	"gorm.io/gorm"
)

type FollowRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) *FollowRepository {
	return &FollowRepository{db: db}
}

func (r *FollowRepository) Create(ctx context.Context, follow *models.Follow) error {
	tx := r.db.WithContext(ctx).Create(follow)
	if tx.Error != nil {
		return fmt.Errorf("create follow: %w", tx.Error)
	}
	return nil
}

func (r *FollowRepository) Delete(ctx context.Context, followerID, followingID string) error {
	tx := r.db.WithContext(ctx).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&models.Follow{})
	if tx.Error != nil {
		return fmt.Errorf("delete follow: %w", tx.Error)
	}
	return nil
}

func (r *FollowRepository) IsFollowing(ctx context.Context, followerID, followingID string) (bool, error) {
	var count int64
	tx := r.db.WithContext(ctx).
		Model(&models.Follow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count)
	if tx.Error != nil {
		return false, fmt.Errorf("check is following: %w", tx.Error)
	}
	return count > 0, nil
}

func (r *FollowRepository) GetFollowerCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	tx := r.db.WithContext(ctx).
		Model(&models.Follow{}).
		Where("following_id = ?", userID).
		Count(&count)
	if tx.Error != nil {
		return 0, fmt.Errorf("get follower count: %w", tx.Error)
	}
	return count, nil
}

func (r *FollowRepository) GetFollowingCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	tx := r.db.WithContext(ctx).
		Model(&models.Follow{}).
		Where("follower_id = ?", userID).
		Count(&count)
	if tx.Error != nil {
		return 0, fmt.Errorf("get following count: %w", tx.Error)
	}
	return count, nil
}
