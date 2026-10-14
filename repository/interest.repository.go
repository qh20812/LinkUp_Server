package repository

import (
	"context"
	"strings"
	"time"

	"linkup/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InterestRepository struct {
	db *gorm.DB
}

func NewInterestRepository(db *gorm.DB) *InterestRepository {
	return &InterestRepository{db: db}
}

// AddInterestScore cộng điểm interest cho (user, tag), tạo mới nếu chưa có.
// Tag được chuẩn hóa lowercase để khớp với lookup hashtag.
func (r *InterestRepository) AddInterestScore(ctx context.Context, userID, tag string, delta float64) error {
	tag = strings.ToLower(strings.TrimSpace(tag))
	if tag == "" || delta == 0 {
		return nil
	}
	interest := models.UserInterest{
		UserID:    userID,
		Tag:       tag,
		Score:     delta,
		UpdatedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "tag"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"score":      gorm.Expr("score + ?", delta),
			"updated_at": time.Now(),
		}),
	}).Create(&interest).Error
}

// GetTopInterests trả về các tag có điểm cao nhất của user (đã decay lười
// theo thời gian từ lần cập nhật cuối).
func (r *InterestRepository) GetTopInterests(ctx context.Context, userID string, limit int, decayRate float64) ([]models.UserInterest, error) {
	if limit < 1 {
		limit = 20
	}
	var interests []models.UserInterest
	err := r.db.WithContext(ctx).
		Model(&models.UserInterest{}).
		Select("user_id, tag, (score * EXP(-? * ((UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(updated_at)) / 86400.0))) AS score, updated_at", decayRate).
		Where("user_id = ?", userID).
		Order("score DESC").
		Limit(limit).
		Scan(&interests).Error
	return interests, err
}
