package repository

import (
	"errors"
	"fmt"
	"linkup/models"
	"linkup/utils"
	"time"

	"gorm.io/gorm"
)

type PackageRepository interface {
	GetAllPackages() ([]models.AdPackage, error)
	GetPackageByID(id string) (*models.AdPackage, error)
	GetActiveSubscription(userID string) (*models.PartnerSubscription, error)
	CreateSubscription(sub *models.PartnerSubscription) error
	UpdateSubscription(sub *models.PartnerSubscription) error
	IncrementSlotsUsed(subID string) error
	DecrementSlotsUsed(subID string) error
	UpdateUserRole(userID string, roleName models.RoleName) error
	SubscribeWithTransaction(activeSub *models.PartnerSubscription, newSub *models.PartnerSubscription) error
}

type packageRepositoryImpl struct {
	db *gorm.DB
}

func NewPackageRepository(db *gorm.DB) PackageRepository {
	return &packageRepositoryImpl{db: db}
}

func (r *packageRepositoryImpl) GetAllPackages() ([]models.AdPackage, error) {
	var packages []models.AdPackage
	err := r.db.Order("sort_order ASC").Find(&packages).Error
	return packages, err
}

func (r *packageRepositoryImpl) GetPackageByID(id string) (*models.AdPackage, error) {
	var pkg models.AdPackage
	err := r.db.First(&pkg, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("không tìm thấy gói quảng cáo")
		}
		return nil, err
	}
	return &pkg, nil
}

func (r *packageRepositoryImpl) GetActiveSubscription(userID string) (*models.PartnerSubscription, error) {
	var sub models.PartnerSubscription
	err := r.db.Preload("Package").
		Where("user_id = ? AND status = ? AND expires_at > ?", userID, models.SubscriptionStatusActive, time.Now()).
		First(&sub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *packageRepositoryImpl) CreateSubscription(sub *models.PartnerSubscription) error {
	return r.db.Create(sub).Error
}

func (r *packageRepositoryImpl) UpdateSubscription(sub *models.PartnerSubscription) error {
	return r.db.Save(sub).Error
}

func (r *packageRepositoryImpl) IncrementSlotsUsed(subID string) error {
	return r.db.Model(&models.PartnerSubscription{}).
		Where("id = ?", subID).
		UpdateColumn("slots_used", gorm.Expr("slots_used + 1")).Error
}

func (r *packageRepositoryImpl) DecrementSlotsUsed(subID string) error {
	return r.db.Model(&models.PartnerSubscription{}).
		Where("id = ? AND slots_used > 0", subID).
		UpdateColumn("slots_used", gorm.Expr("slots_used - 1")).Error
}

// UpdateUserRole cập nhật hoặc gán platform role mới cho user (ScopeID = NULL)
func (r *packageRepositoryImpl) UpdateUserRole(userID string, roleName models.RoleName) error {
	var role models.Role
	if err := r.db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return fmt.Errorf("không tìm thấy role %s: %w", roleName, err)
	}

	var userRole models.UserRole
	err := r.db.Where("user_id = ? AND scope_id IS NULL", userID).First(&userRole).Error

	if err == nil {
		return r.db.Model(&userRole).Update("role_id", role.ID).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		newUserRole := models.UserRole{
			ID:         utils.GenerateUUID(),
			UserID:     userID,
			RoleID:     role.ID,
			ScopeID:    nil,
			ScopeType:  nil,
			AssignedAt: time.Now(),
		}
		return r.db.Create(&newUserRole).Error
	}

	return err
}

// SubscribeWithTransaction thực hiện toàn bộ thao tác đổi gói + cập nhật role PARTNER trong 1 Transaction
func (r *packageRepositoryImpl) SubscribeWithTransaction(activeSub *models.PartnerSubscription, newSub *models.PartnerSubscription) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Hủy gói cước cũ nếu có
		if activeSub != nil {
			activeSub.Status = models.SubscriptionStatusCancelled
			if err := tx.Save(activeSub).Error; err != nil {
				return err
			}
		}

		// 2. Tạo bản ghi gói cước mới
		if err := tx.Create(newSub).Error; err != nil {
			return err
		}

		// 3. Lấy ID của role PARTNER từ bảng roles
		var partnerRole models.Role
		if err := tx.Where("name = ?", models.RolePartner).First(&partnerRole).Error; err != nil {
			return fmt.Errorf("không tìm thấy role PARTNER trong DB: %w", err)
		}

		// 4. Cập nhật hoặc chèn bản ghi vào bảng user_roles cho Platform-level (ScopeID = NULL)
		var userRole models.UserRole
		err := tx.Where("user_id = ? AND scope_id IS NULL", newSub.UserID).First(&userRole).Error

		if err == nil {
			// Đã có platform role -> Cập nhật sang PARTNER
			if err := tx.Model(&userRole).Update("role_id", partnerRole.ID).Error; err != nil {
				return err
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// Chưa có -> Tạo mới bản ghi UserRole
			newUserRole := models.UserRole{
				ID:         utils.GenerateUUID(),
				UserID:     newSub.UserID,
				RoleID:     partnerRole.ID,
				ScopeID:    nil,
				ScopeType:  nil,
				AssignedAt: time.Now(),
			}
			if err := tx.Create(&newUserRole).Error; err != nil {
				return err
			}
		} else {
			return err
		}

		return nil
	})
}
