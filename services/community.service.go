package services

import (
	"context"
	"errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"time"
)

type CommunityService struct {
	repo       *repository.CommunityRepository
	validation *validations.CommunityValidation
}

func NewCommunityService(repo *repository.CommunityRepository, validation *validations.CommunityValidation) *CommunityService {
	return &CommunityService{repo: repo, validation: validation}
}

func (s *CommunityService) CreateCommunity(ctx context.Context, creatorID, name, description, avatarURI string) (*models.Community, error) {
	if err := s.validation.ValidateCreateCommunity(name, description, avatarURI); err != nil {
		return nil, err
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
