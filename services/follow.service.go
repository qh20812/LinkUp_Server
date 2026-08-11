package services

import (
	"context"
	"fmt"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
)

type FollowService struct {
	followRepository *repository.FollowRepository
	authRepository   *repository.AuthRepository
	notifService     *NotificationService
}

func NewFollowService(followRepository *repository.FollowRepository, authRepository *repository.AuthRepository, notifService *NotificationService) *FollowService {
	return &FollowService{
		followRepository: followRepository,
		authRepository:   authRepository,
		notifService:     notifService,
	}
}

func (s *FollowService) FollowToggle(ctx context.Context, followerID, followingID string) (string, error) {
	if followerID == followingID {
		return "", errorsapp.New(errorsapp.ErrCodeFollowSelf)
	}

	followingUser, err := s.authRepository.FindByID(ctx, followingID)
	if err != nil {
		return "", errorsapp.New(errorsapp.ErrCodeFollowUserNotFound)
	}

	if !followingUser.IsActive() {
		return "", errorsapp.New(errorsapp.ErrCodeFollowUserInactive)
	}

	isSuperAdmin, err := s.authRepository.HasRole(ctx, followingID, models.RoleSuperAdmin)
	if err != nil {
		return "", fmt.Errorf("Kiểm tra vai trò: %w", err)
	}
	if isSuperAdmin {
		return "", errorsapp.New(errorsapp.ErrCodeFollowSuperAdmin)
	}

	isAdmin, err := s.authRepository.HasRole(ctx, followingID, models.RoleAdmin)
	if err != nil {
		return "", fmt.Errorf("kiểm tra vai trò admin: %w", err)
	}
	if isAdmin {
		return "", errorsapp.New(errorsapp.ErrCodeFollowAdminRestricted)
	}

	isFollowing, err := s.followRepository.IsFollowing(ctx, followerID, followingID)
	if err != nil {
		return "", fmt.Errorf("kiểm tra trạng thái follow: %w", err)
	}

	action := ""
	if isFollowing {
		err = s.followRepository.Delete(ctx, followerID, followingID)
		if err != nil {
			return "", err
		}
		action = "unfollowed"
	} else {
		follow := models.Follow{
			ID:          utils.GenerateUUID(),
			FollowerID:  followerID,
			FollowingID: followingID,
		}
		err = s.followRepository.Create(ctx, &follow)
		if err != nil {
			return "", err
		}
		action = "followed"
		if followingID != followerID {
			s.notifService.Create(ctx, followingID, &followerID, models.NotificationTypeFollow, "đã theo dõi bạn", nil, &followerID, nil)
		}
	}

	return action, nil
}

func (s *FollowService) GetFollowerStats(ctx context.Context, userID string, viewerID *string) (map[string]interface{}, error) {
	followerCount, err := s.followRepository.GetFollowerCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	followingCount, err := s.followRepository.GetFollowingCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"follower_count":  followerCount,
		"following_count": followingCount,
	}

	if viewerID != nil && *viewerID != "" {
		isFollowing, err := s.followRepository.IsFollowing(ctx, *viewerID, userID)
		if err == nil {
			result["is_following"] = isFollowing
		}
	}

	return result, nil
}

func (s *FollowService) GetSuggestions(ctx context.Context, userID string, page, pageSize int) (dto.FollowSuggestionsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 20 {
		pageSize = 5
	}

	items, total, err := s.followRepository.GetSuggestions(ctx, userID, page, pageSize)
	if err != nil {
		return dto.FollowSuggestionsResponse{}, fmt.Errorf("get suggestions: %w", err)
	}

	return dto.FollowSuggestionsResponse{
		Data:     items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasMore:  int64(page*pageSize) < total,
	}, nil
}
