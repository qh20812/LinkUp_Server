package repository

import (
	"context"
	"fmt"
	"time"

	"linkup/dto"
	"linkup/models"

	"gorm.io/gorm"
)

type SearchRepository struct {
	db *gorm.DB
}

func NewSearchRepository(db *gorm.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

func (r *SearchRepository) SearchUsers(ctx context.Context, keyword string) ([]dto.UserSearchResult, error) {
	var results []dto.UserSearchResult
	like := "%" + keyword + "%"

	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Select("users.id, users.username, COALESCE(profiles.display_name, '') AS display_name, COALESCE(profiles.avatar_uri, '') AS avatar_uri").
		Joins("LEFT JOIN profiles ON profiles.user_id = users.id").
		Joins("LEFT JOIN user_settings ON user_settings.user_id = users.id").
		Where("users.status = ?", models.UserStatusActive).
		Where("(user_settings.discoverable_in_search IS NULL OR user_settings.discoverable_in_search = ?)", true).
		Where("users.username LIKE ? OR users.email LIKE ? OR profiles.display_name LIKE ?", like, like, like).
		Where("NOT EXISTS (SELECT 1 FROM user_roles JOIN roles ON roles.id = user_roles.role_id WHERE user_roles.user_id = users.id AND roles.name IN (?, ?))", models.RoleSuperAdmin, models.RoleAdmin).
		Limit(10).
		Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	return results, nil
}

func (r *SearchRepository) SearchPosts(ctx context.Context, keyword string) ([]dto.PostSearchResult, error) {
	var results []dto.PostSearchResult
	like := "%" + keyword + "%"

	err := r.db.WithContext(ctx).
		Model(&models.Post{}).
		Select("posts.id, posts.title, posts.user_id, users.username, posts.created_at").
		Joins("LEFT JOIN users ON users.id = posts.user_id").
		Where("posts.status = ?", models.PostStatusPublic).
		Where("posts.title LIKE ?", like).
		Limit(10).
		Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("search posts: %w", err)
	}
	return results, nil
}

func (r *SearchRepository) GetTrendingHashtags(ctx context.Context) ([]dto.HashtagSearchResult, error) {
	type trendingRow struct {
		Name         string
		CurrentCount int64
		PrevCount    int64
	}

	now := time.Now()
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)
	fourteenDaysAgo := now.Add(-14 * 24 * time.Hour)

	var rows []trendingRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			tags.name,
			COUNT(DISTINCT CASE WHEN posts.created_at >= ? THEN posts.id END) AS current_count,
			COUNT(DISTINCT CASE WHEN posts.created_at >= ? AND posts.created_at < ? THEN posts.id END) AS prev_count
		FROM tags
		JOIN posts ON posts.id = tags.post_id
		WHERE tags.tag_type = ?
		  AND posts.status = ?
		  AND posts.created_at >= ?
		GROUP BY tags.name
		HAVING current_count >= 2
		ORDER BY (current_count - COALESCE(prev_count, 0)) * 100 + current_count * 5 DESC
		LIMIT 10
	`, sevenDaysAgo, fourteenDaysAgo, sevenDaysAgo, models.TagTypeHashtag, models.PostStatusPublic, fourteenDaysAgo).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("trending hashtags: %w", err)
	}

	results := make([]dto.HashtagSearchResult, len(rows))
	for i, r := range rows {
		results[i] = dto.HashtagSearchResult{
			Name:      r.Name,
			PostCount: r.CurrentCount,
		}
	}
	return results, nil
}

func (r *SearchRepository) SearchHashtags(ctx context.Context, keyword string) ([]dto.HashtagSearchResult, error) {
	var results []dto.HashtagSearchResult
	like := "%" + keyword + "%"

	err := r.db.WithContext(ctx).
		Model(&models.Tag{}).
		Select("name, COUNT(DISTINCT post_id) AS post_count").
		Where("tag_type = ?", models.TagTypeHashtag).
		Where("name LIKE ?", like).
		Group("name").
		Order("post_count DESC").
		Limit(10).
		Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("search hashtags: %w", err)
	}
	return results, nil
}
