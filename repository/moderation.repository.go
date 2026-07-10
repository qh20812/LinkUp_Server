package repository

import (
	"context"
	"fmt"
	"linkup/dto"
	"linkup/models"
	"time"

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

func (r *ModerationRepository) ListLogsByTarget(ctx context.Context, targetType models.ModerationTargetType, targetID string, pageSize, offset int) ([]dto.AdminModerationLogItem, error) {
	type logRow struct {
		ID            string    `gorm:"column:id"`
		ModeratorID   string    `gorm:"column:moderator_id"`
		ModeratorName string    `gorm:"column:moderator_name"`
		Action        string    `gorm:"column:action"`
		TargetType    string    `gorm:"column:target_type"`
		TargetID      string    `gorm:"column:target_id"`
		Reason        string    `gorm:"column:reason"`
		CreatedAt     time.Time `gorm:"column:created_at"`
	}
	var rows []logRow
	err := r.db.WithContext(ctx).
		Table("moderation_logs").
		Select(`moderation_logs.id, moderation_logs.moderator_id,
			COALESCE(profiles.display_name, '') AS moderator_name,
			moderation_logs.action, moderation_logs.target_type,
			moderation_logs.target_id, moderation_logs.reason,
			moderation_logs.created_at`).
		Joins("LEFT JOIN profiles ON profiles.user_id = moderation_logs.moderator_id").
		Where("moderation_logs.target_type = ? AND moderation_logs.target_id = ?", targetType, targetID).
		Order("moderation_logs.created_at DESC").
		Offset(offset).Limit(pageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list moderation logs: %w", err)
	}

	items := make([]dto.AdminModerationLogItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.AdminModerationLogItem{
			ID:            r.ID,
			ModeratorID:   r.ModeratorID,
			ModeratorName: r.ModeratorName,
			Action:        r.Action,
			TargetType:    r.TargetType,
			TargetID:      r.TargetID,
			Reason:        r.Reason,
			CreatedAt:     r.CreatedAt,
		})
	}
	return items, nil
}

func (r *ModerationRepository) CountLogsByTarget(ctx context.Context, targetType models.ModerationTargetType, targetID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.ModerationLog{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("count moderation logs: %w", err)
	}
	return total, nil
}
