package repository

import (
	"context"
	"fmt"
	"linkup/models"
	"linkup/utils"
	"time"

	"gorm.io/gorm"
)

type CommunityRepository struct {
	db *gorm.DB
}

func NewCommunityRepository(db *gorm.DB) *CommunityRepository {
	return &CommunityRepository{db: db}
}

// CreateWithRoles tạo community, group_member, và user_roles trong 1 transaction.
func (r *CommunityRepository) CreateWithRoles(ctx context.Context, community *models.Community, member *models.GroupMember, userRoles []models.UserRole) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(community).Error; err != nil {
			return fmt.Errorf("lỗi khi lưu thông tin cộng đồng: %w", err)
		}
		if err := tx.Create(member).Error; err != nil {
			return fmt.Errorf("lỗi khi lưu thông tin thành viên: %w", err)
		}
		for i := range userRoles {
			userRoles[i].ID = utils.GenerateUUID()
			userRoles[i].AssignedAt = time.Now().UTC()
			if err := tx.Create(&userRoles[i]).Error; err != nil {
				return fmt.Errorf("lỗi khi gán role cho người tạo: %w", err)
			}
		}
		return nil
	})
}

// AssignUserRole gán một role cho user trong user_roles (scoped role).
func (r *CommunityRepository) AssignUserRole(ctx context.Context, userID string, roleName models.RoleName, scopeID string, scopeType models.ScopeType) error {
	var role models.Role
	if err := r.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		return fmt.Errorf("find role %s: %w", roleName, err)
	}

	userRole := models.NewScopedUserRole(userID, role.ID, scopeID, scopeType)
	userRole.ID = utils.GenerateUUID()
	userRole.AssignedAt = time.Now().UTC()

	if err := r.db.WithContext(ctx).Create(&userRole).Error; err != nil {
		return fmt.Errorf("assign role %s: %w", roleName, err)
	}
	return nil
}

// FindRoleByName tìm role theo tên, trả lỗi nếu không tìm thấy.
func (r *CommunityRepository) FindRoleByName(ctx context.Context, roleName models.RoleName, out *models.Role) error {
	return r.db.WithContext(ctx).Where("name = ?", roleName).First(out).Error
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

// IsUserAdmin kiểm tra user có phải admin của community không (dựa trên user_roles).
func (r *CommunityRepository) IsUserAdmin(ctx context.Context, communityID, userID string) (bool, error) {
	var count int64
	scopeType := models.ScopeTypeCommunity
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND user_roles.scope_id = ? AND user_roles.scope_type = ?",
			userID, communityID, scopeType).
		Where("roles.name IN (?, ?)", models.RoleGroupAdmin, models.RoleCommunityAdmin).
		Count(&count).Error
	return count > 0, err
}

// IsUserMember kiểm tra user có phải member của community không (dựa trên user_roles).
func (r *CommunityRepository) IsUserMember(ctx context.Context, communityID, userID string) (bool, error) {
	var count int64
	scopeType := models.ScopeTypeCommunity
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Where("user_roles.user_id = ? AND user_roles.scope_id = ? AND user_roles.scope_type = ?",
			userID, communityID, scopeType).
		Count(&count).Error
	return count > 0, err
}

// GetUserRole lấy role của user trong community từ user_roles.
func (r *CommunityRepository) GetUserRole(ctx context.Context, communityID, userID string) (models.GroupRole, error) {
	type userRoleWithName struct {
		models.UserRole
		RoleName models.RoleName `gorm:"column:name"`
	}
	var result userRoleWithName
	scopeType := models.ScopeTypeCommunity
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Select("user_roles.*, roles.name").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND user_roles.scope_id = ? AND user_roles.scope_type = ?",
			userID, communityID, scopeType).
		Where("roles.name IN (?, ?, ?, ?)",
			models.RoleGroupAdmin, models.RoleGroupMod, models.RoleGroupMember, models.RoleCommunityAdmin).
		Order("CASE roles.name WHEN 'COMMUNITY_ADMIN' THEN 1 WHEN 'GROUP_ADMIN' THEN 2 WHEN 'GROUP_MOD' THEN 3 ELSE 4 END").
		First(&result).Error
	if err != nil {
		return "", err
	}
	return mapRoleNameToGroupRole(result.RoleName), nil
}

// mapRoleNameToGroupRole ánh xạ role name từ roles table sang GroupRole enum.
func mapRoleNameToGroupRole(name models.RoleName) models.GroupRole {
	switch name {
	case models.RoleCommunityAdmin, models.RoleGroupAdmin:
		return models.GroupRoleAdmin
	case models.RoleGroupMod:
		return models.GroupRoleMod
	default:
		return models.GroupRoleMember
	}
}
