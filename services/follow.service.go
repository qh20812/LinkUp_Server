package services

import (
	"context"
	"fmt"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
)

type FollowService struct {
	followRepository *repository.FollowRepository
	authRepository   *repository.AuthRepository
}

func NewFollowService(followRepository *repository.FollowRepository, authRepository *repository.AuthRepository) *FollowService {
	return &FollowService{
		followRepository: followRepository,
		authRepository:   authRepository,
	}
}

func (s *FollowService) FollowToggle(ctx context.Context, followerID, followingID string) (string, error) {
	if followerID == followingID {
		return "", fmt.Errorf("không thể follow chính mình")
	}

	followingUser, err := s.authRepository.FindByID(ctx, followingID)
	if err != nil {
		return "", fmt.Errorf("người dùng không tồn tại")
	}

	if !followingUser.IsActive() {
		return "", fmt.Errorf("không thể follow người dùng này")
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

