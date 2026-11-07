package repository

import (
	"context"

	"linkup/models"

	"gorm.io/gorm"
)

// AdminAuditRepository ghi nhật ký hành động quản trị.
type AdminAuditRepository struct {
	db *gorm.DB
}

func NewAdminAuditRepository(db *gorm.DB) *AdminAuditRepository {
	return &AdminAuditRepository{db: db}
}

func (r *AdminAuditRepository) Create(ctx context.Context, log *models.AdminAuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
