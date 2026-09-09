package services

import (
	"context"
	"fmt"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"log"
	"net/url"
	"path"
	"strings"
	"time"
)

type ProfileService struct {
	profileRepository *repository.ProfileRepository
	mediaRepo         *repository.MediaRepository
}

func NewProfileService(profileRepository *repository.ProfileRepository) *ProfileService {
	return &ProfileService{
		profileRepository: profileRepository,
	}
}

// SetMediaRepo gán media repository để validate avatar/cover URIs.
func (s *ProfileService) SetMediaRepo(repo *repository.MediaRepository) {
	s.mediaRepo = repo
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
		if err := s.validateMediaURI(ctx, *input.AvatarURI); err != nil {
			return nil, err
		}
		existingProfile.AvatarURI = *input.AvatarURI
	}
	if input.CoverURI != nil {
		if err := s.validateMediaURI(ctx, *input.CoverURI); err != nil {
			return nil, err
		}
		existingProfile.CoverURI = *input.CoverURI
	}
	if input.Bio != nil {
		existingProfile.Bio = *input.Bio
	}
	if input.Location != nil {
		existingProfile.Location = *input.Location
	}
	if input.Work != nil {
		existingProfile.Work = *input.Work
	}
	if input.Education != nil {
		existingProfile.Education = *input.Education
	}
	if input.Website != nil {
		existingProfile.Website = *input.Website
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

// validateMediaURI kiểm tra media URI có tồn tại và không bị reject.
func (s *ProfileService) validateMediaURI(ctx context.Context, uri string) error {
	if s.mediaRepo == nil {
		return nil
	}
	publicID, _, err := parseProfileMediaID(uri)
	if err != nil {
		// Không parse được → bỏ qua validation (client tự chịu trách nhiệm)
		log.Printf("[Profile] không thể parse media URI: %s: %v", uri, err)
		return nil
	}
	media, err := s.mediaRepo.GetByID(ctx, publicID)
	if err != nil {
		// Media không tồn tại trong DB → có thể là URL bên ngoài, bỏ qua
		return nil
	}
	if media.Status == models.MediaStatusRejected {
		return errorsapp.New(errorsapp.ErrCodeMediaRejected)
	}
	return nil
}

// parseProfileMediaID trích xuất public ID từ Cloudinary URL.
func parseProfileMediaID(uri string) (string, string, error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", "", err
	}
	segments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	idx := -1
	for i, seg := range segments {
		if seg == "upload" {
			idx = i
			break
		}
	}
	if idx <= 0 || idx+1 >= len(segments) {
		return "", "", fmt.Errorf("invalid cloudinary path")
	}
	resourceType := segments[idx-1]
	partsAfterUpload := segments[idx+1:]
	if len(partsAfterUpload) == 0 {
		return "", "", fmt.Errorf("missing public id")
	}
	if strings.HasPrefix(partsAfterUpload[0], "v") || strings.HasPrefix(partsAfterUpload[0], "f") || strings.HasPrefix(partsAfterUpload[0], "c") {
		partsAfterUpload = partsAfterUpload[1:]
	}
	if len(partsAfterUpload) == 0 {
		return "", "", fmt.Errorf("missing public id")
	}
	publicID := strings.Join(partsAfterUpload, "/")
	if ext := path.Ext(publicID); ext != "" {
		publicID = strings.TrimSuffix(publicID, ext)
	}
	return publicID, resourceType, nil
}
