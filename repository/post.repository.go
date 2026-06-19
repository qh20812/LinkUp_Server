package repository

import (
	"context"
	"gorm.io/gorm"
	"linkup/models"
)

type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	FetchActive(ctx context.Context, limit, offset int) ([]models.Post, error)
}

type postRepository struct {
	db *gorm.DB // Thay đổi từ *sql.DB thành *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *models.Post) error {
	// GORM tự động map struct sang câu lệnh INSERT dựa trên cấu trúc model
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *postRepository) FetchActive(ctx context.Context, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.WithContext(ctx).
		Where("status = ?", models.PostStatusActive).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	return posts, err
}
