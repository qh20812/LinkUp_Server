package services

import (
	"context"
	"fmt"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"time"
)

type ProfileService struct {
	profileRepository *repository.ProfileRepository
}

func NewProfileService(profileRepository *repository.ProfileRepository) *ProfileService {
	return &ProfileService{
		profileRepository: profileRepository,
	}
}

func (s *ProfileService) ViewProfile(ctx context.Context, userID string) (*repository.ProfileView, error) {
	profile, err := s.profileRepository.FindEnrichedByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("view profile: %w", err)
	}
	if profile == nil {
		return nil, errorsapp.New(errorsapp.ErrCodeProfileNotFound)
	}
	return profile, nil
}

func (s *ProfileService) ViewProfileByID(ctx context.Context, viewerID string, targetUserID string) (*repository.ProfileView, error) {
	profile, err := s.profileRepository.FindEnrichedByUserID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("view profile: %w", err)
	}
	if profile == nil {
		return nil, errorsapp.New(errorsapp.ErrCodeProfileNotFound)
	}

	if viewerID == targetUserID {
		return profile, nil
	}

	if profile.IsPrivateProfile {
		return nil, errorsapp.New(errorsapp.ErrCodeProfilePrivate)
	}

	return profile, nil
}

func (s *ProfileService) EditProfile(ctx context.Context, userID string, input dto.EditProfileInput) (*models.Profile, error) {
	existingProfile, err := s.profileRepository.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("edit profile: %w", err)
	}
	if existingProfile == nil {
		return nil, errorsapp.New(errorsapp.ErrCodeProfileNotFound)
	}

	if input.PhoneNumber != nil && *input.PhoneNumber != "" {
		existingPhone, err := s.profileRepository.FindByPhoneNumber(ctx, *input.PhoneNumber, userID)
		if err != nil {
			return nil, fmt.Errorf("edit profile: %w", err)
		}
		if existingPhone != nil {
			return nil, errorsapp.New(errorsapp.ErrCodePhoneExists)
		}
	}

	if input.DisplayName != nil && *input.DisplayName != "" {
		existingProfile.DisplayName = *input.DisplayName
	}
	if input.PhoneNumber != nil {
		existingProfile.PhoneNumber = *input.PhoneNumber
	}
	if input.DateOfBirth != nil {
		existingProfile.DateOfBirth = input.DateOfBirth
	}
	if input.AvatarURI != nil {
		existingProfile.AvatarURI = *input.AvatarURI
	}
	if input.CoverURI != nil {
		existingProfile.CoverURI = *input.CoverURI
	}
	if input.Bio != nil {
		existingProfile.Bio = *input.Bio
	}
	if input.IsPrivateProfile != nil {
		existingProfile.IsPrivateProfile = *input.IsPrivateProfile
	}
	if input.IsPrivatePosts != nil {
		existingProfile.IsPrivatePosts = *input.IsPrivatePosts
	}
	if input.AllowStrangerFriendRequest != nil {
		existingProfile.AllowStrangerFriendRequest = *input.AllowStrangerFriendRequest
	}

	now := time.Now().Truncate(0)
	existingProfile.UpdatedAt = &now

	result, err := s.profileRepository.Update(ctx, userID, existingProfile)
	if err != nil {
		return nil, fmt.Errorf("edit profile: %w", err)
	}

	return result, nil
}
