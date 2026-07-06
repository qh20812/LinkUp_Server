package repository

import (
	"context"
	"errors"
	"fmt"

	"linkup/dto"
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

func (r *ReportRepository) FindByID(ctx context.Context, id string) (*models.Report, error) {
	var report models.Report
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&report).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("report not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("find report: %w", err)
	}
	return &report, nil
}

func (r *ReportRepository) UpdateStatus(ctx context.Context, reportID string, status models.ReportStatus) error {
	tx := r.db.WithContext(ctx).Model(&models.Report{}).Where("id = ?", reportID).Update("status", status)
	if tx.Error != nil {
		return fmt.Errorf("update report status: %w", tx.Error)
	}
	return nil
}

func (r *ReportRepository) CountAdminReports(ctx context.Context, keyword, status, targetType string) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Table("reports").Joins("LEFT JOIN users reporter ON reporter.id = reports.reporter_id")

	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("reporter.username LIKE ? OR reporter.email LIKE ? OR reports.reason_detail LIKE ? OR reports.target_user_id LIKE ? OR reports.target_post_id LIKE ? OR reports.target_comment_id LIKE ?", like, like, like, like, like, like)
	}

	if status != "" {
		q = q.Where("reports.status = ?", status)
	}

	if targetType != "" {
		q = q.Where(`(
            (? = 'post' AND reports.target_post_id IS NOT NULL) OR
            (? = 'user' AND reports.target_user_id IS NOT NULL) OR
            (? = 'comment' AND reports.target_comment_id IS NOT NULL)
        )`, targetType, targetType, targetType)
	}

	if err := q.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count reports: %w", err)
	}
	return count, nil
}

func (r *ReportRepository) ListAdminReports(ctx context.Context, keyword, status, targetType, sortBy, order string, limit, offset int) ([]dto.AdminReportListItem, error) {
	var reports []dto.AdminReportListItem

	if order != "asc" && order != "desc" {
		order = "desc"
	}

	orderBy := "reports.created_at desc"
	switch sortBy {
	case "created_at":
		orderBy = "reports.created_at " + order
	case "target_type":
		orderBy = "CASE WHEN reports.target_post_id IS NOT NULL THEN 'post' WHEN reports.target_user_id IS NOT NULL THEN 'user' WHEN reports.target_comment_id IS NOT NULL THEN 'comment' ELSE '' END " + order
	}

	q := r.db.WithContext(ctx).
		Table("reports").
		Select(`reports.id,
                reports.reporter_id,
                reporter.username AS reporter_username,
                reporter.email AS reporter_email,
                reports.target_user_id,
                reports.target_post_id,
                reports.target_comment_id,
                reports.report_type,
                reports.violation_rule_id,
                reports.reason_detail,
                reports.status,
                reports.created_at,
                CASE WHEN reports.target_post_id IS NOT NULL THEN 'post' WHEN reports.target_user_id IS NOT NULL THEN 'user' WHEN reports.target_comment_id IS NOT NULL THEN 'comment' ELSE '' END AS target_type`).
		Joins("LEFT JOIN users reporter ON reporter.id = reports.reporter_id").
		Order(orderBy).
		Limit(limit).
		Offset(offset)

	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("reporter.username LIKE ? OR reporter.email LIKE ? OR reports.reason_detail LIKE ? OR reports.target_user_id LIKE ? OR reports.target_post_id LIKE ? OR reports.target_comment_id LIKE ?", like, like, like, like, like, like)
	}

	if status != "" {
		q = q.Where("reports.status = ?", status)
	}

	if targetType != "" {
		q = q.Where(`(
            (? = 'post' AND reports.target_post_id IS NOT NULL) OR
            (? = 'user' AND reports.target_user_id IS NOT NULL) OR
            (? = 'comment' AND reports.target_comment_id IS NOT NULL)
        )`, targetType, targetType, targetType)
	}

	if err := q.Scan(&reports).Error; err != nil {
		return nil, fmt.Errorf("list admin reports: %w", err)
	}

	return reports, nil
}
