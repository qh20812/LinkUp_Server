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

// ViewProfileByID - Người khác xem profile (kiểm tra private)
func (s *ProfileService) ViewProfileByID(ctx context.Context, viewerID string, targetUserID string) (*models.Profile, error) {
    profile, err := s.profileRepository.FindByUserID(ctx, targetUserID)
    if err != nil {
        return nil, fmt.Errorf("profile not found")
    }

    if viewerID == targetUserID {
        return profile, nil
    }

    if profile.IsPrivateProfile {
        return nil, fmt.Errorf("this profile is private")
    }

    return profile, nil
}

func (s *ProfileService) EditProfile(ctx context.Context, userID string, input dto.EditProfileInput) (*models.Profile, error) {
    existingProfile, err := s.profileRepository.FindByUserID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("edit profile: %w", err)
    }

    now := time.Now().Truncate(0)

    updateProfile := models.Profile{
        ID:                         existingProfile.ID,
        UserID:                     existingProfile.UserID,
        DisplayName:                input.DisplayName,
        PhoneNumber:                input.PhoneNumber,
        DateOfBirth:                input.DateOfBirth,
        AvatarURI:                  input.AvatarURI,
        Bio:                        input.Bio,
        IsPrivateProfile:           input.IsPrivateProfile,
        IsPrivatePosts:             input.IsPrivatePosts,
        AllowStrangerFriendRequest: input.AllowStrangerFriendRequest,
        UpdatedAt:                  &now,
    }

    result, err := s.profileRepository.Update(ctx, userID, &updateProfile)
    if err != nil {
        return nil, fmt.Errorf("edit profile: %w", err)
    }

    return result, nil
}