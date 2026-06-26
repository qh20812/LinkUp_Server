package repository

import (
	"context"
	"errors"
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

func (r *FriendRepository) FindBySenderAndReceiver(ctx context.Context, senderID, receiverID string) (*models.Friend, error) {
	var friend models.Friend
	err := r.db.WithContext(ctx).
		Where("sender_id = ? AND receiver_id = ?", senderID, receiverID).
		First(&friend).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find friend request: %w", err)
	}
	return &friend, nil
}

func (r *FriendRepository) Create(ctx context.Context, friend *models.Friend) error {
	tx := r.db.WithContext(ctx).Create(friend)
	if tx.Error != nil {
		return fmt.Errorf("create friend request: %w", tx.Error)
	}
	return nil
}

func (r *FriendRepository) FindByID(ctx context.Context, id string) (*models.Friend, error) {
	var friend models.Friend
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&friend).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find friend request by id: %w", err)
	}
	return &friend, nil
}

func (r *FriendRepository) FindByReceiverID(ctx context.Context, receiverID string, status models.FriendStatus) ([]models.Friend, error) {
	var friends []models.Friend
	q := r.db.WithContext(ctx).Where("receiver_id = ?", receiverID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("created_at DESC").Find(&friends).Error
	if err != nil {
		return nil, fmt.Errorf("find friend requests by receiver: %w", err)
	}
	return friends, nil
}

func (r *FriendRepository) FindBySenderID(ctx context.Context, senderID string, status models.FriendStatus) ([]models.Friend, error) {
	var friends []models.Friend
	q := r.db.WithContext(ctx).Where("sender_id = ?", senderID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("created_at DESC").Find(&friends).Error
	if err != nil {
		return nil, fmt.Errorf("find friend requests by sender: %w", err)
	}
	return friends, nil
}

func (r *FriendRepository) UpdateStatus(ctx context.Context, id string, status models.FriendStatus) error {
	tx := r.db.WithContext(ctx).Model(&models.Friend{}).Where("id = ?", id).Update("status", status)
	if tx.Error != nil {
		return fmt.Errorf("update friend request status: %w", tx.Error)
	}
	return nil
}

func (r *FriendRepository) Delete(ctx context.Context, id string) error {
	tx := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Friend{})
	if tx.Error != nil {
		return fmt.Errorf("delete friend request: %w", tx.Error)
	}
	return nil
}
