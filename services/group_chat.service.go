package services

import (
	"context"
	"fmt"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
)

type GroupChatService struct {
	groupRepo    *repository.GroupChatRepository
	chatRepo     *repository.ChatRepository // Inject thêm để hỗ trợ một số kiểm tra chéo nếu cần
	notifService *NotificationService
	validation   *validations.GroupChatValidation
}

func NewGroupChatService(groupRepo *repository.GroupChatRepository, chatRepo *repository.ChatRepository, notifService *NotificationService, validation *validations.GroupChatValidation) *GroupChatService {
	return &GroupChatService{
		groupRepo:    groupRepo,
		chatRepo:     chatRepo,
		notifService: notifService,
		validation:   validation,
	}
}

func (s *GroupChatService) CreateGroup(ctx context.Context, userID string, name, avatarURI string, memberIDs []string) (*models.Chat, error) {
	if len(name) < 3 || len(name) > 50 {
		return nil, errorsapp.New(errorsapp.ErrCodeGCInvalidName)
	}

	group := models.NewChat(models.ChatTypeGroup, name, avatarURI)
	group.ID = utils.GenerateUUID()
	group.CreatedAt = time.Now().UTC()

	encKey, err := utils.GenerateEncryptionKey()
	if err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodeGCEncryptionKeyFailed, err)
	}
	group.EncryptionKey = encKey

	adminPart := models.NewChatParticipant(group.ID, userID, models.ChatRoleAdmin)
	adminPart.ID = utils.GenerateUUID()
	adminPart.JoinedAt = time.Now().UTC()

	participants := []models.ChatParticipant{adminPart}

	for _, memberID := range memberIDs {
		if memberID == userID || memberID == "" {
			continue
		}

		memberPart := models.NewChatParticipant(group.ID, memberID, models.ChatRoleMember)
		memberPart.ID = utils.GenerateUUID()
		memberPart.JoinedAt = time.Now().UTC()

		participants = append(participants, memberPart)
	}

	if err := s.groupRepo.CreateGroup(ctx, &group, participants); err != nil {
		return nil, err
	}
	return &group, nil
}

func (s *GroupChatService) LeaveGroup(ctx context.Context, chatID, userID, leaveMode, historyMode string) error {
	isMember, err := s.groupRepo.IsUserMember(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errorsapp.New(errorsapp.ErrCodeGCNotMember)
	}

	leaveMode = strings.ToLower(strings.TrimSpace(leaveMode))
	historyMode = strings.ToLower(strings.TrimSpace(historyMode))

	if leaveMode == "" {
		leaveMode = "public"
	}
	if historyMode == "" {
		historyMode = "keep"
	}

	if leaveMode != "silent" && leaveMode != "public" {
		return errorsapp.New(errorsapp.ErrCodeGCInvalidLeaveMode)
	}
	if historyMode != "keep" && historyMode != "anonymize" {
		return errorsapp.New(errorsapp.ErrCodeGCInvalidHistoryMode)
	}

	if historyMode == "anonymize" {
		if err := s.groupRepo.AnonymizeMessagesBySender(ctx, chatID, userID); err != nil {
			return err
		}
	}

	if err := s.groupRepo.LeaveGroup(ctx, chatID, userID); err != nil {
		return err
	}

	var recipientIDs []string
	if leaveMode == "silent" {
		recipientIDs, err = s.groupRepo.GetAdminIDs(ctx, chatID)
	} else {
		recipientIDs, err = s.chatRepo.GetParticipantIDs(ctx, chatID)
	}
	if err != nil {
		return err
	}

	filtered := make([]string, 0, len(recipientIDs))
	for _, id := range recipientIDs {
		if id != "" && id != userID {
			filtered = append(filtered, id)
		}
	}

	if s.notifService != nil && len(filtered) > 0 {
		content := "một thành viên đã rời nhóm"
		if leaveMode == "public" {
			content = "một thành viên đã công khai rời nhóm"
		}

		_, _ = s.notifService.CreateBulk(
			ctx,
			filtered,
			&userID,
			models.NotificationTypeMessage,
			content,
			nil,
			nil,
			nil,
		)
	}

	return nil
}

// 2. CHỨC NĂNG: BAN THÀNH VIÊN (CHẶN QUAY LẠI)
func (s *GroupChatService) BanMember(ctx context.Context, chatID, adminID, targetUserID string) error {
	if adminID == targetUserID {
		return errorsapp.New(errorsapp.ErrCodeGCSelfBan)
	}

	isAdmin, err := s.groupRepo.IsUserAdmin(ctx, chatID, adminID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errorsapp.New(errorsapp.ErrCodeGCAdminOnly)
	}

	isBanned, err := s.groupRepo.IsUserBanned(ctx, chatID, targetUserID)
	if err != nil {
		return err
	}
	if isBanned {
		return errorsapp.New(errorsapp.ErrCodeGCAlreadyBanned)
	}

	banData := &models.GroupChatBan{
		ID:        utils.GenerateUUID(),
		ChatID:    chatID,
		UserID:    targetUserID,
		BannedBy:  adminID,
		CreatedAt: time.Now().UTC(),
	}

	return s.groupRepo.BanUser(ctx, banData)
}

// 3. CẬP NHẬT CHỨC NĂNG: THÊM THÀNH VIÊN (TÍCH HỢP KIỂM TRA BAN REJOIN)
func (s *GroupChatService) addMemberRequest(ctx context.Context, chatID, requesterID, newMemberID string) (string, error) {
	if requesterID == newMemberID {
		return "", errorsapp.New(errorsapp.ErrCodeGCSelfInvite)
	}

	isRequesterMember, err := s.groupRepo.IsUserMember(ctx, chatID, requesterID)
	if err != nil {
		return "", err
	}
	if !isRequesterMember {
		return "", errorsapp.New(errorsapp.ErrCodeGCNotMember)
	}

	groupSettings, err := s.groupRepo.GetSettings(ctx, chatID)
	if err != nil {
		return "", err
	}

	if !groupSettings.AllowMemberAdd {
		isAdmin, err := s.groupRepo.IsUserAdmin(ctx, chatID, requesterID)
		if err != nil {
			return "", err
		}
		if !isAdmin {
			return "", errorsapp.New(errorsapp.ErrCodeGCNotAdminInvite)
		}
	}

	isBanned, err := s.groupRepo.IsUserBanned(ctx, chatID, newMemberID)
	if err != nil {
		return "", err
	}
	if isBanned {
		return "", errorsapp.New(errorsapp.ErrCodeGCBanned)
	}

	isTargetMember, err := s.groupRepo.IsUserMember(ctx, chatID, newMemberID)
	if err != nil {
		return "", err
	}
	if isTargetMember {
		return "", errorsapp.New(errorsapp.ErrCodeGCAlreadyMember)
	}

	pendingReq, err := s.groupRepo.FindPendingMemberRequest(ctx, chatID, newMemberID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", err
	}
	if pendingReq != nil {
		return "", errorsapp.New(errorsapp.ErrCodeGCPendingRequest)
	}

	req := &models.GroupChatMemberRequest{
		ID:           utils.GenerateUUID(),
		ChatID:       chatID,
		RequesterID:  requesterID,
		TargetUserID: newMemberID,
		Status:       models.GroupChatMemberRequestPending,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.groupRepo.CreateMemberRequest(ctx, req); err != nil {
		return "", err
	}

	if s.notifService != nil {
		_, _ = s.notifService.Create(
			ctx,
			newMemberID,
			&requesterID,
			models.NotificationTypeMessage,
			"Bạn có lời mời tham gia nhóm, hãy chấp nhận nếu đồng ý.",
			nil,
			nil,
			&chatID,
		)
	}

	return req.ID, nil
}

func (s *GroupChatService) AddMember(ctx context.Context, chatID, requesterID, newMemberID string) error {
	_, err := s.addMemberRequest(ctx, chatID, requesterID, newMemberID)
	return err
}

func (s *GroupChatService) AddMemberWithRequestID(ctx context.Context, chatID, requesterID, newMemberID string) (string, error) {
	return s.addMemberRequest(ctx, chatID, requesterID, newMemberID)
}

func (s *GroupChatService) GetSettings(ctx context.Context, chatID, userID string) (*dto.GroupChatSettingsResponse, error) {
	isMember, err := s.groupRepo.IsUserMember(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errorsapp.New(errorsapp.ErrCodeGCNotMember)
	}

	chat, err := s.chatRepo.FindChatByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	groupSettings, err := s.groupRepo.GetSettings(ctx, chatID)
	if err != nil {
		return nil, err
	}

	memberSettings, err := s.groupRepo.GetMemberSettings(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}

	return &dto.GroupChatSettingsResponse{
		ChatID:         chatID,
		Name:           chat.Name,
		AvatarURI:      chat.AvatarURI,
		AllowMemberAdd: groupSettings.AllowMemberAdd,
		MemberSettings: dto.GroupChatMemberSettingsResponse{
			NotificationsEnabled: memberSettings.NotificationsEnabled,
		},
	}, nil
}

func (s *GroupChatService) UpdateSettings(ctx context.Context, chatID, requestID string, input *dto.GroupChatSettingsDTO) (*dto.GroupChatSettingsResponse, error) {
	isMember, err := s.groupRepo.IsUserMember(ctx, chatID, requestID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errorsapp.New(errorsapp.ErrCodeGCNotMember)
	}

	isAdmin, err := s.groupRepo.IsUserAdmin(ctx, chatID, requestID)
	if err != nil {
		return nil, err
	}

	if input.NotificationsEnabled != nil {
		memberSettings, err := s.groupRepo.GetMemberSettings(ctx, chatID, requestID)
		if err != nil {
			return nil, err
		}
		memberSettings.NotificationsEnabled = *input.NotificationsEnabled
		if err := s.groupRepo.UpsertMemberSettings(ctx, memberSettings); err != nil {
			return nil, err
		}
	}

	if input.Name != nil {
		if !isAdmin {
			return nil, errorsapp.New(errorsapp.ErrCodeGCAdminOnlyName)
		}
		if utf8.RuneCountInString(*input.Name) < 3 || utf8.RuneCountInString(*input.Name) > 50 {
			return nil, errorsapp.New(errorsapp.ErrCodeGCInvalidGroupName)
		}
	}

	if input.AvatarURI != nil && !isAdmin {
		return nil, errorsapp.New(errorsapp.ErrCodeGCAdminOnlyAvatar)
	}

	if input.AllowMemberAdd != nil && !isAdmin {
		return nil, errorsapp.New(errorsapp.ErrCodeGCAdminOnlyConfig)
	}

	if input.AllowMemberAdd != nil || input.Name != nil || input.AvatarURI != nil {
		if !isAdmin {
			return nil, errorsapp.New(errorsapp.ErrCodeGCAdminOnlyUpdate)
		}

		groupSettings, err := s.groupRepo.GetSettings(ctx, chatID)
		if err != nil {
			return nil, err
		}

		if input.AllowMemberAdd != nil {
			groupSettings.AllowMemberAdd = *input.AllowMemberAdd
		}

		if err := s.groupRepo.UpsertSettings(ctx, groupSettings); err != nil {
			return nil, err
		}

		if input.Name != nil || input.AvatarURI != nil {
			chat, err := s.chatRepo.FindChatByID(ctx, chatID)
			if err != nil {
				return nil, err
			}
			if input.Name != nil {
				chat.Name = *input.Name
			}
			if input.AvatarURI != nil {
				chat.AvatarURI = *input.AvatarURI
			}
			if err := s.chatRepo.UpdateChat(ctx, chat); err != nil {
				return nil, err
			}
		}
	}

	return s.GetSettings(ctx, chatID, requestID)
}

func (s *GroupChatService) TransferAdmin(ctx context.Context, chatID, requestID, targetUserID string) error {
	if requestID == targetUserID {
		return errorsapp.New(errorsapp.ErrCodeGCTransferToSelf)
	}

	isAdmin, err := s.groupRepo.IsUserAdmin(ctx, chatID, requestID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errorsapp.New(errorsapp.ErrCodeGCAdminOnly)
	}

	isTargetMember, err := s.groupRepo.IsUserMember(ctx, chatID, targetUserID)
	if err != nil {
		return err
	}
	if !isTargetMember {
		return errorsapp.New(errorsapp.ErrCodeGCTargetNotMember)
	}

	settings, err := s.groupRepo.GetSettings(ctx, chatID)
	if err != nil {
		return err
	}
	if settings.LastAdminTransferAt != nil {
		nextAllowed := settings.LastAdminTransferAt.AddDate(0, 1, 0)
		if time.Now().UTC().Before(nextAllowed) {
			return errorsapp.Newf(errorsapp.ErrCodeGCAdminCooldown, map[string]any{
				"previous_at": settings.LastAdminTransferAt.UTC().Format(time.RFC3339),
			})
		}
	}

	return s.groupRepo.TransferAdmin(ctx, chatID, requestID, targetUserID, time.Now().UTC())
}

func (s *GroupChatService) TransferOwnership(ctx context.Context, chatID, requesterID, targetUserID string, keepAdmin bool) error {
	if requesterID == targetUserID {
		return errorsapp.New(errorsapp.ErrCodeGCTransferToSelf)
	}

	chat, err := s.chatRepo.FindChatByID(ctx, chatID)
	if err != nil {
		return err
	}

	if chat.CreatorID == nil || *chat.CreatorID != requesterID {
		return errorsapp.New(errorsapp.ErrCodeGCOnlyCreatorTransfer)
	}

	if chat.Type != models.ChatTypeGroup {
		return errorsapp.New(errorsapp.ErrCodeGCNotGroupTransfer)
	}

	isMember, err := s.groupRepo.IsUserMember(ctx, chatID, targetUserID)
	if err != nil {
		return err
	}
	if !isMember {
		return errorsapp.New(errorsapp.ErrCodeGCTargetNotMember)
	}

	return s.groupRepo.TransferOwnership(ctx, chatID, requesterID, targetUserID, keepAdmin, time.Now().UTC())
}

func (s *GroupChatService) MuteMember(ctx context.Context, chatID, adminID, targetUserID, reason string, durationMinutes int) (*models.GroupChatMute, error) {
	if adminID == targetUserID {
		return nil, errorsapp.New(errorsapp.ErrCodeGCSelfMute)
	}

	isAdmin, err := s.groupRepo.IsUserAdmin(ctx, chatID, adminID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, errorsapp.New(errorsapp.ErrCodeGCAdminOnly)
	}

	isMember, err := s.groupRepo.IsUserMember(ctx, chatID, targetUserID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errorsapp.New(errorsapp.ErrCodeGCNotMemberMute)
	}

	if err := s.validation.ValidateMuteInput(reason, durationMinutes); err != nil {
		return nil, err
	}

	var expiresAt *time.Time
	if durationMinutes > 0 {
		t := time.Now().UTC().Add(time.Duration(durationMinutes) * time.Minute)
		expiresAt = &t
	}

	mute := &models.GroupChatMute{
		ID:        utils.GenerateUUID(),
		ChatID:    chatID,
		UserID:    targetUserID,
		MutedBy:   adminID,
		Reason:    reason,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.groupRepo.MuteUser(ctx, mute); err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodeGCMuted, err)
	}

	var expiresStr string
	if expiresAt == nil {
		expiresStr = "vĩnh viễn"
	} else {
		expiresStr = expiresAt.UTC().Format(time.RFC3339)
	}
	content := fmt.Sprintf("Bạn đã bị tắt tiếng trong nhóm (lý do: %s). Hết hạn: %s", reason, expiresStr)
	_, _ = s.notifService.Create(ctx, targetUserID, &adminID, models.NotificationTypeMessage, content, nil, nil, nil)

	return mute, nil
}

func (s *GroupChatService) UnmuteMember(ctx context.Context, chatID, adminID, targetUserID string) error {
	isAdmin, err := s.groupRepo.IsUserAdmin(ctx, chatID, adminID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errorsapp.New(errorsapp.ErrCodeGCAdminOnly)
	}

	if err := s.groupRepo.UnmuteUser(ctx, chatID, targetUserID); err != nil {
		return errorsapp.Wrap(errorsapp.ErrCodeGCMuted, err)
	}

	content := "Quyền gửi tin nhắn đã được mở lại trong nhóm."
	_, _ = s.notifService.Create(ctx, targetUserID, &adminID, models.NotificationTypeMessage, content, nil, nil, nil)
	return nil
}

func (s *GroupChatService) ApproveMemberRequest(ctx context.Context, chatID, targetUserID, requestID string) error {
	req, err := s.groupRepo.FindMemberRequestByID(ctx, requestID)
	if err != nil {
		return err
	}

	if req.ChatID != chatID {
		return errorsapp.New(errorsapp.ErrCodeGCRequestNotOwn)
	}
	if req.TargetUserID != targetUserID {
		return errorsapp.New(errorsapp.ErrCodeGCRequestNotOwn)
	}
	if req.Status != models.GroupChatMemberRequestPending {
		return errorsapp.New(errorsapp.ErrCodeGCRequestAlreadyHandled)
	}

	isBanned, err := s.groupRepo.IsUserBanned(ctx, chatID, targetUserID)
	if err != nil {
		return err
	}
	if isBanned {
		return errorsapp.New(errorsapp.ErrCodeGCBanned)
	}

	isTargetMember, err := s.groupRepo.IsUserMember(ctx, chatID, targetUserID)
	if err != nil {
		return err
	}
	if isTargetMember {
		return errorsapp.New(errorsapp.ErrCodeGCAlreadyMember)
	}

	if err := s.groupRepo.ApproveMemberRequest(ctx, requestID); err != nil {
		return err
	}

	if s.notifService != nil {
		_, _ = s.notifService.Create(
			ctx,
			req.RequesterID,
			&targetUserID,
			models.NotificationTypeMessage,
			"Người được mời đã chấp nhận tham gia nhóm.",
			nil,
			nil,
			&chatID,
		)
	}

	return nil
}

func (s *GroupChatService) RejectMemberRequest(ctx context.Context, chatID, targetUserID, requestID string) error {
	req, err := s.groupRepo.FindMemberRequestByID(ctx, requestID)
	if err != nil {
		return err
	}

	if req.ChatID != chatID {
		return errorsapp.New(errorsapp.ErrCodeGCRequestNotOwn)
	}
	if req.TargetUserID != targetUserID {
		return errorsapp.New(errorsapp.ErrCodeGCRequestNotOwn)
	}
	if req.Status != models.GroupChatMemberRequestPending {
		return errorsapp.New(errorsapp.ErrCodeGCRequestAlreadyHandled)
	}

	now := time.Now().UTC()
	req.Status = models.GroupChatMemberRequestRejected
	req.RespondedAt = &now

	return s.groupRepo.RejectMemberRequest(ctx, requestID)
}

func (s *GroupChatService) EnsureGroupMember(ctx context.Context, chatID, userID string) error {
	isMember, err := s.groupRepo.IsUserMember(ctx, chatID, userID)
	if err != nil {
		return errorsapp.Wrap(errorsapp.ErrCodeGCNotMember, err)
	}
	if !isMember {
		return errorsapp.New(errorsapp.ErrCodeGCNotMember)
	}
	return nil
}
