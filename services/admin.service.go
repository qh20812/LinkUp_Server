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
	chatRepo            *repository.ChatRepository
	communityRepo       *repository.CommunityRepository
	profileRepo         *repository.ProfileRepository
	groupChatRepo       *repository.GroupChatRepository
	adminRepo           repository.AdminRepository
	notificationService *NotificationService
}

func NewAdminService(authRepo *repository.AuthRepository, banRepo *repository.BanRepository, postRepo *repository.PostRepository, reportRepo *repository.ReportRepository, moderationRepo *repository.ModerationRepository, chatRepo *repository.ChatRepository, communityRepo *repository.CommunityRepository, profileRepo *repository.ProfileRepository, groupChatRepo *repository.GroupChatRepository, adminRepo repository.AdminRepository, notificationService *NotificationService) *AdminService {
	return &AdminService{
		authRepo:            authRepo,
		banRepo:             banRepo,
		postRepo:            postRepo,
		reportRepo:          reportRepo,
		moderationRepo:      moderationRepo,
		chatRepo:            chatRepo,
		communityRepo:       communityRepo,
		profileRepo:         profileRepo,
		groupChatRepo:       groupChatRepo,
		adminRepo:           adminRepo,
		notificationService: notificationService,
	}
}

func (s *AdminService) GetDashboardAnalytics(ctx context.Context, superAdminID string, input dto.AdminAnalyticsFilterInput) (dto.AdminAnalyticsResponse, error) {
	// Kiểm tra quyền hạn SuperAdmin trước khi tính toán dữ liệu
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return dto.AdminAnalyticsResponse{}, err
	}

	// Xử lý bộ lọc thời gian mặc định (7 ngày gần nhất nếu bỏ trống)
	now := time.Now().UTC()
	if input.EndDate == "" {
		input.EndDate = now.Format("2006-01-02")
	}
	if input.StartDate == "" {
		input.StartDate = now.AddDate(0, 0, -7).Format("2006-01-02")
	}
	if input.Type == "" {
		input.Type = "all"
	}

	totalUsers, err := s.adminRepo.GetTotalUsers()
	if err != nil {
		return dto.AdminAnalyticsResponse{}, fmt.Errorf("lấy tổng số người dùng thất bại: %w", err)
	}

	totalPosts, err := s.adminRepo.GetTotalPosts()
	if err != nil {
		return dto.AdminAnalyticsResponse{}, fmt.Errorf("lấy tổng số bài viết thất bại: %w", err)
	}

	totalReports, err := s.adminRepo.GetTotalReports()
	if err != nil {
		return dto.AdminAnalyticsResponse{}, fmt.Errorf("lấy tổng số báo cáo thất bại: %w", err)
	}

	var chartData []dto.ChartDataPoint
	tableName := ""

	switch strings.ToLower(input.Type) {
	case "users":
		tableName = "users"
	case "posts":
		tableName = "posts"
	case "reports":
		tableName = "reports"
	case "all":
		tableName = "posts"
	default:
		tableName = "posts"
	}

	chartData, err = s.adminRepo.GetChartData(tableName, input.StartDate, input.EndDate)
	if err != nil {
		chartData = []dto.ChartDataPoint{} // Fallback mảng rỗng để không crash giao diện frontend
	}

	return dto.AdminAnalyticsResponse{
		TotalUsers:   totalUsers,
		TotalPosts:   totalPosts,
		TotalReports: totalReports,
		ChartData:    chartData,
		GeneratedAt:  time.Now().UTC(),
	}, nil
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

func (s *AdminService) BanUser(ctx context.Context, superAdminID, targetUserID string, input dto.AdminUserBanInput) (dto.AdminBanUserResponse, error) {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return dto.AdminBanUserResponse{}, err
	}

	targetUser, err := s.authRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return dto.AdminBanUserResponse{}, err
	}

	if targetUser.Status == models.UserStatusBanned {
		return dto.AdminBanUserResponse{}, fmt.Errorf("người dùng đã bị ban")
	}

	var expiresAt *time.Time
	durationKey := strings.TrimSpace(input.Duration)

	if durationKey != "permanent" {
		duration, ok := banDurationMap[durationKey]
		if !ok {
			return dto.AdminBanUserResponse{}, fmt.Errorf("thời hạn ban không hợp lệ")
		}
		t := time.Now().UTC().Add(duration)
		expiresAt = &t
	}

	ban := models.NewBan(targetUserID, superAdminID, input.Reason, expiresAt)
	ban.ID = utils.GenerateUUID()
	ban.CreatedAt = time.Now().UTC()

	if err := s.banRepo.CreateBan(ctx, &ban); err != nil {
		return dto.AdminBanUserResponse{}, err
	}

	s.transferOwnershipOnBan(ctx, targetUserID)

	if err := s.authRepo.UpdateUserStatus(ctx, targetUserID, models.UserStatusBanned); err != nil {
		return dto.AdminBanUserResponse{}, err
	}

	return dto.AdminBanUserResponse{
		Message: "ban user thành công",
		BanUtil: expiresAt,
	}, nil
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
		case string(models.PostStatusPublic),
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
	case
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

	if isReportFinalStatus(report.Status) {
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

		expiresAt, err := s.resolveBanExpiresAt(input.Duration)
		if err != nil {
			return err
		}

		if err := s.authRepo.UpdateUserStatus(ctx, *report.TargetUserID, models.UserStatusBanned); err != nil {
			return fmt.Errorf("ban user: %w", err)
		}

		ban := models.NewBan(*report.TargetUserID, superAdminID, input.Reason, expiresAt)
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

// ── Shared helpers ─────────────────────────────────────────────────────────

func (s *AdminService) ensureAdmin(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("không có quyền truy cập")
	}
	isSuperAdmin, err := s.authRepo.HasRole(ctx, userID, models.RoleSuperAdmin)
	if err != nil {
		return fmt.Errorf("check role: %w", err)
	}
	if isSuperAdmin {
		return nil
	}
	isAdmin, err := s.authRepo.HasRole(ctx, userID, models.RoleAdmin)
	if err != nil {
		return fmt.Errorf("check role: %w", err)
	}
	if isAdmin {
		return nil
	}
	return errors.New("chỉ có admin/superadmin mới được phép")
}

func (s *AdminService) resolvePageSize(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// ── Group Chat: Admin endpoints ─────────────────────────────────────────────

func (s *AdminService) ListGroups(ctx context.Context, userID string, input dto.AdminGroupFilterInput) (dto.AdminGroupListResponse, error) {
	if err := s.ensureAdmin(ctx, userID); err != nil {
		return dto.AdminGroupListResponse{}, err
	}

	page, pageSize := s.resolvePageSize(input.Page, input.PageSize)

	keyword := strings.TrimSpace(input.Keyword)
	status := strings.TrimSpace(strings.ToLower(input.Status))
	if status != "" {
		switch status {
		case string(models.ChatStatusActive), string(models.ChatStatusHidden), string(models.ChatStatusArchived):
		default:
			return dto.AdminGroupListResponse{}, fmt.Errorf("trạng thái group không hợp lệ")
		}
	}

	items, err := s.chatRepo.ListGroups(ctx, keyword, status, pageSize, (page-1)*pageSize)
	if err != nil {
		return dto.AdminGroupListResponse{}, err
	}

	total, err := s.chatRepo.CountGroups(ctx, keyword, status)
	if err != nil {
		return dto.AdminGroupListResponse{}, err
	}

	return dto.AdminGroupListResponse{
		Groups:   items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AdminService) GetGroupDetail(ctx context.Context, userID, chatID string) (dto.AdminGroupDetailResponse, error) {
	if err := s.ensureAdmin(ctx, userID); err != nil {
		return dto.AdminGroupDetailResponse{}, err
	}

	chat, err := s.chatRepo.FindChatByID(ctx, chatID)
	if err != nil {
		return dto.AdminGroupDetailResponse{}, fmt.Errorf("group không tồn tại")
	}
	if chat.Type != models.ChatTypeGroup {
		return dto.AdminGroupDetailResponse{}, fmt.Errorf("ID không phải là group chat")
	}

	members, err := s.chatRepo.GetGroupMembers(ctx, chatID)
	if err != nil {
		return dto.AdminGroupDetailResponse{}, err
	}

	creatorName := ""
	if chat.CreatorID != nil {
		creator, err := s.profileRepo.FindByUserID(ctx, *chat.CreatorID)
		if err == nil && creator != nil {
			creatorName = creator.DisplayName
		}
	}

	return dto.AdminGroupDetailResponse{
		ID:          chat.ID,
		Name:        chat.Name,
		AvatarURI:   chat.AvatarURI,
		CreatorID:   chat.CreatorID,
		CreatorName: creatorName,
		Type:        string(chat.Type),
		Status:      string(chat.Status),
		MemberCount: len(members),
		Members:     members,
		CreatedAt:   chat.CreatedAt,
	}, nil
}

func (s *AdminService) ListGroupMembers(ctx context.Context, userID, chatID string) ([]dto.AdminGroupMember, error) {
	if err := s.ensureAdmin(ctx, userID); err != nil {
		return nil, err
	}

	chat, err := s.chatRepo.FindChatByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("group không tồn tại")
	}
	if chat.Type != models.ChatTypeGroup {
		return nil, fmt.Errorf("ID không phải là group chat")
	}

	return s.chatRepo.GetGroupMembers(ctx, chatID)
}

func (s *AdminService) GetGroupModerationLogs(ctx context.Context, userID, chatID string, input dto.AdminGroupFilterInput) (dto.AdminModerationLogListResponse, error) {
	if err := s.ensureAdmin(ctx, userID); err != nil {
		return dto.AdminModerationLogListResponse{}, err
	}

	page, pageSize := s.resolvePageSize(input.Page, input.PageSize)

	logs, err := s.moderationRepo.ListLogsByTarget(ctx, models.ModerationTargetGroupChat, chatID, pageSize, (page-1)*pageSize)
	if err != nil {
		return dto.AdminModerationLogListResponse{}, err
	}

	total, err := s.moderationRepo.CountLogsByTarget(ctx, models.ModerationTargetGroupChat, chatID)
	if err != nil {
		return dto.AdminModerationLogListResponse{}, err
	}

	return dto.AdminModerationLogListResponse{
		Logs:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AdminService) moderateGroup(ctx context.Context, superAdminID, chatID string, action models.ModerationAction, logAction models.ModerationAction, newStatus models.ChatStatus, actionLabel string, reason string) error {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return err
	}

	chat, err := s.chatRepo.FindChatByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("group không tồn tại")
	}
	if chat.Type != models.ChatTypeGroup {
		return fmt.Errorf("ID không phải là group chat")
	}

	if chat.Status == models.ChatStatusArchived && newStatus != models.ChatStatusArchived {
		return fmt.Errorf("không thể thao tác trên group đã bị đình chỉ")
	}

	if chat.Status == newStatus {
		return fmt.Errorf("group đã ở trạng thái %s", actionLabel)
	}

	moderation := models.NewModerationLog(superAdminID, logAction, models.ModerationTargetGroupChat, chatID, reason)
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()
	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return err
	}

	if err := s.chatRepo.UpdateChatStatus(ctx, chatID, newStatus); err != nil {
		return err
	}

	if chat.CreatorID != nil {
		content := fmt.Sprintf("Nhóm chat '%s' đã bị %s bởi quản trị viên. Lý do: %s", chat.Name, actionLabel, reason)
		senderID := superAdminID
		_, _ = s.notificationService.Create(ctx, *chat.CreatorID, &senderID, models.NotificationTypeMessage, content, nil, nil, nil)
	}

	return nil
}

func (s *AdminService) HideGroup(ctx context.Context, superAdminID, chatID string, input dto.AdminModerateInput) error {
	return s.moderateGroup(ctx, superAdminID, chatID, models.ModerationActionHide, models.ModerationActionHide, models.ChatStatusHidden, "ẩn", input.Reason)
}

func (s *AdminService) UnhideGroup(ctx context.Context, superAdminID, chatID string) error {
	return s.moderateGroup(ctx, superAdminID, chatID, models.ModerationActionUpdate, models.ModerationActionUpdate, models.ChatStatusActive, "bỏ ẩn", "Bỏ ẩn group")
}

func (s *AdminService) ArchiveGroup(ctx context.Context, superAdminID, chatID string, input dto.AdminModerateInput) error {
	return s.moderateGroup(ctx, superAdminID, chatID, models.ModerationActionSuspend, models.ModerationActionSuspend, models.ChatStatusArchived, "đình chỉ", input.Reason)
}

func (s *AdminService) WarnGroup(ctx context.Context, superAdminID, chatID string, input dto.AdminWarnInput) error {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return err
	}

	chat, err := s.chatRepo.FindChatByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("group không tồn tại")
	}
	if chat.Type != models.ChatTypeGroup {
		return fmt.Errorf("ID không phải là group chat")
	}

	moderation := models.NewModerationLog(superAdminID, models.ModerationActionWarn, models.ModerationTargetGroupChat, chatID, input.Reason)
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()
	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return err
	}

	if chat.CreatorID != nil {
		content := fmt.Sprintf("Cảnh báo nhóm chat '%s': %s", chat.Name, input.Message)
		senderID := superAdminID
		_, _ = s.notificationService.Create(ctx, *chat.CreatorID, &senderID, models.NotificationTypeMessage, content, nil, nil, nil)
	}

	return nil
}

// ── Community: Admin endpoints ──────────────────────────────────────────────

func (s *AdminService) ListCommunities(ctx context.Context, userID string, input dto.AdminCommunityFilterInput) (dto.AdminCommunityListResponse, error) {
	if err := s.ensureAdmin(ctx, userID); err != nil {
		return dto.AdminCommunityListResponse{}, err
	}

	page, pageSize := s.resolvePageSize(input.Page, input.PageSize)

	keyword := strings.TrimSpace(input.Keyword)
	status := strings.TrimSpace(strings.ToLower(input.Status))
	if status != "" {
		switch status {
		case string(models.CommunityStatusActive), string(models.CommunityStatusHidden), string(models.CommunityStatusArchived):
		default:
			return dto.AdminCommunityListResponse{}, fmt.Errorf("trạng thái community không hợp lệ")
		}
	}

	privacy := strings.TrimSpace(strings.ToLower(input.Privacy))
	if privacy != "" {
		switch privacy {
		case string(models.PrivacyPublic), string(models.PrivacyCode), string(models.PrivacyInvitationOnly):
		default:
			return dto.AdminCommunityListResponse{}, fmt.Errorf("quyền riêng tư không hợp lệ")
		}
	}

	items, err := s.communityRepo.ListCommunitiesAdmin(ctx, keyword, status, privacy, pageSize, (page-1)*pageSize)
	if err != nil {
		return dto.AdminCommunityListResponse{}, err
	}

	total, err := s.communityRepo.CountCommunitiesAdmin(ctx, keyword, status, privacy)
	if err != nil {
		return dto.AdminCommunityListResponse{}, err
	}

	return dto.AdminCommunityListResponse{
		Communities: items,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

func (s *AdminService) GetCommunityDetail(ctx context.Context, userID, communityID string) (dto.AdminCommunityDetailResponse, error) {
	if err := s.ensureAdmin(ctx, userID); err != nil {
		return dto.AdminCommunityDetailResponse{}, err
	}

	community, err := s.communityRepo.FindByID(ctx, communityID)
	if err != nil {
		return dto.AdminCommunityDetailResponse{}, fmt.Errorf("cộng đồng không tồn tại")
	}

	members, err := s.communityRepo.FindCommunityMembersWithProfiles(ctx, communityID)
	if err != nil {
		return dto.AdminCommunityDetailResponse{}, err
	}

	creatorName := ""
	creator, err := s.profileRepo.FindByUserID(ctx, community.CreatorID)
	if err == nil && creator != nil {
		creatorName = creator.DisplayName
	}

	return dto.AdminCommunityDetailResponse{
		ID:          community.ID,
		Name:        community.Name,
		Description: community.Description,
		AvatarURI:   community.AvatarURI,
		CreatorID:   community.CreatorID,
		CreatorName: creatorName,
		Privacy:     string(community.Privacy),
		Status:      string(community.Status),
		AutoApprove: community.AutoApprove,
		MemberCount: len(members),
		Members:     members,
		CreatedAt:   community.CreatedAt,
		UpdatedAt:   community.UpdatedAt,
	}, nil
}

func (s *AdminService) ListCommunityMembers(ctx context.Context, userID, communityID string) ([]dto.AdminCommunityMember, error) {
	if err := s.ensureAdmin(ctx, userID); err != nil {
		return nil, err
	}

	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, fmt.Errorf("cộng đồng không tồn tại")
	}

	return s.communityRepo.FindCommunityMembersWithProfiles(ctx, communityID)
}

func (s *AdminService) GetCommunityModerationLogs(ctx context.Context, userID, communityID string, input dto.AdminGroupFilterInput) (dto.AdminModerationLogListResponse, error) {
	if err := s.ensureAdmin(ctx, userID); err != nil {
		return dto.AdminModerationLogListResponse{}, err
	}

	page, pageSize := s.resolvePageSize(input.Page, input.PageSize)

	logs, err := s.moderationRepo.ListLogsByTarget(ctx, models.ModerationTargetCommunity, communityID, pageSize, (page-1)*pageSize)
	if err != nil {
		return dto.AdminModerationLogListResponse{}, err
	}

	total, err := s.moderationRepo.CountLogsByTarget(ctx, models.ModerationTargetCommunity, communityID)
	if err != nil {
		return dto.AdminModerationLogListResponse{}, err
	}

	return dto.AdminModerationLogListResponse{
		Logs:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AdminService) moderateCommunity(ctx context.Context, superAdminID, communityID string, action models.ModerationAction, logAction models.ModerationAction, newStatus models.CommunityStatus, actionLabel string, reason string) error {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return err
	}

	community, err := s.communityRepo.FindByID(ctx, communityID)
	if err != nil {
		return fmt.Errorf("cộng đồng không tồn tại")
	}

	if community.Status == models.CommunityStatusArchived && newStatus != models.CommunityStatusArchived {
		return fmt.Errorf("không thể thao tác trên cộng đồng đã bị đình chỉ")
	}

	if community.Status == newStatus {
		return fmt.Errorf("cộng đồng đã ở trạng thái %s", actionLabel)
	}

	moderation := models.NewModerationLog(superAdminID, logAction, models.ModerationTargetCommunity, communityID, reason)
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()
	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return err
	}

	if err := s.communityRepo.UpdateStatus(ctx, communityID, newStatus); err != nil {
		return err
	}

	content := fmt.Sprintf("Cộng đồng '%s' đã bị %s bởi quản trị viên. Lý do: %s", community.Name, actionLabel, reason)
	senderID := superAdminID
	_, _ = s.notificationService.Create(ctx, community.CreatorID, &senderID, models.NotificationTypeMessage, content, nil, nil, nil)

	return nil
}

func (s *AdminService) HideCommunity(ctx context.Context, superAdminID, communityID string, input dto.AdminModerateInput) error {
	return s.moderateCommunity(ctx, superAdminID, communityID, models.ModerationActionHide, models.ModerationActionHide, models.CommunityStatusHidden, "ẩn", input.Reason)
}

func (s *AdminService) UnhideCommunity(ctx context.Context, superAdminID, communityID string) error {
	return s.moderateCommunity(ctx, superAdminID, communityID, models.ModerationActionUpdate, models.ModerationActionUpdate, models.CommunityStatusActive, "bỏ ẩn", "Bỏ ẩn cộng đồng")
}

func (s *AdminService) ArchiveCommunity(ctx context.Context, superAdminID, communityID string, input dto.AdminModerateInput) error {
	return s.moderateCommunity(ctx, superAdminID, communityID, models.ModerationActionSuspend, models.ModerationActionSuspend, models.CommunityStatusArchived, "đình chỉ", input.Reason)
}

func (s *AdminService) WarnCommunity(ctx context.Context, superAdminID, communityID string, input dto.AdminWarnInput) error {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return err
	}

	community, err := s.communityRepo.FindByID(ctx, communityID)
	if err != nil {
		return fmt.Errorf("cộng đồng không tồn tại")
	}

	moderation := models.NewModerationLog(superAdminID, models.ModerationActionWarn, models.ModerationTargetCommunity, communityID, input.Reason)
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()
	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return err
	}

	content := fmt.Sprintf("Cảnh báo cộng đồng '%s': %s", community.Name, input.Message)
	senderID := superAdminID
	_, _ = s.notificationService.Create(ctx, community.CreatorID, &senderID, models.NotificationTypeMessage, content, nil, nil, nil)

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

func isReportFinalStatus(status models.ReportStatus) bool {
	switch status {
	case models.ReportStatusRejected, models.ReportStatusResolved:
		return true
	default:
		return false
	}
}

func (s *AdminService) resolveBanExpiresAt(durationKey string) (*time.Time, error) {
	durationKey = strings.TrimSpace(durationKey)
	if durationKey == "" || strings.EqualFold(durationKey, "permanent") {
		return nil, nil
	}

	duration, ok := banDurationMap[durationKey]
	if !ok {
		return nil, fmt.Errorf("thời hạn ban không hợp lệ")
	}

	t := time.Now().UTC().Add(duration)
	return &t, nil
}

// ── Auto-transfer ownership on ban ──────────────────────────────────────────

func (s *AdminService) transferOwnershipOnBan(ctx context.Context, targetUserID string) {
	// Transfer community ownership
	communities, err := s.communityRepo.FindByCreator(ctx, targetUserID)
	if err != nil {
		return
	}
	for _, c := range communities {
		nextOwnerID := s.findNextOwnerForCommunity(ctx, c.ID, targetUserID)
		if nextOwnerID == "" {
			continue
		}
		_ = s.communityRepo.TransferCommunityOwnership(ctx, c.ID, targetUserID, nextOwnerID, false)
	}

	// Transfer group chat ownership
	groupChats, err := s.chatRepo.FindGroupChatsByCreator(ctx, targetUserID)
	if err != nil {
		return
	}
	for _, g := range groupChats {
		nextAdminID := s.findNextAdminForGroupChat(ctx, g.ID, targetUserID)
		if nextAdminID == "" {
			continue
		}
		_ = s.groupChatRepo.TransferOwnership(ctx, g.ID, targetUserID, nextAdminID, false, time.Now().UTC())
	}
}

func (s *AdminService) findNextOwnerForCommunity(ctx context.Context, communityID, excludeUserID string) string {
	admins, err := s.communityRepo.FindCommunityAdmins(ctx, communityID)
	if err == nil {
		for _, id := range admins {
			if id != excludeUserID {
				return id
			}
		}
	}

	member, err := s.communityRepo.FindOldestMember(ctx, communityID, excludeUserID)
	if err == nil && member != nil {
		return member.UserID
	}

	mc, err := s.communityRepo.FindHighestContributionMember(ctx, communityID, excludeUserID)
	if err == nil && mc != nil {
		return mc.UserID
	}

	return ""
}

func (s *AdminService) findNextAdminForGroupChat(ctx context.Context, chatID, excludeUserID string) string {
	admins, err := s.groupChatRepo.FindAllAdmins(ctx, chatID, excludeUserID)
	if err == nil && len(admins) > 0 {
		return admins[0]
	}

	member, err := s.groupChatRepo.FindOldestMember(ctx, chatID, excludeUserID)
	if err == nil && member != nil {
		return member.UserID
	}

	allIDs, err := s.chatRepo.GetParticipantIDs(ctx, chatID)
	if err == nil {
		for _, id := range allIDs {
			if id != excludeUserID {
				return id
			}
		}
	}

	return ""
}

// ── Delete Group / Delete Community ─────────────────────────────────────────

func (s *AdminService) DeleteGroup(ctx context.Context, superAdminID, chatID string, input dto.AdminModerateInput) error {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return err
	}

	chat, err := s.chatRepo.FindChatByID(ctx, chatID)
	if err != nil {
		return err
	}
	if chat.Type != models.ChatTypeGroup {
		return fmt.Errorf("chỉ có thể xóa group chat")
	}

	participantIDs, err := s.chatRepo.GetParticipantIDs(ctx, chatID)
	if err != nil {
		return err
	}

	if len(participantIDs) > 1 {
		return fmt.Errorf("không thể xóa group chat còn thành viên khác; hãy chuyển quyền sở hữu trước")
	}

	moderation := models.NewModerationLog(superAdminID, models.ModerationActionDelete, models.ModerationTargetGroupChat, chatID, input.Reason)
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()
	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return err
	}

	return s.chatRepo.DeleteChat(ctx, chatID)
}

func (s *AdminService) DeleteCommunity(ctx context.Context, superAdminID, communityID string, input dto.AdminModerateInput) error {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return err
	}

	community, err := s.communityRepo.FindByID(ctx, communityID)
	if err != nil {
		return fmt.Errorf("cộng đồng không tồn tại")
	}
	if community.CreatorID == "" {
		return fmt.Errorf("cộng đồng không có người tạo")
	}

	memberIDs, err := s.communityRepo.FindCommunityMemberIDs(ctx, communityID)
	if err != nil {
		return err
	}

	if len(memberIDs) > 1 {
		return fmt.Errorf("không thể xóa cộng đồng còn thành viên khác; hãy chuyển quyền sở hữu trước")
	}

	moderation := models.NewModerationLog(superAdminID, models.ModerationActionDelete, models.ModerationTargetCommunity, communityID, input.Reason)
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()
	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return err
	}

	return s.communityRepo.RemoveMember(ctx, communityID, community.CreatorID)
}
