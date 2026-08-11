package services

import (
	"context"
	"fmt"
	errorsapp "linkup/errors"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"mime/multipart"
	"strings"
	"time"
	"unicode/utf8"
)

type CommunityService struct {
	repo         *repository.CommunityRepository
	authRepo     *repository.AuthRepository
	profileRepo  *repository.ProfileRepository
	mediaService MediaService
	notifService *NotificationService
	groupRole    *utils.GroupRoleChecker
	validation   *validations.CommunityValidation
}

func NewCommunityService(repo *repository.CommunityRepository, validation *validations.CommunityValidation, authRepo *repository.AuthRepository, profileRepo *repository.ProfileRepository, mediaService MediaService, notifService *NotificationService) *CommunityService {
	return &CommunityService{
		repo:         repo,
		validation:   validation,
		authRepo:     authRepo,
		profileRepo:  profileRepo,
		mediaService: mediaService,
		notifService: notifService,
		groupRole:    utils.NewGroupRoleChecker(repo.GetUserRole),
	}
}

func (s *CommunityService) CreateCommunity(ctx context.Context, creatorID, name, description, avatarURI string, autoApprove bool) (*models.Community, *models.Chat, error) {
	if err := s.validation.ValidateCreateCommunity(name, description, avatarURI); err != nil {
		return nil, nil, err
	}

	isAdmin, err := s.authRepo.HasRole(ctx, creatorID, models.RoleAdmin)
	if err != nil {
		return nil, nil, errorsapp.Wrap(errorsapp.ErrCodeInternal, err)
	}
	isSuperAdmin, err := s.authRepo.HasRole(ctx, creatorID, models.RoleSuperAdmin)
	if err != nil {
		return nil, nil, errorsapp.Wrap(errorsapp.ErrCodeInternal, err)
	}
	if isAdmin || isSuperAdmin {
		return nil, nil, errorsapp.New(errorsapp.ErrCodeAdminCannotCreate)
	}

	creator, err := s.authRepo.FindByID(ctx, creatorID)
	if err != nil {
		return nil, nil, errorsapp.New(errorsapp.ErrCodeUserNotFound)
	}
	if !creator.IsActive() {
		return nil, nil, errorsapp.New(errorsapp.ErrCodeAccountInactive)
	}

	taken, err := s.repo.IsNameTaken(ctx, name)
	if err != nil {
		return nil, nil, errorsapp.Wrap(errorsapp.ErrCodeInternal, err)
	}
	if taken {
		return nil, nil, validations.ErrCommunityNameExists
	}

	now := time.Now().UTC()
	community := models.NewCommunity(creatorID, name, description, avatarURI)
	community.ID = utils.GenerateUUID()
	community.AutoApprove = autoApprove
	community.CreatedAt = now

	adminMember := models.NewGroupMember(community.ID, creatorID)
	adminMember.ID = utils.GenerateUUID()
	adminMember.JoinedAt = now
	adminMember.Points = 500

	scopeType := models.ScopeTypeCommunity
	userRoles := []models.UserRole{
		models.NewScopedUserRole(creatorID, "", community.ID, scopeType),
		models.NewScopedUserRole(creatorID, "", community.ID, scopeType),
	}

	var communityAdminRole, groupAdminRole models.Role
	if err := s.repo.FindRoleByName(ctx, models.RoleCommunityAdmin, &communityAdminRole); err != nil {
		return nil, nil, err
	}
	if err := s.repo.FindRoleByName(ctx, models.RoleGroupAdmin, &groupAdminRole); err != nil {
		return nil, nil, err
	}
	userRoles[0].RoleID = communityAdminRole.ID
	userRoles[1].RoleID = groupAdminRole.ID

	encKey, err := utils.GenerateEncryptionKey()
	if err != nil {
		return nil, nil, errorsapp.New(errorsapp.ErrCodeEncryptionKeyFailed)
	}

	chat := models.NewChat(models.ChatTypeGroup, community.Name, community.AvatarURI)
	chat.ID = utils.GenerateUUID()
	chat.EncryptionKey = encKey
	chat.CommunityID = &community.ID
	chat.CreatedAt = now

	adminParticipant := models.NewChatParticipant(chat.ID, creatorID, models.ChatRoleAdmin)
	adminParticipant.ID = utils.GenerateUUID()
	adminParticipant.JoinedAt = now

	if err := s.repo.CreateCommunityWithDefaultGroupChat(ctx, &community, &adminMember, userRoles, &chat, []models.ChatParticipant{adminParticipant}); err != nil {
		return nil, nil, err
	}

	s.notifService.Create(ctx, creatorID, nil,
		models.NotificationTypeCommunityGroupChatAdded,
		"Tạo cộng đồng thành công! Group chat mặc định đã sẵn sàng",
		nil, nil, nil)

	return &community, &chat, nil
}

func (s *CommunityService) SetCommunityBackground(ctx context.Context, userID, communityID string, file *multipart.FileHeader) error {
	if err := s.groupRole.RequireRole(ctx, communityID, userID, models.GroupRoleAdmin); err != nil {
		return err
	}

	community, err := s.repo.FindByID(ctx, communityID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeCommunityNotFound)
	}

	user, err := s.authRepo.FindByID(ctx, userID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeUserNotFound)
	}
	if !user.IsActive() {
		return errorsapp.New(errorsapp.ErrCodeAccountInactive)
	}

	src, err := file.Open()
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeImageReadFailed)
	}
	if _, _, err := validations.ValidateImageDimensions(src, validations.DimensionConstraint{
		MinWidth:  800,
		MinHeight: 400,
		MaxWidth:  4096,
		MaxHeight: 4096,
	}); err != nil {
		src.Close()
		return err
	}
	src.Close()

	media, err := s.mediaService.UploadMedia(ctx, userID, file)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeBackgroundUploadFailed)
	}

	if media.Status == models.MediaStatusRejected {
		return errorsapp.New(errorsapp.ErrCodeBackgroundRejected)
	}

	if err := s.validation.ValidateBackgroundURI(media.FileURI); err != nil {
		return err
	}

	if err := s.repo.UpdateBackground(ctx, community.ID, media.FileURI); err != nil {
		return errorsapp.New(errorsapp.ErrCodeBackgroundUpdateFailed)
	}

	return nil
}

func (s *CommunityService) RequestJoin(ctx context.Context, userID, communityID, inviteCode, invitationID string) (*dto.JoinResult, error) {
	user, err := s.authRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeUserNotFound)
	}
	if !user.IsActive() {
		return nil, errorsapp.New(errorsapp.ErrCodeAccountInactive)
	}

	community, err := s.repo.FindByID(ctx, communityID)
	if err != nil {
		return nil, validations.ErrCommunityNotFound
	}

	isMember, err := s.repo.IsUserMember(ctx, communityID, userID)
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeMemberCheckFailed)
	}
	if isMember {
		return nil, validations.ErrAlreadyMember
	}

	switch community.Privacy {
	case models.PrivacyCode:
		return s.joinByCode(ctx, userID, community, inviteCode)
	case models.PrivacyInvitationOnly:
		return s.joinByInvitation(ctx, userID, community, invitationID)
	default:
		// Nếu có mã mời hoặc lời mời được cung cấp, vẫn validate kể cả community công khai
		if inviteCode != "" {
			return s.joinByCode(ctx, userID, community, inviteCode)
		}
		if invitationID != "" {
			return s.joinByInvitation(ctx, userID, community, invitationID)
		}
		return s.joinPublic(ctx, userID, community)
	}
}

func (s *CommunityService) joinPublic(ctx context.Context, userID string, community *models.Community) (*dto.JoinResult, error) {
	if community.AutoApprove {
		groupChat, err := s.repo.FindDefaultGroupChatByCommunity(ctx, community.ID)
		if err != nil {
			return nil, errorsapp.New(errorsapp.ErrCodeGroupChatNotFound)
		}

		if err := s.repo.AddCommunityMemberAndGroupChat(ctx, community.ID, userID, groupChat.ID); err != nil {
			return nil, err
		}

		s.notifService.Create(ctx, userID, &community.CreatorID,
			models.NotificationTypeCommunityGroupChatAdded,
			"Bạn đã tham gia cộng đồng và group chat mặc định",
			nil, &community.ID, nil)

		return &dto.JoinResult{AutoApproved: true}, nil
	}

	existing, err := s.repo.FindPendingJoinRequestByUserAndCommunity(ctx, community.ID, userID)
	if err == nil && existing != nil {
		return nil, validations.ErrJoinRequestPending
	}

	s.repo.DeleteNonPendingJoinRequests(ctx, community.ID, userID)

	now := time.Now().UTC()
	joinReq := models.NewCommunityJoinRequest(community.ID, userID)
	joinReq.ID = utils.GenerateUUID()
	joinReq.CreatedAt = now

	if err := s.repo.CreateJoinRequest(ctx, &joinReq); err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeJoinRequestFailed)
	}

	s.notifService.Create(ctx, community.CreatorID, &userID, models.NotificationTypeCommunityJoinRequest, "đã gửi yêu cầu tham gia cộng đồng", nil, &userID, nil)

	return &dto.JoinResult{RequestID: joinReq.ID, AutoApproved: false}, nil
}

func (s *CommunityService) joinByCode(ctx context.Context, userID string, community *models.Community, code string) (*dto.JoinResult, error) {
	if code == "" {
		return nil, errorsapp.New(errorsapp.ErrCodeInviteCodeRequired)
	}

	inviteCode, err := s.repo.FindInviteCodeByCode(ctx, code)
	if err != nil {
		return nil, validations.ErrInviteCodeNotFound
	}

	if err := s.validation.ValidateInviteCode(inviteCode); err != nil {
		return nil, err
	}

	groupChat, err := s.repo.FindDefaultGroupChatByCommunity(ctx, community.ID)
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeGroupChatNotFound)
	}

	if err := s.repo.AddCommunityMemberAndGroupChat(ctx, community.ID, userID, groupChat.ID); err != nil {
		return nil, err
	}

	s.repo.IncrementInviteCodeUsedCount(ctx, nil, inviteCode.ID)

	s.notifService.Create(ctx, userID, nil,
		models.NotificationTypeCommunityInviteCodeUsed,
		"Tham gia cộng đồng bằng mã mời thành công",
		nil, &community.ID, nil)

	return &dto.JoinResult{AutoApproved: true}, nil
}

func (s *CommunityService) joinByInvitation(ctx context.Context, userID string, community *models.Community, invitationID string) (*dto.JoinResult, error) {
	if invitationID == "" {
		return nil, errorsapp.New(errorsapp.ErrCodeInvitationRequired)
	}

	invitation, err := s.repo.FindInvitationByID(ctx, invitationID)
	if err != nil {
		return nil, validations.ErrInvitationNotFound
	}

	if invitation.InviteeID != userID {
		return nil, validations.ErrInvitationNotFound
	}

	if invitation.Status != models.InvitationStatusPending {
		return nil, validations.ErrInvitationAlreadyHandled
	}

	groupChat, err := s.repo.FindDefaultGroupChatByCommunity(ctx, community.ID)
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeGroupChatNotFound)
	}

	if err := s.repo.AddCommunityMemberAndGroupChat(ctx, community.ID, userID, groupChat.ID); err != nil {
		return nil, err
	}

	s.repo.UpdateInvitationStatus(ctx, nil, invitationID, models.InvitationStatusAccepted)

	s.notifService.Create(ctx, userID, &invitation.InviterID,
		models.NotificationTypeCommunityInvitationAccepted,
		"Bạn đã tham gia cộng đồng theo lời mời",
		nil, &community.ID, nil)

	return &dto.JoinResult{AutoApproved: true}, nil
}

func (s *CommunityService) ListPendingRequests(ctx context.Context, adminID, communityID string) (dto.JoinRequestListResponse, error) {
	if err := s.groupRole.RequireRole(ctx, communityID, adminID, models.GroupRoleAdmin); err != nil {
		return dto.JoinRequestListResponse{}, validations.ErrNotCommunityAdmin
	}

	requests, err := s.repo.FindPendingJoinRequestsByCommunity(ctx, communityID)
	if err != nil {
		return dto.JoinRequestListResponse{}, errorsapp.Wrap(errorsapp.ErrCodeInternal, err)
	}

	items := make([]dto.JoinRequestItem, 0, len(requests))
	for _, req := range requests {
		profile, err := s.profileRepo.FindByUserID(ctx, req.UserID)
		displayName := ""
		avatarURI := ""
		if err == nil && profile != nil {
			displayName = profile.DisplayName
			avatarURI = profile.AvatarURI
		}
		items = append(items, dto.JoinRequestItem{
			ID:          req.ID,
			UserID:      req.UserID,
			DisplayName: displayName,
			AvatarURI:   avatarURI,
			Status:      string(req.Status),
			CreatedAt:   req.CreatedAt,
		})
	}

	return dto.JoinRequestListResponse{Requests: items}, nil
}

func (s *CommunityService) ApproveJoinRequest(ctx context.Context, adminID, requestID string) error {
	req, err := s.repo.FindJoinRequestByID(ctx, requestID)
	if err != nil {
		return validations.ErrJoinRequestNotFound
	}
	if req.Status != models.JoinRequestStatusPending {
		return validations.ErrJoinRequestAlreadyHandled
	}

	if err := s.groupRole.RequireRole(ctx, req.CommunityID, adminID, models.GroupRoleAdmin); err != nil {
		return validations.ErrNotCommunityAdmin
	}

	groupChat, err := s.repo.FindDefaultGroupChatByCommunity(ctx, req.CommunityID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeGroupChatNotFound)
	}

	if err := s.repo.ApproveJoinRequest(ctx, requestID, &groupChat.ID); err != nil {
		return err
	}

	s.notifService.Create(ctx, req.UserID, &adminID,
		models.NotificationTypeCommunityGroupChatAdded,
		"Bạn đã được duyệt vào cộng đồng và thêm vào group chat",
		nil, &req.CommunityID, nil)

	return nil
}

func (s *CommunityService) RejectJoinRequest(ctx context.Context, adminID, requestID string) error {
	req, err := s.repo.FindJoinRequestByID(ctx, requestID)
	if err != nil {
		return validations.ErrJoinRequestNotFound
	}
	if req.Status != models.JoinRequestStatusPending {
		return validations.ErrJoinRequestAlreadyHandled
	}

	if err := s.groupRole.RequireRole(ctx, req.CommunityID, adminID, models.GroupRoleAdmin); err != nil {
		return validations.ErrNotCommunityAdmin
	}

	if err := s.repo.RejectJoinRequest(ctx, requestID); err != nil {
		return err
	}

	s.notifService.Create(ctx, req.UserID, &adminID, models.NotificationTypeCommunityJoinRejected, "đã từ chối yêu cầu tham gia cộng đồng", nil, &adminID, nil)

	return nil
}

func (s *CommunityService) UpdateMemberRole(ctx context.Context, adminID, communityID, memberID, newRole string) error {
	if err := s.validation.ValidateUpdateRole(newRole); err != nil {
		return err
	}

	if _, err := s.repo.FindByID(ctx, communityID); err != nil {
		return validations.ErrCommunityNotFound
	}

	if err := s.groupRole.RequireRole(ctx, communityID, adminID, models.GroupRoleAdmin); err != nil {
		return validations.ErrNotCommunityAdmin
	}

	if adminID == memberID {
		return validations.ErrCannotChangeOwnRole
	}

	isMember, err := s.repo.IsUserMember(ctx, communityID, memberID)
	if err != nil || !isMember {
		return validations.ErrMemberNotFound
	}

	isCreator, err := s.repo.IsUserCreator(ctx, communityID, memberID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeMemberCheckFailed)
	}
	if isCreator {
		return validations.ErrCannotTargetAdmin
	}

	newRoleName := models.RoleName(newRole)
	if err := s.repo.UpdateUserRole(ctx, communityID, memberID, newRoleName); err != nil {
		return errorsapp.New(errorsapp.ErrCodeRoleUpdateFailed)
	}

	s.notifService.Create(ctx, memberID, &adminID, models.NotificationTypeCommunityRoleChanged, "đã thay đổi vai trò của bạn trong cộng đồng", nil, &adminID, nil)

	return nil
}

func (s *CommunityService) KickMember(ctx context.Context, adminID, communityID, memberID, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return validations.ErrKickReasonRequired
	}
	if utf8.RuneCountInString(reason) < 3 {
		return validations.ErrKickReasonTooShort
	}
	if utf8.RuneCountInString(reason) > 500 {
		return validations.ErrKickReasonTooLong
	}

	community, err := s.repo.FindByID(ctx, communityID)
	if err != nil {
		return validations.ErrCommunityNotFound
	}

	if err := s.groupRole.RequireRole(ctx, communityID, adminID, models.GroupRoleAdmin); err != nil {
		return validations.ErrNotCommunityAdmin
	}

	if adminID == memberID {
		return validations.ErrCannotChangeOwnRole
	}

	isMember, err := s.repo.IsUserMember(ctx, communityID, memberID)
	if err != nil || !isMember {
		return validations.ErrMemberNotFound
	}

	isCreator, err := s.repo.IsUserCreator(ctx, communityID, memberID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeMemberCheckFailed)
	}
	if isCreator {
		return validations.ErrCannotKickCreator
	}

	isAdmin, err := s.repo.IsUserAdmin(ctx, communityID, memberID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeMemberCheckFailed)
	}
	if isAdmin && adminID != community.CreatorID {
		return validations.ErrCannotKickAdmin
	}

	if err := s.repo.RemoveMember(ctx, communityID, memberID); err != nil {
		return errorsapp.New(errorsapp.ErrCodeKickFailed)
	}

	content := fmt.Sprintf("bạn đã bị đuổi khỏi cộng đồng %s với lý do: %s", community.Name, reason)
	s.notifService.Create(ctx, memberID, &adminID, models.NotificationTypeCommunityMemberKicked, content, nil, nil, nil)

	return nil
}

func (s *CommunityService) GetCommunityMembers(ctx context.Context, communityID string) (dto.CommunityMemberListResponse, error) {
	if _, err := s.repo.FindByID(ctx, communityID); err != nil {
		return dto.CommunityMemberListResponse{}, validations.ErrCommunityNotFound
	}

	members, err := s.repo.FindMembersByCommunity(ctx, communityID)
	if err != nil {
		return dto.CommunityMemberListResponse{}, err
	}

	return dto.CommunityMemberListResponse{Members: members}, nil
}

func (s *CommunityService) LeaveCommunity(ctx context.Context, userID, communityID string, quiet bool) error {
	community, err := s.repo.FindByID(ctx, communityID)
	if err != nil {
		return validations.ErrCommunityNotFound
	}

	isMember, err := s.repo.IsUserMember(ctx, communityID, userID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeMemberCheckFailed)
	}
	if !isMember {
		return validations.ErrMemberNotFound
	}

	isCreator, err := s.repo.IsUserCreator(ctx, communityID, userID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeMemberCheckFailed)
	}
	if isCreator {
		return validations.ErrCreatorCannotLeave
	}

	if err := s.repo.RemoveMember(ctx, communityID, userID); err != nil {
		return err
	}

	if quiet {
		adminIDs, err := s.repo.FindCommunityAdmins(ctx, communityID)
		if err == nil && len(adminIDs) > 0 {
			s.notifService.CreateBulk(ctx, adminIDs, &userID, models.NotificationTypeCommunityMemberLeft, "đã rời khỏi cộng đồng", nil, &community.CreatorID, nil)
		}
	} else {
		memberIDs, err := s.repo.FindCommunityMemberIDs(ctx, communityID)
		if err == nil {
			var receiverIDs []string
			for _, id := range memberIDs {
				if id != userID {
					receiverIDs = append(receiverIDs, id)
				}
			}
			if len(receiverIDs) > 0 {
				s.notifService.CreateBulk(ctx, receiverIDs, &userID, models.NotificationTypeCommunityMemberLeft, "đã rời khỏi cộng đồng", nil, &community.CreatorID, nil)
			}
		}
	}

	return nil
}

func (s *CommunityService) TransferOwnership(ctx context.Context, requesterID, communityID, targetUserID string, keepAdmin bool) error {
	if requesterID == targetUserID {
		return errorsapp.New(errorsapp.ErrCodeTransferOwnToSelf)
	}

	community, err := s.repo.FindByID(ctx, communityID)
	if err != nil {
		return validations.ErrCommunityNotFound
	}

	if community.CreatorID != requesterID {
		return errorsapp.New(errorsapp.ErrCodeOnlyCreatorCanTransfer)
	}

	isMember, err := s.repo.IsUserMember(ctx, communityID, targetUserID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeMemberCheckFailed)
	}
	if !isMember {
		return validations.ErrMemberNotFound
	}

	if err := s.repo.TransferCommunityOwnership(ctx, communityID, requesterID, targetUserID, keepAdmin); err != nil {
		return errorsapp.New(errorsapp.ErrCodeTransferFailed)
	}

	s.notifService.Create(ctx, targetUserID, &requesterID,
		models.NotificationTypeCommunityRoleChanged,
		"Bạn đã được chuyển quyền sở hữu cộng đồng",
		nil, &communityID, nil)

	return nil
}

// ── Invite code management ──────────────────────────────────────────────────

func (s *CommunityService) CreateInviteCode(ctx context.Context, adminID, communityID string, maxUses int, expiresAt *time.Time) (*dto.InviteCodeResponse, error) {
	if err := s.groupRole.RequireRole(ctx, communityID, adminID, models.GroupRoleAdmin); err != nil {
		return nil, err
	}

	code, err := utils.GenerateInviteCode()
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeInviteCodeCreateFail)
	}

	now := time.Now().UTC()
	inviteCode := &models.CommunityInviteCode{
		ID:          utils.GenerateUUID(),
		CommunityID: communityID,
		Code:        code,
		CreatedBy:   adminID,
		MaxUses:     maxUses,
		ExpiresAt:   expiresAt,
		IsActive:    true,
		CreatedAt:   now,
	}

	if err := s.repo.CreateInviteCode(ctx, inviteCode); err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeInviteCodeSaveFail)
	}

	return &dto.InviteCodeResponse{
		ID:        inviteCode.ID,
		Code:      inviteCode.Code,
		MaxUses:   inviteCode.MaxUses,
		UsedCount: inviteCode.UsedCount,
		ExpiresAt: inviteCode.ExpiresAt,
		IsActive:  inviteCode.IsActive,
		CreatedAt: inviteCode.CreatedAt,
	}, nil
}

func (s *CommunityService) ListInviteCodes(ctx context.Context, adminID, communityID string) ([]dto.InviteCodeResponse, error) {
	if err := s.groupRole.RequireRole(ctx, communityID, adminID, models.GroupRoleAdmin); err != nil {
		return nil, err
	}

	codes, err := s.repo.ListInviteCodesByCommunity(ctx, communityID)
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeInviteCodeListFail)
	}

	items := make([]dto.InviteCodeResponse, 0, len(codes))
	for _, c := range codes {
		items = append(items, dto.InviteCodeResponse{
			ID:        c.ID,
			Code:      c.Code,
			MaxUses:   c.MaxUses,
			UsedCount: c.UsedCount,
			ExpiresAt: c.ExpiresAt,
			IsActive:  c.IsActive,
			CreatedAt: c.CreatedAt,
		})
	}
	return items, nil
}

func (s *CommunityService) DeactivateInviteCode(ctx context.Context, adminID, codeID string) error {
	inviteCode, err := s.repo.FindInviteCodeByID(ctx, codeID)
	if err != nil {
		return validations.ErrInviteCodeNotFound
	}

	if err := s.groupRole.RequireRole(ctx, inviteCode.CommunityID, adminID, models.GroupRoleAdmin); err != nil {
		return err
	}

	if err := s.repo.DeactivateInviteCode(ctx, inviteCode.ID); err != nil {
		return errorsapp.New(errorsapp.ErrCodeInviteCodeDeactFail)
	}

	return nil
}

// ── Direct invitation ──────────────────────────

func (s *CommunityService) SendInvitation(ctx context.Context, inviterID, communityID, inviteeID string) (*dto.InvitationItem, error) {
	if err := s.groupRole.RequireRole(ctx, communityID, inviterID, models.GroupRoleAdmin); err != nil {
		return nil, err
	}

	if inviteeID == inviterID {
		return nil, validations.ErrCannotInviteSelf
	}

	isMember, err := s.repo.IsUserMember(ctx, communityID, inviteeID)
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeMemberCheckFailed)
	}
	if isMember {
		return nil, validations.ErrAlreadyMember
	}

	existing, err := s.repo.FindPendingInvitation(ctx, communityID, inviteeID)
	if err == nil && existing != nil {
		return nil, validations.ErrJoinRequestPending
	}

	now := time.Now().UTC()
	invitation := &models.CommunityInvitation{
		ID:          utils.GenerateUUID(),
		CommunityID: communityID,
		InviterID:   inviterID,
		InviteeID:   inviteeID,
		Status:      models.InvitationStatusPending,
		CreatedAt:   now,
	}

	if err := s.repo.CreateInvitation(ctx, invitation); err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeInvitationSendFailed)
	}

	community, err := s.repo.FindByID(ctx, communityID)
	communityName := ""
	if err == nil {
		communityName = community.Name
	}

	s.notifService.Create(ctx, inviteeID, &inviterID,
		models.NotificationTypeCommunityInvitationReceived,
		"Bạn đã nhận được lời mời tham gia cộng đồng",
		nil, &communityID, nil)

	return &dto.InvitationItem{
		ID:            invitation.ID,
		CommunityID:   communityID,
		CommunityName: communityName,
		InviterID:     inviterID,
		Status:        string(invitation.Status),
		CreatedAt:     invitation.CreatedAt,
	}, nil
}

func (s *CommunityService) ListMyInvitations(ctx context.Context, userID string) ([]dto.InvitationItem, error) {
	invites, err := s.repo.ListPendingInvitationsByInvitee(ctx, userID)
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeInvitationListFailed)
	}

	items := make([]dto.InvitationItem, 0, len(invites))
	for _, inv := range invites {
		items = append(items, dto.InvitationItem{
			ID:            inv.ID,
			CommunityID:   inv.CommunityID,
			CommunityName: inv.CommunityName,
			InviterID:     inv.InviterID,
			Status:        string(inv.Status),
			CreatedAt:     inv.CreatedAt,
		})
	}
	return items, nil
}

func (s *CommunityService) RespondInvitation(ctx context.Context, userID, invitationID string, accept bool) error {
	invitation, err := s.repo.FindInvitationByID(ctx, invitationID)
	if err != nil {
		return validations.ErrInvitationNotFound
	}

	if invitation.InviteeID != userID {
		return validations.ErrInvitationNotFound
	}

	if invitation.Status != models.InvitationStatusPending {
		return validations.ErrInvitationAlreadyHandled
	}

	if accept {
		groupChat, err := s.repo.FindDefaultGroupChatByCommunity(ctx, invitation.CommunityID)
		if err != nil {
			return errorsapp.New(errorsapp.ErrCodeGroupChatNotFound)
		}

		if err := s.repo.AddCommunityMemberAndGroupChat(ctx, invitation.CommunityID, userID, groupChat.ID); err != nil {
			return err
		}

		s.repo.UpdateInvitationStatus(ctx, nil, invitationID, models.InvitationStatusAccepted)

		s.notifService.Create(ctx, userID, &invitation.InviterID,
			models.NotificationTypeCommunityInvitationAccepted,
			"Bạn đã tham gia cộng đồng theo lời mời",
			nil, &invitation.CommunityID, nil)
	} else {
		s.repo.UpdateInvitationStatus(ctx, nil, invitationID, models.InvitationStatusDeclined)
	}

	return nil
}
