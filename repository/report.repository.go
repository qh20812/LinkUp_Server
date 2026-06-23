package repository

import (
	"context"
	"fmt"

	"linkup/models"

	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Create(ctx context.Context, report *models.Report) error {
	tx := r.db.WithContext(ctx).Create(report)
	if tx.Error != nil {
		return fmt.Errorf("insert report: %w", tx.Error)
	}
	return nil
}
