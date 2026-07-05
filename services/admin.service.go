package services

import (
	"context"
	"fmt"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"strings"
)

type AdminService struct {
	authRepo *repository.AuthRepository
}

func NewAdminService(authRepo *repository.AuthRepository) *AdminService {
	return &AdminService{authRepo: authRepo}
}

func (s *AdminService) ListUsers(ctx context.Context, input dto.AdminUserFilterInput) (dto.AdminUserListResponse, error) {
	page := input.Page
	if page <= 0 {
		page = 1
	}

	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	statusFilter := strings.TrimSpace(strings.ToLower(input.Status))
	if statusFilter != "" {
		switch statusFilter {
		case string(models.UserStatusActive), string(models.UserStatusBanned), string(models.UserStatusSuspended):
		default:
			return dto.AdminUserListResponse{}, fmt.Errorf("trạng thái không hợp lệ")
		}
	}

	users, err := s.authRepo.ListUsers(ctx, input.Keyword, statusFilter, pageSize, (page-1)*pageSize)
	if err != nil {
		return dto.AdminUserListResponse{}, err
	}

	total, err := s.authRepo.CountUsers(ctx, input.Keyword, statusFilter)
	if err != nil {
		return dto.AdminUserListResponse{}, err
	}

	resp := dto.AdminUserListResponse{
		Users:    users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	if len(users) == 0 {
		resp.Message = "Không tìm thấy người dùng"
	}
	return resp, nil
}

func (s *AdminService) UpdateUserStatus(ctx context.Context, superAdminID, targetUserID string, input dto.AdminUserUpdateStatusInput) error {
	isSuperAdmin, err := s.authRepo.HasRole(ctx, superAdminID, models.RoleSuperAdmin)
	if err != nil {
		return err
	}
	if !isSuperAdmin {
		return fmt.Errorf("chỉ superadmin mới có quyền")
	}

	targetUser, err := s.authRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return err
	}

	isTargetSuperAdmin, err := s.authRepo.HasRole(ctx, targetUserID, models.RoleSuperAdmin)
	if err != nil {
		return err
	}
	if isTargetSuperAdmin {
		return fmt.Errorf("Không thể chỉnh sửa trạng thái của superadmin")
	}

	statusValue := strings.TrimSpace(strings.ToLower(input.Status))
	var status models.UserStatus
	switch statusValue {
	case string(models.UserStatusActive):
		status = models.UserStatusActive
	case string(models.UserStatusBanned):
		status = models.UserStatusBanned
	case string(models.UserStatusSuspended):
		status = models.UserStatusSuspended
	default:
		return fmt.Errorf("trạng thái không hợp lệ")
	}

	if targetUser.Status == status {
		return nil
	}

	return s.authRepo.UpdateUserStatus(ctx, targetUserID, status)
}
