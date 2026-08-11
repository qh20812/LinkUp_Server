package services

import (
	"context"
	"time"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/ws"
)

type NotificationService struct {
	notifRepo   *repository.NotificationRepository
	prefRepo    *repository.NotificationPreferenceRepository
	profileRepo *repository.ProfileRepository
	hub         *ws.Hub
}

func NewNotificationService(notifRepo *repository.NotificationRepository, prefRepo *repository.NotificationPreferenceRepository, profileRepo *repository.ProfileRepository, hub *ws.Hub) *NotificationService {
	return &NotificationService{
		notifRepo:   notifRepo,
		prefRepo:    prefRepo,
		profileRepo: profileRepo,
		hub:         hub,
	}
}

func (s *NotificationService) Create(ctx context.Context, receiverID string, senderID *string, notifType models.NotificationType, content string, redirectPostID, redirectUserID, redirectCommentID *string) (*models.Notification, error) {
	pref, err := s.prefRepo.GetByUserID(ctx, receiverID)
	if err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodeNotificationCreateFailed, err)
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
		return nil, errorsapp.Wrap(errorsapp.ErrCodeNotificationCreateFailed, err)
	}

	senderMap := s.loadSenderProfiles(ctx, senderID)
	resp := dto.ToNotificationResponseList([]models.Notification{*notification}, senderMap)[0]
	s.hub.SendToUser(receiverID, ws.OutgoingMessage{
		Type: "notification",
		Data: &resp,
	})

	return notification, nil
}

func (s *NotificationService) CreateBulk(ctx context.Context, receiverIDs []string, senderID *string, notifType models.NotificationType, content string, redirectPostID, redirectUserID, redirectCommentID *string) ([]models.Notification, error) {
	if len(receiverIDs) == 0 {
		return nil, nil
	}

	now := time.Now().UTC()
	var notifications []models.Notification

	for _, receiverID := range receiverIDs {
		pref, err := s.prefRepo.GetByUserID(ctx, receiverID)
		if err != nil {
			continue
		}
		if pref != nil && !isNotificationEnabled(pref, notifType) {
			continue
		}

		notifications = append(notifications, models.Notification{
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
		})
	}

	if len(notifications) == 0 {
		return nil, nil
	}

	if err := s.notifRepo.CreateBulk(ctx, notifications); err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodeNotificationBulkFailed, err)
	}

	senderMap := s.loadSenderProfiles(ctx, senderID)
	for i := range notifications {
		resp := dto.ToNotificationResponseList([]models.Notification{notifications[i]}, senderMap)[0]
		s.hub.SendToUser(notifications[i].ReceiverID, ws.OutgoingMessage{
			Type: "notification",
			Data: &resp,
		})
	}

	return notifications, nil
}

func (s *NotificationService) GetList(ctx context.Context, userID string, page, pageSize int, unreadOnly bool) ([]dto.NotificationResponse, int64, error) {
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

	senderIDs := make([]string, 0, len(notifications))
	for _, n := range notifications {
		if n.SenderID != nil {
			senderIDs = append(senderIDs, *n.SenderID)
		}
	}

	senderMap := make(map[string]dto.SenderProfile)
	if len(senderIDs) > 0 {
		profiles, err := s.profileRepo.FindByIDs(ctx, senderIDs)
		if err == nil {
			for _, p := range profiles {
				name := p.DisplayName
				if name == "" {
					name = "User"
				}
				senderMap[p.UserID] = dto.SenderProfile{
					DisplayName: name,
					AvatarURI:   p.AvatarURI,
				}
			}
		}
	}

	return dto.ToNotificationResponseList(notifications, senderMap), total, nil
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

func (s *NotificationService) loadSenderProfiles(ctx context.Context, senderID *string) map[string]dto.SenderProfile {
	if senderID == nil {
		return nil
	}

	profiles, err := s.profileRepo.FindByIDs(ctx, []string{*senderID})
	if err != nil {
		return nil
	}

	senderMap := make(map[string]dto.SenderProfile, len(profiles))
	for _, p := range profiles {
		name := p.DisplayName
		if name == "" {
			name = "User"
		}
		senderMap[p.UserID] = dto.SenderProfile{
			DisplayName: name,
			AvatarURI:   p.AvatarURI,
		}
	}
	return senderMap
}

func isNotificationEnabled(pref *models.NotificationPreference, notifType models.NotificationType) bool {
	switch notifType {
	case models.NotificationTypeLike, models.NotificationTypeShare:
		return pref.LikeEnabled
	case models.NotificationTypeComment:
		return pref.CommentEnabled
	case models.NotificationTypeFollow:
		return pref.FollowEnabled
	case models.NotificationTypeMessage:
		return pref.MessageEnabled
	case models.NotificationTypeFriendRequest, models.NotificationTypeFriendAccepted:
		return pref.FriendRequestEnabled
	case models.NotificationTypeCommunityJoinRequest, models.NotificationTypeCommunityJoinApproved,
		models.NotificationTypeCommunityJoinRejected, models.NotificationTypeCommunityRoleChanged,
		models.NotificationTypeCommunityMemberLeft, models.NotificationTypeCommunityMemberKicked,
		models.NotificationTypeCommunityGroupChatAdded,
		models.NotificationTypeCommunityInviteCodeUsed,
		models.NotificationTypeCommunityInvitationReceived,
		models.NotificationTypeCommunityInvitationAccepted:
		return pref.CommunityEnabled
	case models.NotificationTypeVoiceCall:
		return pref.VoiceCallEnabled
	default:
		return true
	}
}
