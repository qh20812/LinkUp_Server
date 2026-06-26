package repository

import (
	"context"
	"fmt"
	"linkup/models"

	"gorm.io/gorm"
)

type FriendRepository struct {
	db *gorm.DB
}

func NewFriendRepository(db *gorm.DB) *FriendRepository {
	return &FriendRepository{db: db}
}

func (r *FriendRepository) IsAcceptedFriend(ctx context.Context, userA, userB string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("friends").
		Where("((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND status = ?", userA, userB, userB, userA, models.FriendStatusAccepted).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check friend status: %w", err)
	}
	return count > 0, nil
}
