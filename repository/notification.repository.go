package repository

import (
	"context"
	"fmt"

	"linkup/models"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	tx := r.db.WithContext(ctx).Create(notification)
	if tx.Error != nil {
		return fmt.Errorf("create notification: %w", tx.Error)
	}
	return nil
}

func (r *NotificationRepository) FindByReceiverID(ctx context.Context, receiverID string, limit, offset int, unreadOnly bool) ([]models.Notification, error) {
	var notifications []models.Notification
	q := r.db.WithContext(ctx).Where("receiver_id = ?", receiverID)
	if unreadOnly {
		q = q.Where("is_read = ?", false)
	}
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&notifications).Error
	if err != nil {
		return nil, fmt.Errorf("find notifications: %w", err)
	}
	return notifications, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id, receiverID string) error {
	tx := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ? AND receiver_id = ?", id, receiverID).
		Update("is_read", true)
	if tx.Error != nil {
		return fmt.Errorf("mark notification as read: %w", tx.Error)
	}
	return nil
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, receiverID string) error {
	tx := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("receiver_id = ? AND is_read = ?", receiverID, false).
		Update("is_read", true)
	if tx.Error != nil {
		return fmt.Errorf("mark all notifications as read: %w", tx.Error)
	}
	return nil
}

func (r *NotificationRepository) CountByReceiverID(ctx context.Context, receiverID string, unreadOnly bool) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&models.Notification{}).Where("receiver_id = ?", receiverID)
	if unreadOnly {
		q = q.Where("is_read = ?", false)
	}
	err := q.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count notifications: %w", err)
	}
	return count, nil
}

func (r *NotificationRepository) GetUnreadCount(ctx context.Context, receiverID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("receiver_id = ? AND is_read = ?", receiverID, false).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("get unread count: %w", err)
	}
	return count, nil
}

// CreateBulk tạo nhiều notifications trong 1 lần insert.
func (r *NotificationRepository) CreateBulk(ctx context.Context, notifications []models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}
	tx := r.db.WithContext(ctx).CreateInBatches(&notifications, 100)
	if tx.Error != nil {
		return fmt.Errorf("create notifications bulk: %w", tx.Error)
	}
	return nil
}
