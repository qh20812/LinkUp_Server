package services

import (
	"context"
	"fmt"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"log"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
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
	mediaRepo           *repository.MediaRepository
	adRepo              repository.AdRepository
	notificationService *NotificationService
	cloudinary          *cloudinary.Cloudinary
}

func NewAdminService(authRepo *repository.AuthRepository, banRepo *repository.BanRepository, postRepo *repository.PostRepository, reportRepo *repository.ReportRepository, moderationRepo *repository.ModerationRepository, chatRepo *repository.ChatRepository, communityRepo *repository.CommunityRepository, profileRepo *repository.ProfileRepository, groupChatRepo *repository.GroupChatRepository, adminRepo repository.AdminRepository, mediaRepo *repository.MediaRepository, adRepo repository.AdRepository, notificationService *NotificationService) *AdminService {
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
		mediaRepo:           mediaRepo,
		adRepo:              adRepo,
		notificationService: notificationService,
	}
}

// SetCloudinary gán Cloudinary client cho admin service (tránh breaking constructor).
func (s *AdminService) SetCloudinary(cld *cloudinary.Cloudinary) {
	s.cloudinary = cld
}

func (s *AdminService) GetDashboardAnalytics(ctx context.Context, adminID string, input dto.AdminAnalyticsFilterInput) (dto.AdminAnalyticsResponse, error) {
	if err := s.ensureAdmin(ctx, adminID); err != nil {
		return dto.AdminAnalyticsResponse{}, err
	}

	now := time.Now().UTC()
	if input.EndDate == "" {
		input.EndDate = now.Format("2006-01-02")
	}
	if input.StartDate == "" {
		input.StartDate = now.AddDate(0, 0, -7).Format("2006-01-02")
	}

	// ── Error-checked queries (concurrent, stop on first error) ──
	var (
		totalUsers   int64
		totalPosts   int64
		totalReports int64
	)
	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error { var err error; totalUsers, err = s.adminRepo.GetTotalUsers(); return err })
	g.Go(func() error { var err error; totalPosts, err = s.adminRepo.GetTotalPosts(); return err })
	g.Go(func() error { var err error; totalReports, err = s.adminRepo.GetTotalReports(); return err })
	if err := g.Wait(); err != nil {
		return dto.AdminAnalyticsResponse{}, fmt.Errorf("lấy thống kê thất bại: %w", err)
	}

	// ── Unchecked queries (concurrent, errors silently ignored) ──
	var (
		totalComments     int64
		totalMedia        int64
		totalGroups       int64
		totalCommunities  int64
		activeBanCount    int64
		pendingReports    int64
		flaggedMediaCount int64
		activeUsersToday  int64
		totalLikes        int64
		totalShares       int64
		topUsers          []dto.TopActiveUser
		topPosts          []dto.TopEngagedPost
		userDist          []dto.StatusCount
		reportDist        []dto.StatusCount
		chartDataUsers    []dto.ChartDataPoint
		chartDataPosts    []dto.ChartDataPoint
		chartDataReports  []dto.ChartDataPoint
		chartDataComments []dto.ChartDataPoint
	)

	oneMonthAgo := now.AddDate(0, -1, 0)

	var wg sync.WaitGroup
	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}

	run(func() { totalComments, _ = s.adminRepo.GetTotalComments() })
	run(func() { totalMedia, _ = s.adminRepo.GetTotalMedia() })
	run(func() { totalGroups, _ = s.adminRepo.GetTotalGroups() })
	run(func() { totalCommunities, _ = s.adminRepo.GetTotalCommunities() })
	run(func() { activeBanCount, _ = s.adminRepo.GetActiveBanCount() })
	run(func() { pendingReports, _ = s.adminRepo.GetPendingReportCount() })
	run(func() { flaggedMediaCount, _ = s.adminRepo.GetFlaggedMediaCount() })
	run(func() { activeUsersToday, _ = s.adminRepo.GetActiveUsersToday() })
	run(func() { totalLikes, _ = s.adminRepo.GetTotalLikes() })
	run(func() { totalShares, _ = s.adminRepo.GetTotalShares() })
	run(func() { topUsers, _ = s.adminRepo.GetTopActiveUsers(5) })
	run(func() { topPosts, _ = s.adminRepo.GetTopEngagedPosts(5) })
	run(func() { userDist, _ = s.adminRepo.GetUserStatusDistribution() })
	run(func() { reportDist, _ = s.adminRepo.GetReportStatusDistribution() })
	run(func() { chartDataUsers, _ = s.adminRepo.GetChartData("users", input.StartDate, input.EndDate) })
	run(func() { chartDataPosts, _ = s.adminRepo.GetChartData("posts", input.StartDate, input.EndDate) })
	run(func() { chartDataReports, _ = s.adminRepo.GetChartData("reports", input.StartDate, input.EndDate) })
	run(func() { chartDataComments, _ = s.adminRepo.GetChartData("comments", input.StartDate, input.EndDate) })

	// Historical counts: single UNION ALL query instead of 7 individual
	var prevCounts map[string]int64
	run(func() { prevCounts, _ = s.adminRepo.GetCountsBeforeDate(oneMonthAgo) })

	wg.Wait()

	usersChangePct := calcPercentChange(totalUsers, prevCounts["users"])
	postsChangePct := calcPercentChange(totalPosts, prevCounts["posts"])
	reportsChangePct := calcPercentChange(totalReports, prevCounts["reports"])
	commentsChangePct := calcPercentChange(totalComments, prevCounts["comments"])
	mediaChangePct := calcPercentChange(totalMedia, prevCounts["media"])
	groupsChangePct := calcPercentChange(totalGroups, prevCounts["chats"])
	communitiesChangePct := calcPercentChange(totalCommunities, prevCounts["communities"])

	return dto.AdminAnalyticsResponse{
		TotalUsers:              totalUsers,
		TotalPosts:              totalPosts,
		TotalReports:            totalReports,
		TotalComments:           totalComments,
		TotalMedia:              totalMedia,
		TotalGroups:             totalGroups,
		TotalCommunities:        totalCommunities,
		TotalActiveBans:         activeBanCount,
		PendingReports:          pendingReports,
		FlaggedMediaCount:       flaggedMediaCount,
		ActiveUsersToday:        activeUsersToday,
		TotalLikes:              totalLikes,
		TotalShares:             totalShares,
		UsersChangePercent:      usersChangePct,
		PostsChangePercent:      postsChangePct,
		ReportsChangePercent:    reportsChangePct,
		CommentsChangePercent:   commentsChangePct,
		MediaChangePercent:      mediaChangePct,
		GroupsChangePercent:     groupsChangePct,
		CommunitiesChangePercent: communitiesChangePct,
		ChartData:               chartDataUsers,
		ChartDataUsers:          chartDataUsers,
		ChartDataPosts:          chartDataPosts,
		ChartDataReports:        chartDataReports,
		ChartDataComments:       chartDataComments,
		TopUsers:                topUsers,
		TopPosts:                topPosts,
		UserStatusDistribution:  userDist,
		ReportStatusDistribution: reportDist,
		GeneratedAt:             time.Now().UTC(),
	}, nil
}

func (s *AdminService) ListUsers(ctx context.Context, userID string, input dto.AdminUserFilterInput) (dto.AdminUserListResponse, error) {
	if err := s.ensureAdmin(ctx, userID); err != nil {
		return dto.AdminUserListResponse{}, err
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

	statusFilter := strings.TrimSpace(strings.ToLower(input.Status))
	if statusFilter != "" {
		switch statusFilter {
		case string(models.UserStatusActive), string(models.UserStatusBanned), string(models.UserStatusSuspended):
		default:
			return dto.AdminUserListResponse{}, errorsapp.New(errorsapp.ErrCodeAdminInvalidStatus)
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
		return errorsapp.New(errorsapp.ErrCodeAdminNotSuperadmin)
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
		return errorsapp.New(errorsapp.ErrCodeAdminInvalidStatus)
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
		return errorsapp.New(errorsapp.ErrCodeAdminInvalidStatus)
	}

	if targetUser.Status == status {
		return nil
	}

	if err := s.authRepo.UpdateUserStatus(ctx, targetUserID, status); err != nil {
		return err
	}

	statusMsg := fmt.Sprintf("Trạng thái tài khoản của bạn đã được cập nhật thành: %s", status)
	_, _ = s.notificationService.Create(ctx, targetUserID, nil, models.NotificationTypeMessage, statusMsg, nil, &targetUserID, nil)

	return nil
}

func (s *AdminService) BanUser(ctx context.Context, superAdminID, targetUserID string, input dto.AdminUserBanInput) (dto.AdminBanUserResponse, error) {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return dto.AdminBanUserResponse{}, err
	}

	targetUser, err := s.authRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return dto.AdminBanUserResponse{}, err
	}

	isTargetSuperAdmin, err := s.authRepo.HasRole(ctx, targetUserID, models.RoleSuperAdmin)
	if err != nil {
		return dto.AdminBanUserResponse{}, err
	}
	if isTargetSuperAdmin {
		return dto.AdminBanUserResponse{}, errorsapp.New(errorsapp.ErrCodeAdminNotSuperadmin)
	}

	isTargetAdmin, err := s.authRepo.HasRole(ctx, targetUserID, models.RoleAdmin)
	if err != nil {
		return dto.AdminBanUserResponse{}, err
	}
	if isTargetAdmin {
		return dto.AdminBanUserResponse{}, errorsapp.New(errorsapp.ErrCodeAdminNotSuperadmin)
	}

	if targetUser.Status == models.UserStatusBanned {
		return dto.AdminBanUserResponse{}, errorsapp.New(errorsapp.ErrCodeAdminAlreadyInStatus)
	}

	var expiresAt *time.Time
	durationKey := strings.TrimSpace(input.Duration)

	if durationKey != "permanent" {
		duration, ok := banDurationMap[durationKey]
		if !ok {
			return dto.AdminBanUserResponse{}, errorsapp.New(errorsapp.ErrCodeAdminInvalidBanDuration)
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

	banMsg := fmt.Sprintf("Tài khoản của bạn đã bị cấm. Lý do: %s", input.Reason)
	if expiresAt != nil {
		banMsg += fmt.Sprintf(". Thời hạn: %s", expiresAt.Format("02/01/2006 15:04"))
	}
	_, _ = s.notificationService.Create(ctx, targetUserID, nil, models.NotificationTypeMessage, banMsg, nil, &targetUserID, nil)

	return dto.AdminBanUserResponse{
		Message: "ban user thành công",
		BanUtil: expiresAt,
	}, nil
}

func (s *AdminService) ListPosts(ctx context.Context, adminID string, input dto.AdminPostFilterInput) (dto.AdminPostListResponse, error) {
	if adminID == "" {
		return dto.AdminPostListResponse{}, errorsapp.New(errorsapp.ErrCodeAdminNoAccess)
	}

	if err := s.ensureAdmin(ctx, adminID); err != nil {
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
			return dto.AdminPostListResponse{}, errorsapp.New(errorsapp.ErrCodeAdminInvalidStatus)
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

	userIDs := make([]string, 0, len(posts))
	seen := make(map[string]struct{}, len(posts))
	for _, p := range posts {
		if _, dup := seen[p.UserID]; !dup {
			seen[p.UserID] = struct{}{}
			userIDs = append(userIDs, p.UserID)
		}
	}

	userMap := make(map[string]struct {
		Username    string
		DisplayName string
		AvatarURI   string
	}, len(userIDs))

	if users, err := s.authRepo.FindByIDs(ctx, userIDs); err == nil {
		for _, u := range users {
			entry := userMap[u.ID]
			entry.Username = u.Username
			userMap[u.ID] = entry
		}
	}
	if profiles, err := s.profileRepo.FindByIDs(ctx, userIDs); err == nil {
		for _, p := range profiles {
			entry := userMap[p.UserID]
			entry.DisplayName = p.DisplayName
			entry.AvatarURI = p.AvatarURI
			userMap[p.UserID] = entry
		}
	}

	items := make([]dto.AdminPostListItem, 0, len(posts))
	for _, p := range posts {
		info := userMap[p.UserID]
		items = append(items, dto.AdminPostListItem{
			ID:            p.ID,
			UserID:        p.UserID,
			Username:      info.Username,
			DisplayName:   info.DisplayName,
			AvatarURI:     info.AvatarURI,
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

	// Load media URIs for all posts
	postIDs := make([]string, 0, len(posts))
	for _, p := range posts {
		postIDs = append(postIDs, p.ID)
	}
	mediaMap, err := s.mediaRepo.GetByPostIDs(ctx, postIDs)
	if err != nil {
		log.Printf("[ListPosts] không thể lấy media: %v", err)
	} else {
		for i, item := range items {
			if mediaList, ok := mediaMap[item.ID]; ok && len(mediaList) > 0 {
				uris := make([]string, 0, len(mediaList))
				for _, m := range mediaList {
					uris = append(uris, m.FileURI)
				}
				items[i].MediaURIs = uris
			}
		}
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
		return errorsapp.New(errorsapp.ErrCodeAdminNotFound)
	}

	if post.Status == models.PostStatusHidden {
		return errorsapp.New(errorsapp.ErrCodeAdminAlreadyInStatus)
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

	postIDPtr := postID
	_, err = s.notificationService.Create(
		ctx,
		post.UserID,
		nil,
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
		return errorsapp.New(errorsapp.ErrCodeAdminNotFound)
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
		return errorsapp.New(errorsapp.ErrCodeAdminInvalidStatus)
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

	postIDPtr := postID
	_, err = s.notificationService.Create(
		ctx,
		post.UserID,
		nil,
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

func (s *AdminService) ListComments(ctx context.Context, adminID string, input dto.AdminCommentFilterInput) (dto.AdminCommentListResponse, error) {
	if adminID == "" {
		return dto.AdminCommentListResponse{}, errorsapp.New(errorsapp.ErrCodeAdminNoAccess)
	}

	if err := s.ensureAdmin(ctx, adminID); err != nil {
		return dto.AdminCommentListResponse{}, err
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
		case string(models.CommentStatusActive),
			string(models.CommentStatusHidden),
			string(models.CommentStatusDeleted):
		default:
			return dto.AdminCommentListResponse{}, errorsapp.New(errorsapp.ErrCodeAdminInvalidStatus)
		}
	}

	comments, err := s.postRepo.ListComments(ctx, input.Keyword, status, pageSize, (page-1)*pageSize)
	if err != nil {
		return dto.AdminCommentListResponse{}, err
	}

	total, err := s.postRepo.CountComments(ctx, input.Keyword, status)
	if err != nil {
		return dto.AdminCommentListResponse{}, err
	}

	userIDs := make([]string, 0, len(comments))
	seen := make(map[string]struct{}, len(comments))
	for _, c := range comments {
		if _, dup := seen[c.UserID]; !dup {
			seen[c.UserID] = struct{}{}
			userIDs = append(userIDs, c.UserID)
		}
	}

	userMap := make(map[string]struct {
		Username    string
		DisplayName string
		AvatarURI   string
	}, len(userIDs))

	if users, err := s.authRepo.FindByIDs(ctx, userIDs); err == nil {
		for _, u := range users {
			entry := userMap[u.ID]
			entry.Username = u.Username
			userMap[u.ID] = entry
		}
	}
	if profiles, err := s.profileRepo.FindByIDs(ctx, userIDs); err == nil {
		for _, p := range profiles {
			entry := userMap[p.UserID]
			entry.DisplayName = p.DisplayName
			entry.AvatarURI = p.AvatarURI
			userMap[p.UserID] = entry
		}
	}

	items := make([]dto.AdminCommentListItem, 0, len(comments))
	for _, c := range comments {
		info := userMap[c.UserID]
		items = append(items, dto.AdminCommentListItem{
			ID:           c.ID,
			UserID:       c.UserID,
			Username:     info.Username,
			DisplayName:  info.DisplayName,
			PostID:       c.PostID,
			Content:      c.Content,
			Status:       string(c.Status),
			ReviewReason: c.ReviewReason,
			CreatedAt:    c.CreatedAt,
			UpdatedAt:    c.UpdatedAt,
		})
	}

	resp := dto.AdminCommentListResponse{
		Comments: items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	if len(items) == 0 {
		resp.Message = "Không tìm thấy bình luận"
	}
	return resp, nil
}

func (s *AdminService) HideComment(ctx context.Context, superAdminID, commentID string, input dto.AdminHidePostInput) error {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return err
	}

	comment, err := s.postRepo.FindCommentByID(ctx, commentID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeAdminNotFound)
	}

	if comment.Status == models.CommentStatusHidden {
		return errorsapp.New(errorsapp.ErrCodeAdminAlreadyInStatus)
	}

	moderation := models.NewModerationLog(superAdminID, models.ModerationActionDelete, models.ModerationTargetComment, commentID, input.Reason)
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()

	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return err
	}

	if err := s.postRepo.UpdateCommentStatus(ctx, commentID, models.CommentStatusHidden, input.Reason); err != nil {
		return err
	}

	postIDPtr := comment.PostID
	commentIDPtr := commentID
	_, err = s.notificationService.Create(
		ctx,
		comment.UserID,
		nil,
		models.NotificationTypeMessage,
		"Bình luận của bạn đã bị ẩn vì: "+input.Reason,
		&postIDPtr,
		nil,
		&commentIDPtr,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *AdminService) RevealComment(ctx context.Context, superAdminID, commentID string) error {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return err
	}

	comment, err := s.postRepo.FindCommentByID(ctx, commentID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeAdminNotFound)
	}

	if comment.Status != models.CommentStatusHidden {
		return errorsapp.New(errorsapp.ErrCodeAdminAlreadyInStatus)
	}

	moderation := models.NewModerationLog(superAdminID, models.ModerationActionUpdate, models.ModerationTargetComment, commentID, "Hiện bình luận")
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()

	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return err
	}

	if err := s.postRepo.UpdateCommentStatus(ctx, commentID, models.CommentStatusActive, ""); err != nil {
		return err
	}

	postIDPtr := comment.PostID
	commentIDPtr := commentID
	_, err = s.notificationService.Create(
		ctx,
		comment.UserID,
		nil,
		models.NotificationTypeMessage,
		"Bình luận của bạn đã được hiện lại",
		&postIDPtr,
		nil,
		&commentIDPtr,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *AdminService) ListReports(ctx context.Context, adminID string, input dto.AdminReportFilterInput) (dto.AdminReportListResponse, error) {
	if err := s.ensureAdmin(ctx, adminID); err != nil {
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

func (s *AdminService) GetReportDetail(ctx context.Context, adminID, reportID string) (dto.AdminReportDetailResponse, error) {
	if err := s.ensureAdmin(ctx, adminID); err != nil {
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
		comment, err := s.postRepo.FindCommentByID(ctx, *report.TargetCommentID)
		if err == nil {
			detail.TargetType = "comment"
			detail.CommentOwnerID = &comment.UserID
			detail.CommentContent = &comment.Content
		} else {
			detail.TargetType = "comment"
		}
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
		return errorsapp.New(errorsapp.ErrCodeAdminAlreadyInStatus)
	}

	action := strings.TrimSpace(strings.ToLower(input.Action))
	if action != "cancel" && action != "hide" && action != "ban" {
		return errorsapp.New(errorsapp.ErrCodeAdminActionInvalid)
	}

	if (action == "hide" || action == "ban") && strings.TrimSpace(input.Reason) == "" {
		return errorsapp.New(errorsapp.ErrCodeAdminReasonRequired)
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
			return errorsapp.New(errorsapp.ErrCodeAdminActionInvalid)
		} else if report.TargetCommentID != nil {
			if _, err := s.postRepo.FindCommentByID(ctx, *report.TargetCommentID); err != nil {
				return fmt.Errorf("comment không tồn tại: %w", err)
			}
			if err := s.postRepo.UpdateCommentStatus(ctx, *report.TargetCommentID, models.CommentStatusHidden, input.Reason); err != nil {
				return fmt.Errorf("hide comment: %w", err)
			}
			descendantIDs, err := s.postRepo.FindDescendantCommentIDs(ctx, *report.TargetCommentID)
			if err == nil && len(descendantIDs) > 0 {
				_ = s.postRepo.HideCommentsByIDs(ctx, descendantIDs, "Bình luận mà bạn trả lời đã bị ẩn")
			}
			status = models.ReportStatusResolved

			moderation := models.NewModerationLog(superAdminID, models.ModerationActionDelete, models.ModerationTargetComment, *report.TargetCommentID, input.Reason)
			moderation.ID = utils.GenerateUUID()
			moderation.CreatedAt = time.Now().UTC()
			if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
				return fmt.Errorf("create moderation log: %w", err)
			}
		} else {
			return errorsapp.New(errorsapp.ErrCodeAdminActionInvalid)
		}
	}

	if action == "ban" {
		if report.TargetUserID == nil {
			return errorsapp.New(errorsapp.ErrCodeAdminActionInvalid)
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
	_, _ = s.notificationService.Create(ctx, report.ReporterID, nil, models.NotificationTypeMessage, reporterMessage, report.TargetPostID, report.TargetUserID, report.TargetCommentID)

	if report.TargetPostID != nil {
		post, err := s.postRepo.FindByID(ctx, *report.TargetPostID)
		if err == nil {
			targetMessage := fmt.Sprintf("Bài viết của bạn đã bị báo cáo và đã được %s bởi quản trị viên.", action)
			_, _ = s.notificationService.Create(ctx, post.UserID, nil, models.NotificationTypeMessage, targetMessage, report.TargetPostID, nil, nil)
		}
	} else if report.TargetUserID != nil {
		targetMessage := fmt.Sprintf("Tài khoản của bạn đã bị báo cáo và đã được %s bởi quản trị viên.", action)
		_, _ = s.notificationService.Create(ctx, *report.TargetUserID, nil, models.NotificationTypeMessage, targetMessage, nil, report.TargetUserID, nil)
	} else if report.TargetCommentID != nil {
		comment, err := s.postRepo.FindCommentByID(ctx, *report.TargetCommentID)
		if err == nil {
			targetMessage := fmt.Sprintf("Bình luận của bạn đã bị báo cáo và đã được %s bởi quản trị viên.", action)
			_, _ = s.notificationService.Create(ctx, comment.UserID, nil, models.NotificationTypeMessage, targetMessage, nil, nil, report.TargetCommentID)
		}
	}

	if report.TargetUserID != nil && action == "ban" {
		_, _ = s.notificationService.Create(
			ctx,
			*report.TargetUserID,
			nil,
			models.NotificationTypeMessage,
			"Tài khoản của bạn đã bị cấm vì vi phạm báo cáo.",
			nil,
			report.TargetUserID,
			nil,
		)
	}

	return nil
}

// ── Admin Media Management ──

// mediaReviewTransitions defines valid status transitions for admin review.
var mediaReviewTransitions = map[models.MediaStatus][]models.MediaStatus{
	models.MediaStatusFlagged: {models.MediaStatusApproved, models.MediaStatusRejected},
}

func (s *AdminService) ListFlaggedMedia(ctx context.Context, adminID string, input dto.AdminMediaFilterInput) (dto.AdminMediaListResponse, error) {
	if err := s.ensureAdmin(ctx, adminID); err != nil {
		return dto.AdminMediaListResponse{}, err
	}

	page, pageSize := s.resolvePageSize(input.Page, input.PageSize)

	var status models.MediaStatus
	if input.Status != "" && !strings.EqualFold(input.Status, "all") {
		status = models.ParseMediaStatus(input.Status)
	}

	items, total, err := s.mediaRepo.GetByStatus(ctx, status, input.Keyword, page, pageSize)
	if err != nil {
		return dto.AdminMediaListResponse{}, fmt.Errorf("lấy danh sách media thất bại: %w", err)
	}

	// Collect unique user IDs for batch lookup
	userIDs := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, m := range items {
		if _, dup := seen[m.UserID]; !dup {
			seen[m.UserID] = struct{}{}
			userIDs = append(userIDs, m.UserID)
		}
	}

	type userInfo struct {
		Username    string
		DisplayName string
	}
	userMap := make(map[string]userInfo, len(userIDs))
	for _, uid := range userIDs {
		u, err := s.authRepo.FindByID(ctx, uid)
		if err != nil || u == nil {
			continue
		}
		info := userInfo{Username: u.Username}
		profile, err := s.profileRepo.FindByUserID(ctx, uid)
		if err == nil && profile != nil {
			info.DisplayName = profile.DisplayName
		}
		userMap[uid] = info
	}

	mediaItems := make([]dto.AdminMediaItem, 0, len(items))
	for _, m := range items {	
		mediaItems = append(mediaItems, dto.AdminMediaItem{
			ID:           m.ID,
			UserID:       m.UserID,
			FileURI:      m.FileURI,
			FileType:     m.FileType,
			FileSize:     m.FileSize,
			Status:       string(m.Status),
			ReviewReason: m.ReviewReason,
			CreatedAt:    m.CreatedAt.Format(time.RFC3339),
		})
	}

	return dto.AdminMediaListResponse{
		Items: mediaItems,
		Total: total,
		Page:  page,
	}, nil
}

func (s *AdminService) ReviewMedia(ctx context.Context, adminID, mediaID string, input dto.AdminReviewMediaInput) error {
	if err := s.ensureAdmin(ctx, adminID); err != nil {
		return err
	}

	media, err := s.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeAdminNotFound)
	}

	var newStatus models.MediaStatus
	var notificationMsg string

	switch input.Action {
	case "approve":
		newStatus = models.MediaStatusApproved
		notificationMsg = "Ảnh/video của bạn đã được admin phê duyệt."
	case "reject":
		newStatus = models.MediaStatusRejected
		notificationMsg = "Ảnh/video của bạn đã bị admin từ chối: " + input.Reason
	default:
		return errorsapp.New(errorsapp.ErrCodeAdminActionInvalid)
	}

	allowed, ok := mediaReviewTransitions[media.Status]
	if !ok {
		return errorsapp.New(errorsapp.ErrCodeAdminInvalidStatus)
	}
	validTransition := false
	for _, s := range allowed {
		if s == newStatus {
			validTransition = true
			break
		}
	}
	if !validTransition {
		return errorsapp.New(errorsapp.ErrCodeAdminInvalidStatus)
	}

	if err := s.mediaRepo.UpdateStatusAndReview(ctx, mediaID, newStatus, input.Reason); err != nil {
		return fmt.Errorf("cập nhật trạng thái media thất bại: %w", err)
	}

	moderation := models.NewModerationLog(adminID, models.ModerationActionUpdate,
		models.ModerationTargetMedia, mediaID, input.Reason)
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()
	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return fmt.Errorf("ghi log kiểm duyệt thất bại: %w", err)
	}

	var notifType models.NotificationType
	switch newStatus {
	case models.MediaStatusApproved:
		notifType = models.NotificationTypeMediaApproved
	case models.MediaStatusRejected:
		notifType = models.NotificationTypeMediaRejected
	case models.MediaStatusFlagged:
		notifType = models.NotificationTypeMediaFlagged
	default:
		notifType = models.NotificationTypeMessage
	}

	if _, err := s.notificationService.Create(ctx, media.UserID, nil,
		notifType, notificationMsg, media.PostID, nil, nil); err != nil {
		return fmt.Errorf("gửi thông báo thất bại: %w", err)
	}

	// Nếu admin reject, xoá file khỏi Cloudinary để tiết kiệm storage.
	// Không xoá ở AI reject vì admin có thể manual override.
	// Side effect — không block response (admin đã reject, DB đã update).
	if newStatus == models.MediaStatusRejected {
		if s.cloudinary == nil {
			log.Printf("[Admin] cảnh báo: Cloudinary chưa được cấu hình — không thể xoá media %s khỏi Cloudinary", mediaID)
		} else {
			publicID, resourceType, parseErr := parseCloudinaryPublicID(media.FileURI)
			if parseErr == nil {
				if _, err := s.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{
					PublicID:     publicID,
					ResourceType: resourceType,
				}); err != nil {
					return fmt.Errorf("xoá file khỏi Cloudinary thất bại: %w", err)
				}
				// Xoá URL khỏi DB để tránh record trỏ tới file không còn tồn tại.
				if err := s.mediaRepo.ClearFileURI(ctx, mediaID); err != nil {
					log.Printf("[Admin] không thể xoá FileURI của media %s: %v", mediaID, err)
				}
			} else {
				log.Printf("[Admin] không thể xoá media %s khỏi Cloudinary — parse URL thất bại: %v", mediaID, parseErr)
			}
		}
	}

	return nil
}

// CleanupRejectedMedia xoá tất cả media bị AI reject quá 7 ngày
// khỏi Cloudinary và DB để giải phóng storage.
func (s *AdminService) CleanupRejectedMedia(ctx context.Context, adminID string) (int, error) {
	if err := s.ensureAdmin(ctx, adminID); err != nil {
		return 0, err
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -7)
	items, err := s.mediaRepo.GetRejectedOlderThan(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("lấy media reject cũ thất bại: %w", err)
	}

	if len(items) == 0 {
		return 0, nil
	}

	cleaned := 0
	for _, m := range items {
		if s.cloudinary != nil {
			publicID, resourceType, parseErr := parseCloudinaryPublicID(m.FileURI)
			if parseErr == nil {
				if _, err := s.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{
					PublicID:     publicID,
					ResourceType: resourceType,
				}); err != nil {
					log.Printf("[Admin] cleanup — xoá Cloudinary media %s thất bại: %v", m.ID, err)
				}
			}
		}

		if err := s.mediaRepo.DeleteWithStorageAdjustment(ctx, m.UserID, &m); err != nil {
			log.Printf("[Admin] cleanup — xoá DB media %s thất bại: %v", m.ID, err)
			continue
		}
		cleaned++
	}

	return cleaned, nil
}

// ── Admin Ad Management ─────────────────────────────────────────────────────

func (s *AdminService) ListAds(ctx context.Context, adminID string, input dto.AdminAdFilterInput) (dto.AdminAdListResponse, error) {
	if err := s.ensureAdmin(ctx, adminID); err != nil {
		return dto.AdminAdListResponse{}, err
	}

	page, pageSize := s.resolvePageSize(input.Page, input.PageSize)
	keyword := strings.TrimSpace(input.Keyword)
	status := strings.TrimSpace(strings.ToLower(input.Status))

	total, err := s.adRepo.CountAds(ctx, keyword, status)
	if err != nil {
		return dto.AdminAdListResponse{}, fmt.Errorf("đếm quảng cáo thất bại: %w", err)
	}

	ads, err := s.adRepo.ListAds(ctx, keyword, status, pageSize, (page-1)*pageSize)
	if err != nil {
		return dto.AdminAdListResponse{}, fmt.Errorf("lấy danh sách quảng cáo thất bại: %w", err)
	}

	if ads == nil {
		ads = []dto.AdminAdListItem{}
	}

	resp := dto.AdminAdListResponse{
		Ads:      ads,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	if len(ads) == 0 {
		resp.Message = "Không tìm thấy quảng cáo"
	}

	return resp, nil
}

func (s *AdminService) UpdateAdStatus(ctx context.Context, adminID, adID string, input dto.AdminAdStatusInput) error {
	if err := s.ensureAdmin(ctx, adminID); err != nil {
		return err
	}

	ad, err := s.adRepo.FindByID(adID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeAdminNotFound)
	}

	newStatus := strings.TrimSpace(strings.ToLower(input.Status))
	if newStatus == string(ad.Status) {
		return nil
	}

	isSuperAdmin, err := s.authRepo.HasRole(ctx, adminID, models.RoleSuperAdmin)
	if err != nil {
		return fmt.Errorf("kiểm tra quyền thất bại: %w", err)
	}

	if !isSuperAdmin && newStatus == string(models.AdStatusActive) {
		return errorsapp.New(errorsapp.ErrCodeAdminNotSuperadmin)
	}

	ad.Status = models.ParseAdStatus(newStatus)
	if err := s.adRepo.Update(ad); err != nil {
		return fmt.Errorf("cập nhật trạng thái quảng cáo thất bại: %w", err)
	}

	return nil
}

func (s *AdminService) DeleteAd(ctx context.Context, adminID, adID string) error {
	if err := s.ensureSuperAdmin(ctx, adminID); err != nil {
		return err
	}

	if err := s.adRepo.Delete(ctx, adID); err != nil {
		return fmt.Errorf("xoá quảng cáo thất bại: %w", err)
	}

	return nil
}

// ── Shared helpers ─────────────────────────────────────────────────────────

func (s *AdminService) ensureAdmin(ctx context.Context, userID string) error {
	if userID == "" {
		return errorsapp.New(errorsapp.ErrCodeAdminNoAccess)
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
	return errorsapp.New(errorsapp.ErrCodeAdminNoAccess)
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
			return dto.AdminGroupListResponse{}, errorsapp.New(errorsapp.ErrCodeAdminInvalidStatus)
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
		return dto.AdminGroupDetailResponse{}, errorsapp.New(errorsapp.ErrCodeAdminNotFound)
	}
	if chat.Type != models.ChatTypeGroup {
		return dto.AdminGroupDetailResponse{}, errorsapp.New(errorsapp.ErrCodeAdminNotGroupChat)
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
		return nil, errorsapp.New(errorsapp.ErrCodeAdminNotFound)
	}
	if chat.Type != models.ChatTypeGroup {
		return nil, errorsapp.New(errorsapp.ErrCodeAdminNotGroupChat)
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
		return errorsapp.New(errorsapp.ErrCodeAdminNotFound)
	}
	if chat.Type != models.ChatTypeGroup {
		return errorsapp.New(errorsapp.ErrCodeAdminNotGroupChat)
	}

	if chat.Status == models.ChatStatusArchived && newStatus != models.ChatStatusArchived {
		return errorsapp.New(errorsapp.ErrCodeAdminArchivedRestricted)
	}

	if chat.Status == newStatus {
		return errorsapp.New(errorsapp.ErrCodeAdminAlreadyInStatus)
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
		_, _ = s.notificationService.Create(ctx, *chat.CreatorID, nil, models.NotificationTypeMessage, content, nil, nil, nil)
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
		return errorsapp.New(errorsapp.ErrCodeAdminNotFound)
	}
	if chat.Type != models.ChatTypeGroup {
		return errorsapp.New(errorsapp.ErrCodeAdminNotGroupChat)
	}

	moderation := models.NewModerationLog(superAdminID, models.ModerationActionWarn, models.ModerationTargetGroupChat, chatID, input.Reason)
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()
	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return err
	}

	if chat.CreatorID != nil {
		content := fmt.Sprintf("Cảnh báo nhóm chat '%s': %s", chat.Name, input.Message)
		_, _ = s.notificationService.Create(ctx, *chat.CreatorID, nil, models.NotificationTypeMessage, content, nil, nil, nil)
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
			return dto.AdminCommunityListResponse{}, errorsapp.New(errorsapp.ErrCodeAdminInvalidStatus)
		}
	}

	privacy := strings.TrimSpace(strings.ToLower(input.Privacy))
	if privacy != "" {
		switch privacy {
		case string(models.PrivacyPublic), string(models.PrivacyInvitationOnly):
		default:
			return dto.AdminCommunityListResponse{}, errorsapp.New(errorsapp.ErrCodeAdminInvalidStatus)
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
		return dto.AdminCommunityDetailResponse{}, errorsapp.New(errorsapp.ErrCodeAdminNotFound)
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
		ID:            community.ID,
		Name:          community.Name,
		Description:   community.Description,
		AvatarURI:     community.AvatarURI,
		BackgroundURI: community.BackgroundURI,
		CreatorID:     community.CreatorID,
		CreatorName:   creatorName,
		Privacy:       string(community.Privacy),
		Status:        string(community.Status),
		AutoApprove:   community.AutoApprove,
		MemberCount:   len(members),
		Members:       members,
		CreatedAt:     community.CreatedAt,
		UpdatedAt:     community.UpdatedAt,
	}, nil
}

func (s *AdminService) ListCommunityMembers(ctx context.Context, userID, communityID string) ([]dto.AdminCommunityMember, error) {
	if err := s.ensureAdmin(ctx, userID); err != nil {
		return nil, err
	}

	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeAdminNotFound)
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
		return errorsapp.New(errorsapp.ErrCodeAdminNotFound)
	}

	if community.Status == models.CommunityStatusArchived && newStatus != models.CommunityStatusArchived && newStatus != models.CommunityStatusActive {
		return errorsapp.New(errorsapp.ErrCodeAdminArchivedRestricted)
	}

	if community.Status == newStatus {
		return errorsapp.New(errorsapp.ErrCodeAdminAlreadyInStatus)
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
	_, _ = s.notificationService.Create(ctx, community.CreatorID, nil, models.NotificationTypeMessage, content, nil, nil, nil)

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

func (s *AdminService) UnarchiveCommunity(ctx context.Context, superAdminID, communityID string) error {
	return s.moderateCommunity(ctx, superAdminID, communityID, models.ModerationActionUpdate, models.ModerationActionUpdate, models.CommunityStatusActive, "bỏ đình chỉ", "Bỏ đình chỉ cộng đồng")
}

func (s *AdminService) WarnCommunity(ctx context.Context, superAdminID, communityID string, input dto.AdminWarnInput) error {
	if err := s.ensureSuperAdmin(ctx, superAdminID); err != nil {
		return err
	}

	community, err := s.communityRepo.FindByID(ctx, communityID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeAdminNotFound)
	}

	moderation := models.NewModerationLog(superAdminID, models.ModerationActionWarn, models.ModerationTargetCommunity, communityID, input.Reason)
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()
	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return err
	}

	content := fmt.Sprintf("Cảnh báo cộng đồng '%s': %s", community.Name, input.Message)
	_, _ = s.notificationService.Create(ctx, community.CreatorID, nil, models.NotificationTypeMessage, content, nil, nil, nil)

	return nil
}

func (s *AdminService) ensureSuperAdmin(ctx context.Context, userID string) error {
	if userID == "" {
		return errorsapp.New(errorsapp.ErrCodeAdminNoAccess)
	}
	isSuperAdmin, err := s.authRepo.HasRole(ctx, userID, models.RoleSuperAdmin)
	if err != nil {
		return fmt.Errorf("check superadmin: %w", err)
	}
	if !isSuperAdmin {
		return errorsapp.New(errorsapp.ErrCodeAdminNotSuperadmin)
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

func calcPercentChange(current, previous int64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return float64(current-previous) / float64(previous) * 100
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
		return nil, errorsapp.New(errorsapp.ErrCodeAdminInvalidBanDuration)
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
		return errorsapp.New(errorsapp.ErrCodeAdminNotGroupChat)
	}

	participantIDs, err := s.chatRepo.GetParticipantIDs(ctx, chatID)
	if err != nil {
		return err
	}

	if len(participantIDs) > 1 {
		return errorsapp.New(errorsapp.ErrCodeAdminHasOtherMembers)
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
		return errorsapp.New(errorsapp.ErrCodeAdminNotFound)
	}
	if community.CreatorID == "" {
		return errorsapp.New(errorsapp.ErrCodeAdminNoCreator)
	}

	memberIDs, err := s.communityRepo.FindCommunityMemberIDs(ctx, communityID)
	if err != nil {
		return err
	}

	if len(memberIDs) > 1 {
		return errorsapp.New(errorsapp.ErrCodeAdminHasOtherMembers)
	}

	moderation := models.NewModerationLog(superAdminID, models.ModerationActionDelete, models.ModerationTargetCommunity, communityID, input.Reason)
	moderation.ID = utils.GenerateUUID()
	moderation.CreatedAt = time.Now().UTC()
	if err := s.moderationRepo.CreateLog(ctx, &moderation); err != nil {
		return err
	}

	return s.communityRepo.RemoveMember(ctx, communityID, community.CreatorID)
}

func (s *AdminService) ListMediaGroupedByUser(ctx context.Context, adminID string, input dto.AdminMediaGroupFilterInput) (dto.AdminMediaGroupedResponse, error) {
	if err := s.ensureAdmin(ctx, adminID); err != nil {
		return dto.AdminMediaGroupedResponse{}, err
	}

	page, pageSize := s.resolvePageSize(input.Page, input.PageSize)

	groups, total, err := s.mediaRepo.GetMediaGroupsByUser(ctx, "approved", input.Keyword, page, pageSize)
	if err != nil {
		return dto.AdminMediaGroupedResponse{}, fmt.Errorf("lấy media theo user thất bại: %w", err)
	}

	resultGroups := make([]dto.AdminMediaGroupItem, 0, len(groups))
	for _, g := range groups {
		items := make([]dto.AdminMediaItem, 0, len(g.Media))
		for _, m := range g.Media {
			items = append(items, dto.AdminMediaItem{
				ID:           m.ID,
				UserID:       m.UserID,
				FileURI:      m.FileURI,
				FileType:     m.FileType,
				FileSize:     m.FileSize,
				Status:       string(m.Status),
				ReviewReason: m.ReviewReason,
				CreatedAt:    m.CreatedAt.Format(time.RFC3339),
			})
		}

		resultGroups = append(resultGroups, dto.AdminMediaGroupItem{
			UserID:      g.UserID,
			Username:    g.Username,
			DisplayName: g.DisplayName,
			AvatarURI:   g.AvatarURI,
			Media:       items,
		})
	}

	return dto.AdminMediaGroupedResponse{
		Groups:   resultGroups,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
