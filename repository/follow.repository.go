package repository

import (
	"context"
	"fmt"
	"linkup/dto"
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

func (r *FollowRepository) GetSuggestions(ctx context.Context, userID string, page, pageSize int) ([]dto.FollowSuggestionItem, int64, error) {
	eligibleWhere := `u.id != ? AND u.status = ?
		AND NOT EXISTS (SELECT 1 FROM follows WHERE follower_id = ? AND following_id = u.id)
		AND NOT EXISTS (SELECT 1 FROM blocks WHERE user_id = ? AND blocked_user_id = u.id)
		AND NOT EXISTS (SELECT 1 FROM blocks WHERE user_id = u.id AND blocked_user_id = ?)
		AND NOT EXISTS (SELECT 1 FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id = u.id AND r.name IN (?, ?))`

	eligibleArgs := func() []interface{} {
		return []interface{}{userID, models.UserStatusActive, userID, userID, userID, models.RoleSuperAdmin, models.RoleAdmin}
	}

	var total int64
	if err := r.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM users u WHERE `+eligibleWhere, eligibleArgs()...).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count suggestions: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	offset := (page - 1) * pageSize

	type suggestionRow struct {
		ID          string
		Username    string
		DisplayName string
		AvatarURI   string
		MutualCount int
	}

	fetch := func(query string, args ...interface{}) ([]dto.FollowSuggestionItem, error) {
		var rows []suggestionRow
		if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
			return nil, fmt.Errorf("get suggestions: %w", err)
		}
		items := make([]dto.FollowSuggestionItem, len(rows))
		for i, row := range rows {
			items[i] = dto.FollowSuggestionItem{
				ID:          row.ID,
				Username:    row.Username,
				DisplayName: row.DisplayName,
				AvatarURI:   row.AvatarURI,
				MutualCount: row.MutualCount,
			}
		}
		return items, nil
	}

	personalizedQuery := `SELECT u.id, u.username,
		COALESCE(p.display_name, '') AS display_name,
		COALESCE(p.avatar_uri, '') AS avatar_uri,
		COUNT(DISTINCT f1.follower_id) AS mutual_count
	FROM users u
	LEFT JOIN profiles p ON p.user_id = u.id
	LEFT JOIN follows f1 ON f1.following_id = u.id
		AND f1.follower_id IN (SELECT following_id FROM follows WHERE follower_id = ?)
	LEFT JOIN follows f_existing ON f_existing.following_id = u.id AND f_existing.follower_id = ?
	LEFT JOIN blocks b1 ON b1.user_id = ? AND b1.blocked_user_id = u.id
	LEFT JOIN blocks b2 ON b2.user_id = u.id AND b2.blocked_user_id = ?
	WHERE ` + eligibleWhere + `
	GROUP BY u.id
	ORDER BY mutual_count DESC, u.id
	LIMIT ? OFFSET ?`

	// personalized: f1_subq, f_existing, b1_user, b2_target, b2_user, then eligibleArgs, then pageSize, offset
	perArgs := []interface{}{userID, userID, userID, userID}
	perArgs = append(perArgs, eligibleArgs()...)
	perArgs = append(perArgs, pageSize, offset)

	items, err := fetch(personalizedQuery, perArgs...)
	if err != nil {
		return nil, 0, err
	}
	if len(items) == 0 {
		return []dto.FollowSuggestionItem{}, 0, nil
	}

	return items, total, nil
}
