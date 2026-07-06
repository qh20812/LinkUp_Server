package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"linkup/models"

	"gorm.io/gorm"
)

var ErrCallBusy = errors.New("người dùng đang bận")

type CallRepository struct {
	db *gorm.DB
}

func NewCallRepository(db *gorm.DB) *CallRepository {
	return &CallRepository{db: db}
}

func (r *CallRepository) Create(ctx context.Context, call *models.Call) error {
	tx := r.db.WithContext(ctx).Create(call)
	if tx.Error != nil {
		return fmt.Errorf("create call: %w", tx.Error)
	}
	return nil
}

func (r *CallRepository) CreateIfNotBusy(ctx context.Context, call *models.Call) (bool, error) {
	var isBusy bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Call{}).
			Where("(caller_id = ? OR callee_id = ?) AND status IN ?",
				call.CalleeID, call.CalleeID,
				[]models.CallStatus{
					models.CallStatusCalling,
					models.CallStatusRinging,
					models.CallStatusConnected,
				}).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			isBusy = true
			return nil
		}
		return tx.Create(call).Error
	})
	if err != nil {
		return false, fmt.Errorf("create call: %w", err)
	}
	return isBusy, nil
}

func (r *CallRepository) FindByID(ctx context.Context, id string) (*models.Call, error) {
	var call models.Call
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&call).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find call by id: %w", err)
	}
	return &call, nil
}

func (r *CallRepository) FindActiveByUserID(ctx context.Context, userID string) (*models.Call, error) {
	return r.FindActiveByUser(ctx, userID)
}

func (r *CallRepository) FindActiveByUser(ctx context.Context, userID string) (*models.Call, error) {
	var call models.Call
	err := r.db.WithContext(ctx).
		Where("(caller_id = ? OR callee_id = ?) AND status IN ?", userID, userID, []models.CallStatus{
			models.CallStatusCalling,
			models.CallStatusRinging,
			models.CallStatusConnected,
		}).
		Order("created_at DESC").
		First(&call).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find active call by user: %w", err)
	}
	return &call, nil
}

func (r *CallRepository) FindActiveBetween(ctx context.Context, userA, userB string) (*models.Call, error) {
	var call models.Call
	err := r.db.WithContext(ctx).
		Where("((caller_id = ? AND callee_id = ?) OR (caller_id = ? AND callee_id = ?)) AND status IN ?",
			userA, userB, userB, userA,
			[]models.CallStatus{
				models.CallStatusCalling,
				models.CallStatusRinging,
				models.CallStatusConnected,
			}).
		Order("created_at DESC").
		First(&call).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find active call between users: %w", err)
	}
	return &call, nil
}

func (r *CallRepository) UpdateStatus(ctx context.Context, id string, status models.CallStatus, startedAt, endedAt *time.Time, duration int) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if startedAt != nil {
		updates["started_at"] = *startedAt
	}
	if endedAt != nil {
		updates["ended_at"] = *endedAt
	}
	if duration > 0 {
		updates["duration"] = duration
	}
	tx := r.db.WithContext(ctx).Model(&models.Call{}).Where("id = ?", id).Updates(updates)
	if tx.Error != nil {
		return fmt.Errorf("update call status: %w", tx.Error)
	}
	return nil
}

func (r *CallRepository) UpdateMuted(ctx context.Context, id string, mutedCaller, mutedCallee *bool) error {
	updates := map[string]interface{}{}
	if mutedCaller != nil {
		updates["muted_caller"] = *mutedCaller
	}
	if mutedCallee != nil {
		updates["muted_callee"] = *mutedCallee
	}
	if len(updates) == 0 {
		return nil
	}
	tx := r.db.WithContext(ctx).Model(&models.Call{}).Where("id = ?", id).Updates(updates)
	if tx.Error != nil {
		return fmt.Errorf("update call mute: %w", tx.Error)
	}
	return nil
}

func (r *CallRepository) UpdateVideoEnabled(ctx context.Context, id string, videoCaller, videoCallee *bool) error {
	updates := map[string]interface{}{}
	if videoCaller != nil {
		updates["video_enabled_caller"] = *videoCaller
	}
	if videoCallee != nil {
		updates["video_enabled_callee"] = *videoCallee
	}
	if len(updates) == 0 {
		return nil
	}
	tx := r.db.WithContext(ctx).Model(&models.Call{}).Where("id = ?", id).Updates(updates)
	if tx.Error != nil {
		return fmt.Errorf("update call video: %w", tx.Error)
	}
	return nil
}

func (r *CallRepository) GetHistory(ctx context.Context, userID string, limit, offset int) ([]models.Call, error) {
	var calls []models.Call
	err := r.db.WithContext(ctx).
		Where("caller_id = ? OR callee_id = ?", userID, userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&calls).Error
	if err != nil {
		return nil, fmt.Errorf("get call history: %w", err)
	}
	return calls, nil
}

func (r *CallRepository) CountHistory(ctx context.Context, userID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.Call{}).
		Where("caller_id = ? OR callee_id = ?", userID, userID).
		Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("count call history: %w", err)
	}
	return total, nil
}
