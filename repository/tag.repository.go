package repository

import (
	"context"
	"linkup/models"
	"strings"

	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

// CreateTagsInTx Lưu danh sách tag vào DB (Hỗ trợ cả chạy mồi và chạy độc lập)
func (r *TagRepository) CreateTagsInTx(ctx context.Context, tx *gorm.DB, tags []models.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	// Khắc phục lỗi Panic: Nếu tx truyền vào bị nil, sử dụng r.db mặc định bên ngoài độc lập
	executor := tx
	if executor == nil {
		executor = r.db
	}

	return executor.WithContext(ctx).Create(&tags).Error
}

// GetPostIDsByHashtag lấy danh sách ID bài viết theo tên tag
func (r *TagRepository) GetPostIDsByHashtag(ctx context.Context, hashtagName string) ([]string, error) {
	var postIDs []string
	err := r.db.WithContext(ctx).
		Model(&models.Tag{}).
		Where("name = ? AND tag_type = ?", strings.ToLower(strings.TrimSpace(hashtagName)), models.TagTypeHashtag).
		Pluck("post_id", &postIDs).Error
	return postIDs, err
}
