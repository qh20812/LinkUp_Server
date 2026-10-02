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
	notifRepo      *repository.NotificationRepository
	prefRepo       *repository.NotificationPreferenceRepository
	profileRepo    *repository.ProfileRepository
	pushTokenRepo  *repository.PushTokenRepository
	pushService    *PushService
	hub            *ws.Hub
}

func NewNotificationService(notifRepo *repository.NotificationRepository, prefRepo *repository.NotificationPreferenceRepository, profileRepo *repository.ProfileRepository, hub *ws.Hub, pushTokenRepo *repository.PushTokenRepository, pushService *PushService) *NotificationService {
	return &NotificationService{
		notifRepo:     notifRepo,
		prefRepo:      prefRepo,
		profileRepo:   profileRepo,
		pushTokenRepo: pushTokenRepo,
		pushService:   pushService,
		hub:           hub,
	}
}

// PushOption — tùy chọn bổ sung cho push (không ảnh hưởng notification lưu
// trong DB). WithPushBody override nội dung body push khi nội dung push chi
// tiết hơn content hiển thị trong app (vd: text tin nhắn legacy, text comment).
type PushOption func(*pushConfig)

type pushConfig struct {
	body string
}

func WithPushBody(body string) PushOption {
	return func(c *pushConfig) {
		c.body = body
	}
}

func applyPushOptions(opts []PushOption) pushConfig {
	var cfg pushConfig
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}

// pushThreadID — group push theo hội thoại trên iOS (aps.thread-id). Convention:
// type 'message' mang chat id trong redirect_comment_id (chat.service.go,
// group_message.service.go).
func pushThreadID(notifType models.NotificationType, redirectCommentID *string) string {
	if notifType == models.NotificationTypeMessage && redirectCommentID != nil {
		return *redirectCommentID
	}
	return ""
}

func senderProfileFields(senderMap map[string]dto.SenderProfile, senderID *string) (name, avatar string) {
	if senderID == nil {
		return "", ""
	}
	if p, ok := senderMap[*senderID]; ok {
		return p.DisplayName, p.AvatarURI
	}
	return "", ""
}

func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
}

func (s *NotificationService) Create(ctx context.Context, receiverID string, senderID *string, notifType models.NotificationType, content string, redirectPostID, redirectUserID, redirectCommentID *string, opts ...PushOption) (*models.Notification, error) {
	pref, err := s.prefRepo.GetByUserID(ctx, receiverID)
	if err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodeNotificationCreateFailed, err)
	}

	if pref != nil && !isNotificationEnabled(pref, notifType) {
		return nil, nil
	}

	pushBody := applyPushOptions(opts).body
	if pushBody == "" {
		pushBody = content
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

	senderName, senderAvatar := senderProfileFields(senderMap, senderID)
	s.sendPush(ctx, receiverID, senderName, senderAvatar, pushBody, pushThreadID(notifType, redirectCommentID), map[string]interface{}{
		"type":               string(notifType),
		"redirect_post_id":    redirectPostID,
		"redirect_user_id":    redirectUserID,
		"redirect_comment_id": redirectCommentID,
	})

	return notification, nil
}

func (s *NotificationService) CreateBulk(ctx context.Context, receiverIDs []string, senderID *string, notifType models.NotificationType, content string, redirectPostID, redirectUserID, redirectCommentID *string, opts ...PushOption) ([]models.Notification, error) {
	if len(receiverIDs) == 0 {
		return nil, nil
	}

	pushBody := applyPushOptions(opts).body
	if pushBody == "" {
		pushBody = content
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
	senderName, senderAvatar := senderProfileFields(senderMap, senderID)
	threadID := pushThreadID(notifType, redirectCommentID)
	for i := range notifications {
		resp := dto.ToNotificationResponseList([]models.Notification{notifications[i]}, senderMap)[0]
		s.hub.SendToUser(notifications[i].ReceiverID, ws.OutgoingMessage{
			Type: "notification",
			Data: &resp,
		})

		s.sendPush(ctx, notifications[i].ReceiverID, senderName, senderAvatar, pushBody, threadID, map[string]interface{}{
			"type":               string(notifType),
			"redirect_post_id":    redirectPostID,
			"redirect_user_id":    redirectUserID,
			"redirect_comment_id": redirectCommentID,
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

func (s *NotificationService) UpsertPushToken(ctx context.Context, token *models.PushToken) error {
	return s.pushTokenRepo.Upsert(ctx, token)
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

// sendPush — push rich: title = tên người gửi (fallback "LinkUp" với sender
// hệ thống/admin), body = chi tiết, image = avatar (Android), badge = số chưa
// đọc (iOS), threadId = grouping theo hội thoại (iOS). Token chết
// (DeviceNotRegistered) được xóa khỏi DB ngay khi Expo trả ticket lỗi.
func (s *NotificationService) sendPush(ctx context.Context, receiverID, senderName, senderAvatar, body, threadID string, data map[string]interface{}) {
	if s.pushTokenRepo == nil || s.pushService == nil {
		return
	}
	tokens, err := s.pushTokenRepo.FindByUserID(ctx, receiverID)
	if err != nil || len(tokens) == 0 {
		return
	}

	title := senderName
	if title == "" {
		title = "LinkUp"
	}

	opts := PushOptions{
		Title:    title,
		Body:     truncateRunes(body, 160),
		Data:     data,
		ThreadID: threadID,
	}
	if senderAvatar != "" {
		opts.ImageURL = senderAvatar
	}
	if unread, err := s.notifRepo.GetUnreadCount(ctx, receiverID); err == nil {
		badge := int(unread)
		opts.Badge = &badge
	}

	for _, t := range tokens {
		token := t.PushToken
		go func() {
			if s.pushService.SendBatch(token, opts) {
				_ = s.pushTokenRepo.DeleteByPushToken(context.Background(), token)
			}
		}()
	}
}

func isNotificationEnabled(pref *models.NotificationPreference, notifType models.NotificationType) bool {
	switch notifType {
	case models.NotificationTypeLike:
		return pref.LikeEnabled
	case models.NotificationTypeStoryReact:
		return pref.StoryReactEnabled
	case models.NotificationTypeComment:
		return pref.CommentEnabled
	case models.NotificationTypeFollow:
		return pref.FollowEnabled
	case models.NotificationTypeMessage:
		return pref.MessageEnabled
	case models.NotificationTypeFriendRequest, models.NotificationTypeFriendAccepted:
		return pref.FriendRequestEnabled
	case models.NotificationTypeShare:
		return pref.ShareEnabled
	case models.NotificationTypeMediaApproved, models.NotificationTypeMediaRejected,
		models.NotificationTypeMediaFlagged:
		return pref.MediaEnabled
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
