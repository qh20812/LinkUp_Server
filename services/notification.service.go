package services

import (
	"context"
	"fmt"
	"time"

	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/ws"
)

type NotificationService struct {
	notifRepo *repository.NotificationRepository
	prefRepo  *repository.NotificationPreferenceRepository
	hub       *ws.Hub
}

func NewNotificationService(notifRepo *repository.NotificationRepository, prefRepo *repository.NotificationPreferenceRepository, hub *ws.Hub) *NotificationService {
	return &NotificationService{
		notifRepo: notifRepo,
		prefRepo:  prefRepo,
		hub:       hub,
	}
}

func (s *NotificationService) Create(ctx context.Context, receiverID string, senderID *string, notifType models.NotificationType, content string, redirectPostID, redirectUserID, redirectCommentID *string) (*models.Notification, error) {
	pref, err := s.prefRepo.GetByUserID(ctx, receiverID)
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	if pref != nil && !isNotificationEnabled(pref, notifType) {
		return nil, nil
	}

	now := time.Now().UTC()
	notification := &models.Notification{
		ID:                utils.GenerateUUID(),
		ReceiverID:        receiverID,
		SenderID:          senderID,
		Type:              notifType,
		RedirectPostID:    redirectPostID,
		RedirectUserID:    redirectUserID,
		RedirectCommentID: redirectCommentID,
		Content:           content,
		IsRead:            false,
		CreatedAt:         now,
	}

	if err := s.notifRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	s.hub.SendToUser(receiverID, ws.OutgoingMessage{
		Type: "notification",
		Data: notification,
	})

	return notification, nil
}

func (s *NotificationService) GetList(ctx context.Context, userID string, page, pageSize int, unreadOnly bool) ([]models.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	notifications, err := s.notifRepo.FindByReceiverID(ctx, userID, pageSize, offset, unreadOnly)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.notifRepo.CountByReceiverID(ctx, userID, unreadOnly)
	if err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID string) error {
	return s.notifRepo.MarkAsRead(ctx, notificationID, userID)
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.notifRepo.MarkAllAsRead(ctx, userID)
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	return s.notifRepo.GetUnreadCount(ctx, userID)
}

func (s *NotificationService) GetPreferences(ctx context.Context, userID string) (*models.NotificationPreference, error) {
	return s.prefRepo.GetByUserID(ctx, userID)
}

func (s *NotificationService) UpdatePreferences(ctx context.Context, pref *models.NotificationPreference) error {
	return s.prefRepo.Upsert(ctx, pref)
}

func isNotificationEnabled(pref *models.NotificationPreference, notifType models.NotificationType) bool {
	switch notifType {
	case models.NotificationTypeLike:
		return pref.LikeEnabled
	case models.NotificationTypeComment:
		return pref.CommentEnabled
	case models.NotificationTypeFollow:
		return pref.FollowEnabled
	case models.NotificationTypeMessage:
		return pref.MessageEnabled
	case models.NotificationTypeFriendRequest, models.NotificationTypeFriendAccepted:
		return pref.FriendRequestEnabled
	default:
		return true
	}
}
