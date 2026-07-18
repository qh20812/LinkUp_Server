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

	GetTotalComments() (int64, error)
	GetTotalMedia() (int64, error)
	GetTotalGroups() (int64, error)
	GetTotalCommunities() (int64, error)
	GetActiveBanCount() (int64, error)

	GetPendingReportCount() (int64, error)
	GetFlaggedMediaCount() (int64, error)
	GetActiveUsersToday() (int64, error)

	GetTotalLikes() (int64, error)
	GetTotalShares() (int64, error)

	GetTopActiveUsers(limit int) ([]dto.TopActiveUser, error)
	GetTopEngagedPosts(limit int) ([]dto.TopEngagedPost, error)
	GetUserStatusDistribution() ([]dto.StatusCount, error)
	GetReportStatusDistribution() ([]dto.StatusCount, error)
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

	err := r.db.Table(tableName).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", startDate+" 00:00:00", endDate+" 23:59:59").
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&points).Error

	return points, err
}

func (r *adminRepository) GetTotalComments() (int64, error) {
	var count int64
	err := r.db.Model(&models.Comment{}).Count(&count).Error
	return count, err
}

func (r *adminRepository) GetTotalMedia() (int64, error) {
	var count int64
	err := r.db.Model(&models.Media{}).Count(&count).Error
	return count, err
}

func (r *adminRepository) GetTotalGroups() (int64, error) {
	var count int64
	err := r.db.Model(&models.Chat{}).
		Where("type = ?", models.ChatTypeGroup).
		Count(&count).Error
	return count, err
}

func (r *adminRepository) GetTotalCommunities() (int64, error) {
	var count int64
	err := r.db.Model(&models.Community{}).Count(&count).Error
	return count, err
}

func (r *adminRepository) GetActiveBanCount() (int64, error) {
	var count int64
	now := time.Now()
	err := r.db.Model(&models.Ban{}).
		Where("expires_at > ? OR expires_at IS NULL", now).
		Count(&count).Error
	return count, err
}

func (r *adminRepository) GetPendingReportCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Report{}).
		Where("status = ?", "pending").
		Count(&count).Error
	return count, err
}

func (r *adminRepository) GetFlaggedMediaCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Media{}).
		Where("status = ?", "flagged").
		Count(&count).Error
	return count, err
}

func (r *adminRepository) GetActiveUsersToday() (int64, error) {
	var count int64
	today := time.Now().UTC().Format("2006-01-02")
	err := r.db.Table("posts").
		Where("DATE(created_at) = ?", today).
		Select("COUNT(DISTINCT user_id)").
		Scan(&count).Error
	return count, err
}

func (r *adminRepository) GetTotalLikes() (int64, error) {
	var count int64
	err := r.db.Model(&models.PostReaction{}).Count(&count).Error
	return count, err
}

func (r *adminRepository) GetTotalShares() (int64, error) {
	var count int64
	err := r.db.Model(&models.PostShare{}).Count(&count).Error
	return count, err
}

func (r *adminRepository) GetTopActiveUsers(limit int) ([]dto.TopActiveUser, error) {
	var users []dto.TopActiveUser
	err := r.db.Table("posts").
		Select("u.id as user_id, u.username, COALESCE(p.display_name, u.username) as display_name, COALESCE(p.avatar_uri, '') as avatar_uri, COUNT(*) as post_count").
		Joins("JOIN users u ON u.id = posts.user_id").
		Joins("LEFT JOIN profiles p ON p.user_id = u.id").
		Where("posts.status != ?", string(models.PostStatusDeleted)).
		Group("posts.user_id, u.id, u.username, p.display_name, p.avatar_uri").
		Order("post_count DESC").
		Limit(limit).
		Scan(&users).Error
	return users, err
}

func (r *adminRepository) GetTopEngagedPosts(limit int) ([]dto.TopEngagedPost, error) {
	var posts []dto.TopEngagedPost
	err := r.db.Raw(`
		SELECT p.id AS post_id, p.title, u.username, p.views_count,
		       (SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.id) AS likes_count,
		       (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id) AS comments_count,
		       EXISTS(SELECT 1 FROM media m WHERE m.post_id = p.id) AS has_media
		FROM posts p
		JOIN users u ON u.id = p.user_id
		WHERE p.status != ?
		ORDER BY (p.views_count +
		          (SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.id) * 2 +
		          (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id) * 3) DESC
		LIMIT ?
	`, string(models.PostStatusDeleted), limit).Scan(&posts).Error
	return posts, err
}

func (r *adminRepository) GetUserStatusDistribution() ([]dto.StatusCount, error) {
	var dist []dto.StatusCount
	err := r.db.Model(&models.User{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&dist).Error
	return dist, err
}

func (r *adminRepository) GetReportStatusDistribution() ([]dto.StatusCount, error) {
	var dist []dto.StatusCount
	err := r.db.Model(&models.Report{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&dist).Error
	return dist, err
}
