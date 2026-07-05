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
	reportRepo          *repository.ReportRepository
	moderationRepo      *repository.ModerationRepository
	notificationService *NotificationService
}

func NewAdminService(authRepo *repository.AuthRepository, banRepo *repository.BanRepository, postRepo *repository.PostRepository, reportRepo *repository.ReportRepository, moderationRepo *repository.ModerationRepository, notificationService *NotificationService) *AdminService {
	return &AdminService{
		authRepo:            authRepo,
		banRepo:             banRepo,
		postRepo:            postRepo,
		reportRepo:          reportRepo,
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
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return err
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

	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return dto.AdminPostListResponse{}, err
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
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return err
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
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return err
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

func (s *AdminService) ListReports(ctx context.Context, superAdminID string, input dto.AdminReportFilterInput) (dto.AdminReportListResponse, error) {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return dto.AdminReportListResponse{}, err
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
		statusValue := models.ParseReportStatus(status)
		status = statusValue.String()
	}

	sortBy := strings.TrimSpace(strings.ToLower(input.SortBy))
	if sortBy != "created_at" && sortBy != "target_type" {
		sortBy = "created_at"
	}

	order := strings.TrimSpace(strings.ToLower(input.Order))
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	reports, err := s.reportRepo.ListAdminReports(ctx, input.Keyword, status, strings.TrimSpace(strings.ToLower(input.TargetType)), sortBy, order, pageSize, (page-1)*pageSize)
	if err != nil {
		return dto.AdminReportListResponse{}, err
	}

	total, err := s.reportRepo.CountAdminReports(ctx, input.Keyword, status, strings.TrimSpace(strings.ToLower(input.TargetType)))
	if err != nil {
		return dto.AdminReportListResponse{}, err
	}

	return dto.AdminReportListResponse{
		Reports:  reports,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AdminService) GetReportDetail(ctx context.Context, superAdminID, reportID string) (dto.AdminReportDetailResponse, error) {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return dto.AdminReportDetailResponse{}, err
	}

	report, err := s.reportRepo.FindByID(ctx, reportID)
	if err != nil {
		return dto.AdminReportDetailResponse{}, err
	}

	if report.Status == models.ReportStatusPending {
		if err := s.reportRepo.UpdateStatus(ctx, reportID, models.ReportStatusReviewed); err != nil {
			return dto.AdminReportDetailResponse{}, err
		}
		report.Status = models.ReportStatusReviewed
	}

	reporter, err := s.authRepo.FindByID(ctx, report.ReporterID)
	if err != nil {
		return dto.AdminReportDetailResponse{}, err
	}

	detail := dto.AdminReportDetailResponse{
		ID:               report.ID,
		ReporterID:       report.ReporterID,
		ReporterUsername: reporter.Username,
		ReporterEmail:    reporter.Email,
		TargetType:       "unknown",
		TargetUserID:     report.TargetUserID,
		TargetPostID:     report.TargetPostID,
		TargetCommentID:  report.TargetCommentID,
		ReportType:       report.ReportType,
		ViolationRuleID:  report.ViolationRuleID,
		ReasonDetail:     report.ReasonDetail,
		Status:           report.Status.String(),
		CreatedAt:        report.CreatedAt,
	}

	if report.TargetPostID != nil {
		post, err := s.postRepo.FindByID(ctx, *report.TargetPostID)
		if err == nil {
			detail.TargetType = "post"
			detail.PostOwnerID = &post.UserID
		} else {
			detail.TargetType = "post"
		}
	} else if report.TargetUserID != nil {
		detail.TargetType = "user"
	} else if report.TargetCommentID != nil {
		detail.TargetType = "comment"
	}

	return detail, nil
}

func (s *AdminService) ReviewReport(ctx context.Context, superAdminID, reportID string, input dto.AdminReportReviewInput) error {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return err
	}

	report, err := s.reportRepo.FindByID(ctx, reportID)
	if err != nil {
		return err
	}

	if report.Status != models.ReportStatusPending {
		return errors.New("báo cáo đã được xử lý")
	}

	action := strings.TrimSpace(strings.ToLower(input.Action))
	if action != "cancel" && action != "hide" && action != "ban" {
		return errors.New("action không hợp lệ, chỉ chấp nhận cancel, hide hoặc ban")
	}

	status := models.ReportStatusRejected
	if action == "hide" {
		if report.TargetPostID != nil {
			if err := s.postRepo.UpdateStatus(ctx, *report.TargetPostID, models.PostStatusHidden); err != nil {
				return fmt.Errorf("hide post: %w", err)
			}
			status = models.ReportStatusResolved

			moderation := models.NewModerationLog(superAdminID, models.ModerationActionDelete, models.ModerationTargetPost, *report.TargetPostID, input.Reason)
			moderation.ID = utils.GenerateUUID()
			moderation.CreatedAt = time.Now().UTC()
			if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
				return fmt.Errorf("create moderation log: %w", err)
			}
		} else if report.TargetUserID != nil {
			return errors.New("hide chỉ hỗ trợ với báo cáo bài viết; báo cáo người dùng cần logic xử lý khác")
		} else {
			return errors.New("loại báo cáo không được hỗ trợ")
		}
	}

	if action == "ban" {
		if report.TargetUserID == nil {
			return errors.New("ban chỉ hỗ trợ cho báo cáo người dùng")
		}

		if err := s.authRepo.UpdateUserStatus(ctx, *report.TargetUserID, models.UserStatusBanned); err != nil {
			return fmt.Errorf("ban user: %w", err)
		}

		ban := models.NewBan(*report.TargetUserID, superAdminID, input.Reason, nil)
		ban.ID = utils.GenerateUUID()
		ban.CreatedAt = time.Now().UTC()

		if err := s.banRepo.CreateBan(ctx, &ban); err != nil {
			return fmt.Errorf("create ban: %w", err)
		}

		status = models.ReportStatusResolved

		moderation := models.NewModerationLog(superAdminID, models.ModerationActionBan, models.ModerationTargetUser, *report.TargetUserID, input.Reason)
		moderation.ID = utils.GenerateUUID()
		moderation.CreatedAt = time.Now().UTC()

		if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
			return fmt.Errorf("create moderation log: %w", err)
		}
	}

	if err := s.reportRepo.UpdateStatus(ctx, reportID, status); err != nil {
		return err
	}

	reporterMessage := fmt.Sprintf("Báo cáo %s đã được xử lý bằng hành động: %s.", report.ID, action)
	_, _ = s.notificationService.Create(ctx, report.ReporterID, &superAdminID, models.NotificationTypeMessage, reporterMessage, report.TargetPostID, report.TargetUserID, report.TargetCommentID)

	if report.TargetPostID != nil {
		post, err := s.postRepo.FindByID(ctx, *report.TargetPostID)
		if err == nil {
			targetMessage := fmt.Sprintf("Bài viết của bạn đã bị báo cáo và đã được %s bởi quản trị viên.", action)
			_, _ = s.notificationService.Create(ctx, post.UserID, &superAdminID, models.NotificationTypeMessage, targetMessage, report.TargetPostID, nil, nil)
		}
	} else if report.TargetUserID != nil {
		targetMessage := fmt.Sprintf("Tài khoản của bạn đã bị báo cáo và đã được %s bởi quản trị viên.", action)
		_, _ = s.notificationService.Create(ctx, *report.TargetUserID, &superAdminID, models.NotificationTypeMessage, targetMessage, nil, report.TargetUserID, nil)
	}

	if report.TargetUserID != nil && action == "ban" {
		_, _ = s.notificationService.Create(
			ctx,
			*report.TargetUserID,
			&superAdminID,
			models.NotificationTypeMessage,
			"Tài khoản của bạn đã bị cấm vì vi phạm báo cáo.",
			nil,
			report.TargetUserID,
			nil,
		)
	}

	return nil
}

func (s *AdminService) ensureSuperAdmin(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("không có quyền truy cập")
	}
	isSuperAdmin, err := s.authRepo.HasRole(ctx, userID, models.RoleSuperAdmin)
	if err != nil {
		return fmt.Errorf("check superadmin: %w", err)
	}
	if !isSuperAdmin {
		return errors.New("chỉ có superadmin mới có được phép")
	}
	return nil
}

func reportTargetType(report *models.Report) string {
	if report.TargetPostID != nil {
		return "post"
	}
	if report.TargetUserID != nil {
		return "user"
	}
	if report.TargetCommentID != nil {
		return "comment"
	}
	return "unknown"
}
