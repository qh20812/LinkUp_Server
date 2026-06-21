package repository

import (
	"context"
	"gorm.io/gorm"
	"linkup/models"
)

type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	FetchActive(ctx context.Context, limit, offset int) ([]models.Post, error)
	FindByID(ctx context.Context, id string) (*models.Post, error)
	IncrementViewsCount(ctx context.Context, id string) error
	// --- Các hàm xử lý Reaction ---
	FindReaction(ctx context.Context, userID, postID, emojiID string) (*models.PostReaction, error)
	CreateReaction(ctx context.Context, reaction models.PostReaction) error
	DeleteReaction(ctx context.Context, id string) error
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *models.Post) error {
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

func (r *postRepository) FindByID(ctx context.Context, id string) (*models.Post, error) {
	var post models.Post
	// Kiểm tra cả trạng thái Active để bảo vệ dữ liệu bị ẩn/xóa
	err := r.db.WithContext(ctx).Where("id = ? AND status = ?", id, models.PostStatusActive).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) IncrementViewsCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&models.Post{}).Where("id = ?", id).
		Update("views_count", gorm.Expr("views_count + ?", 1)).Error
}

func (r *postRepository) CreateReaction(ctx context.Context, reaction models.PostReaction) error {
	return r.db.WithContext(ctx).Create(&reaction).Error
}


func (r *postRepository) FindReaction(ctx context.Context, userID, postID, emojiID string) (*models.PostReaction, error) {
	var reaction models.PostReaction
	err := r.db.WithContext(ctx).Where("user_id = ? AND post_id = ? AND emoji_id = ?", userID, postID, emojiID).First(&reaction).Error
	if err != nil {
		return nil, err // Trả về lỗi (ví dụ: gorm.ErrRecordNotFound) nếu chưa từng like
	}
	return &reaction, nil
}

func (r *postRepository) DeleteReaction(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.PostReaction{}, "id = ?", id).Error
}