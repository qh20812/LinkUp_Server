package repository

import (
	"context"
	"fmt"

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
		Where("users.status = ?", models.UserStatusActive).
		Where("users.username LIKE ? OR users.email LIKE ? OR profiles.display_name LIKE ?", like, like, like).
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
		Where("posts.status = ?", models.PostStatusActive).
		Where("posts.title LIKE ?", like).
		Limit(10).
		Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("search posts: %w", err)
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
