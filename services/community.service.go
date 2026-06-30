package services

import (
	"context"
	"errors"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"mime/multipart"
	"time"
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

func (s *CommunityService) CreateCommunity(ctx context.Context, creatorID, name, description, avatarURI string) (*models.Community, error) {
	if err := s.validation.ValidateCreateCommunity(name, description, avatarURI); err != nil {
		return nil, err
	}

	isAdmin, err := s.authRepo.HasRole(ctx, creatorID, models.RoleAdmin)
	if err != nil {
		return nil, errors.New("lỗi khi kiểm tra quyền người dùng")
	}
	isSuperAdmin, err := s.authRepo.HasRole(ctx, creatorID, models.RoleSuperAdmin)
	if err != nil {
		return nil, errors.New("lỗi khi kiểm tra quyền người dùng")
	}
	if isAdmin || isSuperAdmin {
		return nil, errors.New("quản trị viên không được tạo cộng đồng")
	}

	creator, err := s.authRepo.FindByID(ctx, creatorID)
	if err != nil {
		return nil, errors.New("người dùng không tồn tại")
	}
	if !creator.IsActive() {
		return nil, errors.New("tài khoản chưa được kích hoạt, không thể tạo cộng đồng")
	}

	taken, err := s.repo.IsNameTaken(ctx, name)
	if err != nil {
		return nil, errors.New("lỗi khi kiểm tra tên cộng đồng")
	}
	if taken {
		return nil, validations.ErrCommunityNameExists
	}

	now := time.Now().UTC()
	community := models.NewCommunity(creatorID, name, description, avatarURI)
	community.ID = utils.GenerateUUID()
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
		return nil, err
	}
	if err := s.repo.FindRoleByName(ctx, models.RoleGroupAdmin, &groupAdminRole); err != nil {
		return nil, err
	}
	userRoles[0].RoleID = communityAdminRole.ID
	userRoles[1].RoleID = groupAdminRole.ID

	if err := s.repo.CreateWithRoles(ctx, &community, &adminMember, userRoles); err != nil {
		return nil, err
	}
	return &community, nil
}

func (s *CommunityService) SetCommunityBackground(ctx context.Context, userID, communityID string, file *multipart.FileHeader) error {
	if err := s.groupRole.RequireRole(ctx, communityID, userID, models.GroupRoleAdmin); err != nil {
		return err
	}

	community, err := s.repo.FindByID(ctx, communityID)
	if err != nil {
		return errors.New("cộng đồng không tồn tại")
	}

	user, err := s.authRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("người dùng không tồn tại")
	}
	if !user.IsActive() {
		return errors.New("tài khoản chưa được kích hoạt")
	}

	media, err := s.mediaService.UploadMedia(ctx, userID, file)
	if err != nil {
		return errors.New("tải ảnh background thất bại")
	}

	if err := s.validation.ValidateBackgroundURI(media.FileURI); err != nil {
		return err
	}

	if err := s.repo.UpdateBackground(ctx, community.ID, media.FileURI); err != nil {
		return errors.New("cập nhật background cộng đồng thất bại")
	}

	return nil
}

func (s *CommunityService) RequestJoin(ctx context.Context, userID, communityID string) (string, error) {
	user, err := s.authRepo.FindByID(ctx, userID)
	if err != nil {
		return "", errors.New("người dùng không tồn tại")
	}
	if !user.IsActive() {
		return "", errors.New("tài khoản chưa được kích hoạt")
	}

	community, err := s.repo.FindByID(ctx, communityID)
	if err != nil {
		return "", validations.ErrCommunityNotFound
	}

	isMember, err := s.repo.IsUserMember(ctx, communityID, userID)
	if err != nil {
		return "", errors.New("lỗi khi kiểm tra thành viên")
	}
	if isMember {
		return "", validations.ErrAlreadyMember
	}

	existing, err := s.repo.FindPendingJoinRequestByUserAndCommunity(ctx, communityID, userID)
	if err == nil && existing != nil {
		return "", validations.ErrJoinRequestPending
	}

	now := time.Now().UTC()
	joinReq := models.NewCommunityJoinRequest(communityID, userID)
	joinReq.ID = utils.GenerateUUID()
	joinReq.CreatedAt = now

	if err := s.repo.CreateJoinRequest(ctx, &joinReq); err != nil {
		return "", errors.New("gửi yêu cầu tham gia thất bại")
	}

	s.notifService.Create(ctx, community.CreatorID, &userID, models.NotificationTypeCommunityJoinRequest, "đã gửi yêu cầu tham gia cộng đồng", nil, &userID, nil)

	return joinReq.ID, nil
}

func (s *CommunityService) ListPendingRequests(ctx context.Context, adminID, communityID string) (dto.JoinRequestListResponse, error) {
	if err := s.groupRole.RequireRole(ctx, communityID, adminID, models.GroupRoleAdmin); err != nil {
		return dto.JoinRequestListResponse{}, validations.ErrNotCommunityAdmin
	}

	requests, err := s.repo.FindPendingJoinRequestsByCommunity(ctx, communityID)
	if err != nil {
		return dto.JoinRequestListResponse{}, errors.New("lỗi khi lấy danh sách yêu cầu")
	}

	items := make([]dto.JoinRequestItem, 0, len(requests))
	for _, req := range requests {
		profile, err := s.profileRepo.FindByUserID(ctx, req.UserID)
		displayName := ""
		avatarURI := ""
		if err == nil {
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

	if err := s.repo.ApproveJoinRequest(ctx, requestID); err != nil {
		return err
	}

	s.notifService.Create(ctx, req.UserID, &adminID, models.NotificationTypeCommunityJoinApproved, "đã chấp nhận yêu cầu tham gia cộng đồng", nil, &adminID, nil)

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
		return errors.New("lỗi khi kiểm tra thông tin thành viên")
	}
	if isCreator {
		return validations.ErrCannotTargetAdmin
	}

	newRoleName := models.RoleName(newRole)
	if err := s.repo.UpdateUserRole(ctx, communityID, memberID, newRoleName); err != nil {
		return errors.New("cập nhật vai trò thất bại")
	}

	s.notifService.Create(ctx, memberID, &adminID, models.NotificationTypeCommunityRoleChanged, "đã thay đổi vai trò của bạn trong cộng đồng", nil, &adminID, nil)

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
		return errors.New("lỗi khi kiểm tra thành viên")
	}
	if !isMember {
		return validations.ErrMemberNotFound
	}

	isCreator, err := s.repo.IsUserCreator(ctx, communityID, userID)
	if err != nil {
		return errors.New("lỗi khi kiểm tra thông tin thành viên")
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
