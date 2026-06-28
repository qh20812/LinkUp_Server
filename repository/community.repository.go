package repository

import (
	"context"
	"fmt"
	"linkup/models"

	"gorm.io/gorm"
)

type CommunityRepository struct {
	db *gorm.DB
}

func NewCommunityRepository(db *gorm.DB) *CommunityRepository {
	return &CommunityRepository{db: db}
}

func (r *CommunityRepository) Create(ctx context.Context, community *models.Community, member *models.GroupMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(community).Error; err != nil {
			return fmt.Errorf("lỗi khi lưu thông tin cộng đồng: %w", err)
		}
		if err := tx.Create(member).Error; err != nil {
			return fmt.Errorf("lỗi khi lưu thông tin thành viên: %w", err)
		}
		return nil
	})
}

func (r *CommunityRepository) FindByID(ctx context.Context, id string) (*models.Community, error) {
	var community models.Community
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&community).Error
	if err != nil {
		return nil, err
	}
	return &community, nil
}

func (r *CommunityRepository) FindByName(ctx context.Context, name string) (*models.Community, error) {
	var community models.Community
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&community).Error
	if err != nil {
		return nil, err
	}
	return &community, nil
}

func (r *CommunityRepository) IsNameTaken(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Community{}).
		Where("name = ?", name).
		Count(&count).Error
	return count > 0, err
}
