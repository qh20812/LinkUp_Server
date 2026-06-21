package services

import (
	"context"
	"fmt"
	"linkup/models"
	"linkup/repository"
)

type ProfileService struct {
	profileRepository *repository.ProfileRepository
}

func NewProfileService(profileRepository *repository.ProfileRepository) *ProfileService {
	return &ProfileService{
		profileRepository: profileRepository,
	}
}

func (s *ProfileService) ViewProfile(ctx context.Context, userID string) (*models.Profile, error) {
	profile, err := s.profileRepository.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("view profile: %w", err)
	}
	return profile, nil
}
