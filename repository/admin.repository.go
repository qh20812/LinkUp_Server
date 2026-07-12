package repository

import (
	"time"
	"linkup/dto"
	"linkup/models"

	"gorm.io/gorm"
)

type AdminRepository interface {
	GetTotalUsers() (int64, error)
	GetTotalPosts() (int64, error)
	GetTotalReports() (int64, error)
	GetCountBeforeDate(tableName string, date time.Time) (int64, error)
	GetChartData(tableName string, startDate, endDate string) ([]dto.ChartDataPoint, error)
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetTotalUsers() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

func (r *adminRepository) GetTotalPosts() (int64, error) {
	var count int64
	// Chỉ đếm những bài viết hoạt động (không tính các bài trạng thái deleted)
	err := r.db.Model(&models.Post{}).Where("status != ?", string(models.PostStatusDeleted)).Count(&count).Error
	return count, err
}

func (r *adminRepository) GetTotalReports() (int64, error) {
	var count int64
	err := r.db.Model(&models.Report{}).Count(&count).Error
	return count, err
}

func (r *adminRepository) GetCountBeforeDate(tableName string, date time.Time) (int64, error) {
	var count int64
	err := r.db.Table(tableName).
		Where("created_at < ?", date).
		Count(&count).Error
	return count, err
}

func (r *adminRepository) GetChartData(tableName string, startDate, endDate string) ([]dto.ChartDataPoint, error) {
	var points []dto.ChartDataPoint

	// Sử dụng hàm DATE() của MySQL để nhóm dữ liệu tạo mới chính xác theo từng ngày
	err := r.db.Table(tableName).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", startDate+" 00:00:00", endDate+" 23:59:59").
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&points).Error

	return points, err
}
