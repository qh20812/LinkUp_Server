package services

import (
	"context"
	"errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"mime/multipart"
	"time"
)

type CommunityService struct {
	repo       *repository.CommunityRepository
	authRepo   *repository.AuthRepository
	mediaService MediaService
	groupRole  *utils.GroupRoleChecker
	validation *validations.CommunityValidation
}

func NewCommunityService(repo *repository.CommunityRepository, validation *validations.CommunityValidation, authRepo *repository.AuthRepository, mediaService MediaService) *CommunityService {
	return &CommunityService{
		repo:         repo,
		validation:   validation,
		authRepo:     authRepo,
		mediaService: mediaService,
		groupRole:    utils.NewGroupRoleChecker(repo.GetUserRole),
	}
}

func (s *CommunityService) CreateCommunity(ctx context.Context, creatorID, name, description, avatarURI string) (*models.Community, error) {
	if err := s.validation.ValidateCreateCommunity(name, description, avatarURI); err != nil {
		return nil, err
	}

	isAdmin, err := s.authRepo.HasRole(ctx, creatorID, models.RoleAdmin)
	if err != nil {
		return nil, errors.New("lỗi khi kiểm tra quyền người dùng")
	}
	isSuperAdmin, err := s.authRepo.HasRole(ctx, creatorID, models.RoleSuperAdmin)
	if err != nil {
		return nil, errors.New("lỗi khi kiểm tra quyền người dùng")
	}
	if isAdmin || isSuperAdmin {
		return nil, errors.New("quản trị viên không được tạo cộng đồng")
	}

	creator, err := s.authRepo.FindByID(ctx, creatorID)
	if err != nil {
		return nil, errors.New("người dùng không tồn tại")
	}
	if !creator.IsActive() {
		return nil, errors.New("tài khoản chưa được kích hoạt, không thể tạo cộng đồng")
	}

	taken, err := s.repo.IsNameTaken(ctx, name)
	if err != nil {
		return nil, errors.New("lỗi khi kiểm tra tên cộng đồng")
	}
	if taken {
		return nil, validations.ErrCommunityNameExists
	}

	now := time.Now().UTC()
	community := models.NewCommunity(creatorID, name, description, avatarURI)
	community.ID = utils.GenerateUUID()
	community.CreatedAt = now

	adminMember := models.NewGroupMember(community.ID, creatorID)
	adminMember.ID = utils.GenerateUUID()
	adminMember.JoinedAt = now
	adminMember.Points = 500

	scopeType := models.ScopeTypeCommunity
	userRoles := []models.UserRole{
		models.NewScopedUserRole(creatorID, "", community.ID, scopeType),
		models.NewScopedUserRole(creatorID, "", community.ID, scopeType),
	}

	var communityAdminRole, groupAdminRole models.Role
	if err := s.repo.FindRoleByName(ctx, models.RoleCommunityAdmin, &communityAdminRole); err != nil {
		return nil, err
	}
	if err := s.repo.FindRoleByName(ctx, models.RoleGroupAdmin, &groupAdminRole); err != nil {
		return nil, err
	}
	userRoles[0].RoleID = communityAdminRole.ID
	userRoles[1].RoleID = groupAdminRole.ID

	if err := s.repo.CreateWithRoles(ctx, &community, &adminMember, userRoles); err != nil {
		return nil, err
	}
	return &community, nil
}

func (s *CommunityService) SetCommunityBackground(ctx context.Context, userID, communityID string, file *multipart.FileHeader) error {
	if err := s.groupRole.RequireRole(ctx, communityID, userID, models.GroupRoleAdmin); err != nil {
		return err
	}

	community, err := s.repo.FindByID(ctx, communityID)
	if err != nil {
		return errors.New("cộng đồng không tồn tại")
	}

	user, err := s.authRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("người dùng không tồn tại")
	}
	if !user.IsActive() {
		return errors.New("tài khoản chưa được kích hoạt")
	}

	media, err := s.mediaService.UploadMedia(ctx, userID, file)
	if err != nil {
		return errors.New("tải ảnh background thất bại")
	}

	if err := s.validation.ValidateBackgroundURI(media.FileURI); err != nil {
		return err
	}

	if err := s.repo.UpdateBackground(ctx, community.ID, media.FileURI); err != nil {
		return errors.New("cập nhật background cộng đồng thất bại")
	}

	return nil
}
