package dto

import "time"

type AdminUserFilterInput struct {
	Keyword  string `json:"keyword" form:"keyword"`
	Status   string `json:"status" form:"status"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
}

type AdminUserListItem struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Status      string     `json:"status"`
	DisplayName string     `json:"display_name"`
	AvatarURI   string     `json:"avatar_uri"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type AdminUserListResponse struct {
	Users    []AdminUserListItem `json:"users"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Message  string              `json:"message,omitempty"`
}

type AdminUserUpdateStatusInput struct {
	Status string `json:"status" binding:"required"`
}

type AdminUserBanInput struct {
	Reason   string `json:"reason" binding:"required"`
	Duration string `json:"duration" binding:"required"`
}

type AdminBanUserResponse struct {
	Message string     `json:"message"`
	BanUtil *time.Time `json:"ban_util,omitempty"`
}

type AdminPostFilterInput struct {
	Keyword  string `json:"keyword" form:"keyword"`
	Status   string `json:"status" form:"status"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
}

type AdminPostListItem struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	Status        string     `json:"status"`
	ViewsCount    int        `json:"views_count"`
	LikesCount    int        `json:"likes_count"`
	CommentsCount int        `json:"comments_count"`
	SharesCount   int        `json:"shares_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

type AdminPostListResponse struct {
	Posts    []AdminPostListItem `json:"posts"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Message  string              `json:"message,omitempty"`
}

type AdminHidePostInput struct {
	Reason string `json:"reason" binding:"required"`
}

type AdminUpdatePostStatusInput struct {
	Status string `json:"status" binding:"required"`
}

type AdminReportFilterInput struct {
	Keyword    string `json:"keyword" form:"keyword"`
	Status     string `json:"status" form:"status"`
	TargetType string `json:"target_type" form:"target_type"`
	SortBy     string `json:"sort_by" form:"sort_by"`
	Order      string `json:"order" form:"order"`
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"page_size" form:"page_size"`
}

type AdminReportListItem struct {
	ID               string    `json:"id"`
	ReporterID       string    `json:"reporter_id"`
	ReporterUsername string    `json:"reporter_username"`
	ReporterEmail    string    `json:"reporter_email"`
	TargetType       string    `json:"target_type"`
	TargetUserID     *string   `json:"target_user_id,omitempty"`
	TargetPostID     *string   `json:"target_post_id,omitempty"`
	TargetCommentID  *string   `json:"target_comment_id,omitempty"`
	ReportType       string    `json:"report_type"`
	ViolationRuleID  *string   `json:"violation_rule_id,omitempty"`
	ReasonDetail     string    `json:"reason_detail"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

type AdminReportListResponse struct {
	Reports  []AdminReportListItem `json:"reports"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

type AdminReportDetailResponse struct {
	ID               string    `json:"id"`
	ReporterID       string    `json:"reporter_id"`
	ReporterUsername string    `json:"reporter_username"`
	ReporterEmail    string    `json:"reporter_email"`
	TargetType       string    `json:"target_type"`
	TargetUserID     *string   `json:"target_user_id,omitempty"`
	TargetPostID     *string   `json:"target_post_id,omitempty"`
	TargetCommentID  *string   `json:"target_comment_id,omitempty"`
	ReportType       string    `json:"report_type"`
	ViolationRuleID  *string   `json:"violation_rule_id,omitempty"`
	ReasonDetail     string    `json:"reason_detail"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	PostOwnerID      *string   `json:"post_owner_id,omitempty"`
}

type AdminReportReviewInput struct {
	Action   string `json:"action" binding:"required"`
	Reason   string `json:"reason,omitempty"`
	Duration string `json:"duration,omitempty"`
}

// ── Admin Group Chat Management ──

type AdminGroupFilterInput struct {
	Keyword  string `json:"keyword" form:"keyword"`
	Status   string `json:"status" form:"status"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
}

type AdminGroupListItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CreatorID   *string   `json:"creator_id,omitempty"`
	CreatorName string    `json:"creator_name"`
	MemberCount int       `json:"member_count"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type AdminGroupListResponse struct {
	Groups   []AdminGroupListItem `json:"groups"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

type AdminGroupDetailResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	AvatarURI   string             `json:"avatar_uri"`
	CreatorID   *string            `json:"creator_id,omitempty"`
	CreatorName string             `json:"creator_name"`
	Type        string             `json:"type"`
	Status      string             `json:"status"`
	MemberCount int                `json:"member_count"`
	Members     []AdminGroupMember `json:"members"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   *time.Time         `json:"updated_at,omitempty"`
}

type AdminGroupMember struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	AvatarURI   string `json:"avatar_uri"`
	Role        string `json:"role"`
}

// ── Admin Media Management ──

type AdminMediaFilterInput struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size" binding:"max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=flagged rejected"`
}

type AdminMediaItem struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	FileURI   string  `json:"file_uri"`
	FileType  string  `json:"file_type"`
	FileSize  float64 `json:"file_size"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

type AdminMediaListResponse struct {
	Items []AdminMediaItem `json:"items"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
}

type AdminReviewMediaInput struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Reason string `json:"reason" binding:"required"`
}

// ── Admin Community Management ──

type AdminCommunityFilterInput struct {
	Keyword  string `json:"keyword" form:"keyword"`
	Status   string `json:"status" form:"status"`
	Privacy  string `json:"privacy" form:"privacy"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
}

type AdminCommunityListItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CreatorID   string    `json:"creator_id"`
	CreatorName string    `json:"creator_name"`
	MemberCount int       `json:"member_count"`
	Privacy     string    `json:"privacy"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type AdminCommunityListResponse struct {
	Communities []AdminCommunityListItem `json:"communities"`
	Total       int64                    `json:"total"`
	Page        int                      `json:"page"`
	PageSize    int                      `json:"page_size"`
}

type AdminCommunityDetailResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	AvatarURI   string                 `json:"avatar_uri"`
	CreatorID   string                 `json:"creator_id"`
	CreatorName string                 `json:"creator_name"`
	Privacy     string                 `json:"privacy"`
	Status      string                 `json:"status"`
	AutoApprove bool                   `json:"auto_approve"`
	MemberCount int                    `json:"member_count"`
	Members     []AdminCommunityMember `json:"members"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   *time.Time             `json:"updated_at,omitempty"`
}

type AdminCommunityMember struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	AvatarURI   string `json:"avatar_uri"`
	Role        string `json:"role"`
}

// ── Admin Moderation Actions ──

type AdminModerateInput struct {
	Reason string `json:"reason" binding:"required"`
}

type AdminWarnInput struct {
	Reason  string `json:"reason" binding:"required"`
	Message string `json:"message" binding:"required"`
}

// ── Shared Moderation Log ──

type AdminModerationLogItem struct {
	ID            string    `json:"id"`
	ModeratorID   string    `json:"moderator_id"`
	ModeratorName string    `json:"moderator_name"`
	Action        string    `json:"action"`
	TargetType    string    `json:"target_type"`
	TargetID      string    `json:"target_id"`
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
}

type AdminModerationLogListResponse struct {
	Logs     []AdminModerationLogItem `json:"logs"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}
type AdminAnalyticsFilterInput struct {
	StartDate string `form:"start_date"` // Định dạng: YYYY-MM-DD
	EndDate   string `form:"end_date"`   // Định dạng: YYYY-MM-DD
	Type      string `form:"type"`       // Lọc biểu đồ theo: "users", "posts", "reports", "all"
}

type ChartDataPoint struct {
	Date  string `json:"date"`  // Trục X: Ngày dạng chuỗi "YYYY-MM-DD"
	Count int64  `json:"count"` // Trục Y: Lượng tạo mới trong ngày
}

type AdminAnalyticsResponse struct {
	TotalUsers            int64            `json:"total_users"`
	TotalPosts            int64            `json:"total_posts"`
	TotalReports          int64            `json:"total_reports"`
	UsersChangePercent    float64          `json:"users_change_percent"`
	PostsChangePercent    float64          `json:"posts_change_percent"`
	ReportsChangePercent  float64          `json:"reports_change_percent"`
	ChartData             []ChartDataPoint `json:"chart_data,omitempty"`
	GeneratedAt           time.Time        `json:"generated_at"`
}
