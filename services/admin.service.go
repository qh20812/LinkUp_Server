package services

import (
	"context"
	"errors"
	"fmt"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"strings"
	"time"
)

var banDurationMap = map[string]time.Duration{
	"1m":  time.Minute,
	"30m": 30 * time.Minute,
	"1d":  24 * time.Hour,
	"3d":  3 * 24 * time.Hour,
	"1w":  7 * 24 * time.Hour,
	"2w":  14 * 24 * time.Hour,
	"1M":  30 * 24 * time.Hour,
	"3M":  90 * 24 * time.Hour,
	"6M":  180 * 24 * time.Hour,
	"9M":  270 * 24 * time.Hour,
	"1y":  365 * 24 * time.Hour,
}

type AdminService struct {
	authRepo            *repository.AuthRepository
	banRepo             *repository.BanRepository
	postRepo            *repository.PostRepository
	moderationRepo      *repository.ModerationRepository
	notificationService *NotificationService
}

func NewAdminService(authRepo *repository.AuthRepository, banRepo *repository.BanRepository, postRepo *repository.PostRepository, moderationRepo *repository.ModerationRepository, notificationService *NotificationService) *AdminService {
	return &AdminService{
		authRepo:            authRepo,
		banRepo:             banRepo,
		postRepo:            postRepo,
		moderationRepo:      moderationRepo,
		notificationService: notificationService,
	}
}

func (s *AdminService) ListUsers(ctx context.Context, input dto.AdminUserFilterInput) (dto.AdminUserListResponse, error) {
	page := input.Page
	if page <= 0 {
		page = 1
	}

	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	statusFilter := strings.TrimSpace(strings.ToLower(input.Status))
	if statusFilter != "" {
		switch statusFilter {
		case string(models.UserStatusActive), string(models.UserStatusBanned), string(models.UserStatusSuspended):
		default:
			return dto.AdminUserListResponse{}, fmt.Errorf("trạng thái không hợp lệ")
		}
	}

	users, err := s.authRepo.ListUsers(ctx, input.Keyword, statusFilter, pageSize, (page-1)*pageSize)
	if err != nil {
		return dto.AdminUserListResponse{}, err
	}

	total, err := s.authRepo.CountUsers(ctx, input.Keyword, statusFilter)
	if err != nil {
		return dto.AdminUserListResponse{}, err
	}

	resp := dto.AdminUserListResponse{
		Users:    users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	if len(users) == 0 {
		resp.Message = "Không tìm thấy người dùng"
	}
	return resp, nil
}

func (s *AdminService) UpdateUserStatus(ctx context.Context, superAdminID, targetUserID string, input dto.AdminUserUpdateStatusInput) error {
	isSuperAdmin, err := s.authRepo.HasRole(ctx, superAdminID, models.RoleSuperAdmin)
	if err != nil {
		return err
	}
	if !isSuperAdmin {
		return fmt.Errorf("chỉ superadmin mới có quyền")
	}

	targetUser, err := s.authRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return err
	}

	isTargetSuperAdmin, err := s.authRepo.HasRole(ctx, targetUserID, models.RoleSuperAdmin)
	if err != nil {
		return err
	}
	if isTargetSuperAdmin {
		return fmt.Errorf("Không thể chỉnh sửa trạng thái của superadmin")
	}

	statusValue := strings.TrimSpace(strings.ToLower(input.Status))
	var status models.UserStatus
	switch statusValue {
	case string(models.UserStatusActive):
		status = models.UserStatusActive
	case string(models.UserStatusBanned):
		status = models.UserStatusBanned
	case string(models.UserStatusSuspended):
		status = models.UserStatusSuspended
	default:
		return fmt.Errorf("trạng thái không hợp lệ")
	}

	if targetUser.Status == status {
		return nil
	}

	return s.authRepo.UpdateUserStatus(ctx, targetUserID, status)
}

func (s *AdminService) BanUser(ctx context.Context, superAdminID, targetUserID string, input dto.AdminUserBanInput) error {
	isSuperAdmin, err := s.authRepo.HasRole(ctx, superAdminID, models.RoleSuperAdmin)
	if err != nil {
		return err
	}
	if !isSuperAdmin {
		return fmt.Errorf("chỉ superadmin mới có quyền")
	}

	targetUser, err := s.authRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return err
	}

	if targetUser.Status == models.UserStatusBanned {
		return fmt.Errorf("người dùng đã bị ban")
	}

	var expiresAt *time.Time
	durationKey := strings.TrimSpace(input.Duration)

	if durationKey != "permanent" {
		duration, ok := banDurationMap[durationKey]
		if !ok {
			return fmt.Errorf("thời hạn ban không hợp lệ")
		}
		t := time.Now().UTC().Add(duration)
		expiresAt = &t
	}

	ban := models.NewBan(targetUserID, superAdminID, input.Reason, expiresAt)
	ban.ID = utils.GenerateUUID()
	ban.CreatedAt = time.Now().UTC()

	if err := s.banRepo.CreateBan(ctx, &ban); err != nil {
		return err
	}

	return s.authRepo.UpdateUserStatus(ctx, targetUserID, models.UserStatusBanned)
}

func (s *AdminService) ListPosts(ctx context.Context, superAdminID string, input dto.AdminPostFilterInput) (dto.AdminPostListResponse, error) {
	if superAdminID == "" {
		return dto.AdminPostListResponse{}, errors.New("không có quyền truy cập")
	}

	isSuperAdmin, err := s.authRepo.HasRole(ctx, superAdminID, models.RoleSuperAdmin)
	if err != nil {
		return dto.AdminPostListResponse{}, err
	}
	if !isSuperAdmin {
		return dto.AdminPostListResponse{}, errors.New("chỉ superadmin mới có quyền")
	}

	page := input.Page
	if page <= 0 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	status := strings.TrimSpace(strings.ToLower(input.Status))
	if status != "" {
		switch status {
		case string(models.PostStatusActive),
			string(models.PostStatusPublic),
			string(models.PostStatusPrivate),
			string(models.PostStatusHidden),
			string(models.PostStatusFriend),
			string(models.PostStatusDeleted):
		default:
			return dto.AdminPostListResponse{}, fmt.Errorf("trạng thái bài viết không hợp lệ")
		}
	}

	posts, err := s.postRepo.ListPosts(ctx, input.Keyword, status, pageSize, (page-1)*pageSize)
	if err != nil {
		return dto.AdminPostListResponse{}, err
	}

	total, err := s.postRepo.CountPosts(ctx, input.Keyword, status)
	if err != nil {
		return dto.AdminPostListResponse{}, err
	}

	items := make([]dto.AdminPostListItem, 0, len(posts))
	for _, p := range posts {
		items = append(items, dto.AdminPostListItem{
			ID:            p.ID,
			UserID:        p.UserID,
			Title:         p.Title,
			Content:       p.Content,
			Status:        string(p.Status),
			ViewsCount:    p.ViewsCount,
			LikesCount:    p.LikesCount,
			CommentsCount: p.CommentsCount,
			SharesCount:   p.SharesCount,
			CreatedAt:     p.CreatedAt,
			UpdatedAt:     p.UpdatedAt,
		})
	}

	resp := dto.AdminPostListResponse{
		Posts:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	if len(items) == 0 {
		resp.Message = "Không tìm thấy bài viết"
	}
	return resp, nil
}

func (s *AdminService) HidePost(ctx context.Context, superAdminID, postID string, input dto.AdminHidePostInput) error {
	if superAdminID == "" {
		return errors.New("không có quyền truy cập")
	}

	isSuperAdmin, err := s.authRepo.HasRole(ctx, superAdminID, models.RoleSuperAdmin)
	if err != nil {
		return err
	}
	if !isSuperAdmin {
		return errors.New("chỉ superadmin mới có quyền")
	}

	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("bài viết không tồn tại")
	}

	if post.Status == models.PostStatusHidden {
		return errors.New("bài viết đã ở trạng thái ẩn")
	}

	moderation := models.NewModerationLog(superAdminID, models.ModerationActionDelete, models.ModerationTargetPost, postID, input.Reason)
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()

	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return err
	}

	if err := s.postRepo.UpdateStatus(ctx, postID, models.PostStatusHidden); err != nil {
		return err
	}

	senderID := superAdminID
	postIDPtr := postID
	_, err = s.notificationService.Create(
		ctx,
		post.UserID,
		&senderID,
		models.NotificationTypeMessage,
		"Bài viết của bạn đã bị ẩn vì: "+input.Reason,
		&postIDPtr,
		nil,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *AdminService) ChangePostStatus(ctx context.Context, superAdminID, postID string, input dto.AdminUpdatePostStatusInput) error {
	isSuperAdmin, err := s.authRepo.HasRole(ctx, superAdminID, models.RoleSuperAdmin)
	if err != nil {
		return err
	}
	if !isSuperAdmin {
		return fmt.Errorf("chỉ superadmin mới có quyền")
	}

	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("bài viết không tồn tại")
	}

	statusValue := strings.TrimSpace(strings.ToLower(input.Status))
	switch statusValue {
	case string(models.PostStatusActive),
		string(models.PostStatusPublic),
		string(models.PostStatusPrivate),
		string(models.PostStatusHidden),
		string(models.PostStatusFriend),
		string(models.PostStatusDeleted):
	default:
		return fmt.Errorf("trạng thái bài viết không hợp lệ")
	}

	newStatus := models.ParsePostStatus(statusValue)
	if post.Status == newStatus {
		return nil
	}

	moderation := models.NewModerationLog(
		superAdminID,
		models.ModerationActionUpdate,
		models.ModerationTargetPost,
		postID,
		fmt.Sprintf("Cập nhật trạng thái thành %s", newStatus),
	)
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()

	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return err
	}

	if err := s.postRepo.UpdateStatus(ctx, postID, newStatus); err != nil {
		return err
	}

	senderID := superAdminID
	postIDPtr := postID
	_, err = s.notificationService.Create(
		ctx,
		post.UserID,
		&senderID,
		models.NotificationTypeMessage,
		fmt.Sprintf("Bài viết của bạn đã được cập nhật trạng thái thành %s", newStatus),
		&postIDPtr,
		nil,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}
