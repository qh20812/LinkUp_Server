package services

import (
    "context"
    "fmt"
    "linkup/dto"
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

func (s *ProfileService) ViewProfile(ctx context.Context, userID string) (*models.Profile, error) {
    profile, err := s.profileRepository.FindByUserID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("view profile: %w", err)
    }
    return profile, nil
}

func (s *ProfileService) ViewProfileByID(ctx context.Context, viewerID string, targetUserID string) (*models.Profile, error) {
    profile, err := s.profileRepository.FindByUserID(ctx, targetUserID)
    if err != nil {
		return nil, fmt.Errorf("không tìm thấy hồ sơ")
    }

    if viewerID == targetUserID {
        return profile, nil
    }

    if profile.IsPrivateProfile {
		return nil, fmt.Errorf("hồ sơ này ở chế độ riêng tư")
    }

    return profile, nil
}

func (s *ProfileService) EditProfile(ctx context.Context, userID string, input dto.EditProfileInput) (*models.Profile, error) {
    existingProfile, err := s.profileRepository.FindByUserID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("edit profile: %w", err)
    }

    if input.PhoneNumber != nil && *input.PhoneNumber != "" {
        existingPhone, err := s.profileRepository.FindByPhoneNumber(ctx, *input.PhoneNumber, userID)
        if err != nil {
            return nil, fmt.Errorf("edit profile: %w", err)
        }
        if existingPhone != nil {
		return nil, fmt.Errorf("số điện thoại đã tồn tại")
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