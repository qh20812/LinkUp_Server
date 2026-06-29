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

	adminMember := models.NewGroupMember(community.ID, creatorID, models.GroupRoleAdmin)
	adminMember.ID = utils.GenerateUUID()
	adminMember.JoinedAt = now
	adminMember.Points = 500

	if err := s.repo.Create(ctx, &community, &adminMember); err != nil {
		return nil, err
	}
	return &community, nil
}
