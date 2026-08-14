package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
)

type FriendService struct {
	friendRepo  *repository.FriendRepository
	authRepo    *repository.AuthRepository
	profileRepo *repository.ProfileRepository
	validation  *validations.FriendValidation
	notifService *NotificationService
}

func NewFriendService(friendRepo *repository.FriendRepository, authRepo *repository.AuthRepository, profileRepo *repository.ProfileRepository, validation *validations.FriendValidation, notifService *NotificationService) *FriendService {
	return &FriendService{
		friendRepo:   friendRepo,
		authRepo:     authRepo,
		profileRepo:  profileRepo,
		validation:   validation,
		notifService: notifService,
	}
}

func (s *FriendService) ToggleFriendRequest(ctx context.Context, userID, targetUserID string) (dto.FriendRequestResponse, error) {
	if err := s.validation.ValidateToggleFriendRequest(userID, targetUserID); err != nil {
		return dto.FriendRequestResponse{}, err
	}

	targetUser, err := s.authRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return dto.FriendRequestResponse{}, errorsapp.New(errorsapp.ErrCodeFriendUserNotFound)
	}

	if !targetUser.IsActive() {
		return dto.FriendRequestResponse{}, errorsapp.New(errorsapp.ErrCodeFriendUserInactive)
	}

	isAdmin, err := s.authRepo.HasRole(ctx, targetUserID, models.RoleAdmin)
	if err != nil {
		return dto.FriendRequestResponse{}, fmt.Errorf("toggle friend request: %w", err)
	}
	isSuperAdmin, err := s.authRepo.HasRole(ctx, targetUserID, models.RoleSuperAdmin)
	if err != nil {
		return dto.FriendRequestResponse{}, fmt.Errorf("toggle friend request: %w", err)
	}
	if isAdmin || isSuperAdmin {
		return dto.FriendRequestResponse{}, errorsapp.New(errorsapp.ErrCodeFriendAdminRestricted)
	}

	existing, err := s.friendRepo.FindBySenderAndReceiver(ctx, userID, targetUserID)
	if err != nil {
		return dto.FriendRequestResponse{}, fmt.Errorf("toggle friend request: %w", err)
	}

	if existing != nil {
		if existing.Status != models.FriendStatusPending {
			return dto.FriendRequestResponse{}, errorsapp.New(errorsapp.ErrCodeFriendRequestHandled)
		}
		if err := s.friendRepo.Delete(ctx, existing.ID); err != nil {
			return dto.FriendRequestResponse{}, fmt.Errorf("toggle friend request: %w", err)
		}
		return dto.FriendRequestResponse{
			Status:  "revoked",
			Message: "Đã thu hồi lời mời kết bạn thành công",
		}, nil
	}

	friend := models.Friend{
		ID:         utils.GenerateUUID(),
		SenderID:   userID,
		ReceiverID: targetUserID,
		Status:     models.FriendStatusPending,
		CreatedAt:  time.Now().UTC(),
	}
	if err := s.friendRepo.Create(ctx, &friend); err != nil {
		return dto.FriendRequestResponse{}, fmt.Errorf("toggle friend request: %w", err)
	}

	s.notifService.Create(ctx, targetUserID, &userID, models.NotificationTypeFriendRequest, "đã gửi lời mời kết bạn", nil, &userID, nil)

	return dto.FriendRequestResponse{
		Status:  "sent",
		Message: "Đã gửi lời mời kết bạn thành công",
	}, nil
}

func (s *FriendService) GetFriendRequests(ctx context.Context, userID string) (dto.FriendRequestListResponse, error) {
	sent, err := s.friendRepo.FindBySenderID(ctx, userID, models.FriendStatusPending)
	if err != nil {
		return dto.FriendRequestListResponse{}, fmt.Errorf("get sent requests: %w", err)
	}

	received, err := s.friendRepo.FindByReceiverID(ctx, userID, models.FriendStatusPending)
	if err != nil {
		return dto.FriendRequestListResponse{}, fmt.Errorf("get received requests: %w", err)
	}

	sentItems := make([]dto.FriendRequestItem, 0, len(sent))
	for _, f := range sent {
		profile, err := s.profileRepo.FindByUserID(ctx, f.ReceiverID)
		if err != nil || profile == nil {
			continue
		}
		sentItems = append(sentItems, dto.FriendRequestItem{
			ID:          f.ID,
			UserID:      f.ReceiverID,
			DisplayName: profile.DisplayName,
			AvatarURI:   profile.AvatarURI,
			Status:      string(f.Status),
			CreatedAt:   f.CreatedAt,
			Direction:   "sent",
		})
	}

	receivedItems := make([]dto.FriendRequestItem, 0, len(received))
	for _, f := range received {
		profile, err := s.profileRepo.FindByUserID(ctx, f.SenderID)
		if err != nil || profile == nil {
			continue
		}
		receivedItems = append(receivedItems, dto.FriendRequestItem{
			ID:          f.ID,
			UserID:      f.SenderID,
			DisplayName: profile.DisplayName,
			AvatarURI:   profile.AvatarURI,
			Status:      string(f.Status),
			CreatedAt:   f.CreatedAt,
			Direction:   "received",
		})
	}

	return dto.FriendRequestListResponse{
		Sent:     sentItems,
		Received: receivedItems,
	}, nil
}

func (s *FriendService) AcceptFriendRequest(ctx context.Context, userID, requestID string) (dto.FriendActionResponse, error) {
	friend, err := s.friendRepo.FindByID(ctx, requestID)
	if err != nil {
		return dto.FriendActionResponse{}, fmt.Errorf("lỗi khi tìm lời mời: %w", err)
	}
	if friend == nil {
		return dto.FriendActionResponse{}, errorsapp.New(errorsapp.ErrCodeFriendNotFound)
	}
	if friend.ReceiverID != userID {
		return dto.FriendActionResponse{}, errorsapp.New(errorsapp.ErrCodeFriendNotAuthorized)
	}
	if friend.Status != models.FriendStatusPending {
		return dto.FriendActionResponse{}, errorsapp.New(errorsapp.ErrCodeFriendAlreadyProcessed)
	}

	if err := s.friendRepo.UpdateStatus(ctx, requestID, models.FriendStatusAccepted); err != nil {
		return dto.FriendActionResponse{}, fmt.Errorf("chấp nhận lời mời thất bại: %w", err)
	}

	s.notifService.Create(ctx, friend.SenderID, &userID, models.NotificationTypeFriendAccepted, "đã chấp nhận lời mời kết bạn", nil, &userID, nil)

	return dto.FriendActionResponse{
		Status:  "accepted",
		Message: "Đã chấp nhận lời mời kết bạn",
	}, nil
}

func (s *FriendService) RejectFriendRequest(ctx context.Context, userID, requestID string) (dto.FriendActionResponse, error) {
	friend, err := s.friendRepo.FindByID(ctx, requestID)
	if err != nil {
		return dto.FriendActionResponse{}, fmt.Errorf("lỗi khi tìm lời mời: %w", err)
	}
	if friend == nil {
		return dto.FriendActionResponse{}, errorsapp.New(errorsapp.ErrCodeFriendNotFound)
	}
	if friend.ReceiverID != userID {
		return dto.FriendActionResponse{}, errorsapp.New(errorsapp.ErrCodeFriendNotAuthorized)
	}
	if friend.Status != models.FriendStatusPending {
		return dto.FriendActionResponse{}, errorsapp.New(errorsapp.ErrCodeFriendAlreadyProcessed)
	}

	if err := s.friendRepo.Delete(ctx, requestID); err != nil {
		return dto.FriendActionResponse{}, fmt.Errorf("từ chối lời mời thất bại: %w", err)
	}

	return dto.FriendActionResponse{
		Status:  "rejected",
		Message: "Đã từ chối lời mời kết bạn",
	}, nil
}

func (s *FriendService) GetFriends(ctx context.Context, userID string, page, pageSize int) (dto.FriendListResponse, error) {	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	items, total, err := s.friendRepo.FindAcceptedFriends(ctx, userID, page, pageSize)
	if err != nil {
		return dto.FriendListResponse{}, fmt.Errorf("get friends: %w", err)
	}

	return dto.FriendListResponse{
		Data:     items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasMore:  int64(page*pageSize) < total,
	}, nil
}

func (s *FriendService) SearchFriends(ctx context.Context, userID, keyword string) ([]dto.UserSearchResult, error) {
	if strings.TrimSpace(keyword) == "" {
		return []dto.UserSearchResult{}, nil
	}

	results, err := s.friendRepo.SearchAcceptedFriends(ctx, userID, strings.TrimSpace(keyword), 10)
	if err != nil {
		return nil, fmt.Errorf("search friends: %w", err)
	}
	return results, nil
}

func (s *FriendService) Unfriend(ctx context.Context, userID, targetUserID string) (dto.FriendActionResponse, error) {
	if targetUserID == "" {
		return dto.FriendActionResponse{}, errorsapp.New(errorsapp.ErrCodeFriendTargetRequiredUnfriend)
	}
	if userID == targetUserID {
		return dto.FriendActionResponse{}, errorsapp.New(errorsapp.ErrCodeFriendSelfUnfriend)
	}

	friend, err := s.friendRepo.FindPair(ctx, userID, targetUserID)
	if err != nil {
		return dto.FriendActionResponse{}, fmt.Errorf("lỗi khi tìm quan hệ bạn bè: %w", err)
	}
	if friend == nil || friend.Status != models.FriendStatusAccepted {
		return dto.FriendActionResponse{}, errorsapp.New(errorsapp.ErrCodeFriendNotFriends)
	}

	if err := s.friendRepo.DeletePair(ctx, userID, targetUserID); err != nil {
		return dto.FriendActionResponse{}, fmt.Errorf("hủy kết bạn thất bại: %w", err)
	}

	return dto.FriendActionResponse{
		Status:  "unfriended",
		Message: "Đã hủy kết bạn",
	}, nil
}

func (s *FriendService) GetFriendSuggestions(ctx context.Context, userID string, page, pageSize int) (dto.FriendSuggestionsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	items, total, err := s.friendRepo.GetFriendSuggestions(ctx, userID, page, pageSize)
	if err != nil {
		return dto.FriendSuggestionsResponse{}, fmt.Errorf("get friend suggestions: %w", err)
	}

	return dto.FriendSuggestionsResponse{
		Data:     items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasMore:  int64(page*pageSize) < total,
	}, nil
}
