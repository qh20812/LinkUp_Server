package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"linkup/config"
	"linkup/db"
)

type userRow struct {
	ID    int
	Email string
}

type mediaRow struct {
	ID int
}

type postRow struct {
	ID int
}

type commentRow struct {
	ID int
}

type violationRuleRow struct {
	ID int
}

type adSeed struct {
	AdminID   int
	Title     string
	Content   string
	MediaID   sql.NullInt64
	TargetURL string
	Status    string
	Budget    float64
	StartedAt sql.NullString
	ExpiresAt sql.NullString
}

type adAnalyticsSeed struct {
	AdID       int
	UserID     sql.NullInt64
	ActionType string
	IPAddress  string
}

type notificationSeed struct {
	ReceiverID        int
	SenderID          sql.NullInt64
	Type              string
	RedirectPostID    sql.NullInt64
	RedirectUserID    sql.NullInt64
	RedirectCommentID sql.NullInt64
	Content           string
	IsRead            bool
}

type reportSeed struct {
	ReporterID      int
	ReportType      string
	TargetUserID    sql.NullInt64
	TargetPostID    sql.NullInt64
	TargetCommentID sql.NullInt64
	ViolationRuleID int
	ReasonDetail    string
	Status          string
}

type banSeed struct {
	UserID    int
	AdminID   int
	Reason    string
	ExpiresAt sql.NullString
}

type moderationLogSeed struct {
	ModeratorID int
	Action      string
	TargetType  string
	TargetID    int
	Reason      string
}

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	conn, err := db.ConnectDb(config.GetEnv())
	if err != nil {
		log.Fatalf("DB connection: failed (%v)", err)
	}
	defer conn.Close()

	if err := ensurePhase6Tables(conn); err != nil {
		log.Fatalf("ensure phase6 tables failed: %v", err)
	}

	users, err := fetchUsers(conn)
	if err != nil {
		log.Fatalf("fetch users failed: %v", err)
	}
	if len(users) < 4 {
		log.Fatalf("need at least 4 users for phase6 seeding, found %d", len(users))
	}

	media, err := fetchMedia(conn)
	if err != nil {
		log.Fatalf("fetch media failed: %v", err)
	}
	if len(media) == 0 {
		log.Fatalf("no media found for phase6 seeding")
	}

	posts, err := fetchPosts(conn)
	if err != nil {
		log.Fatalf("fetch posts failed: %v", err)
	}
	if len(posts) == 0 {
		log.Fatalf("no posts found for phase6 seeding")
	}

	comments, err := fetchComments(conn)
	if err != nil {
		log.Fatalf("fetch comments failed: %v", err)
	}

	violationRules, err := fetchViolationRules(conn)
	if err != nil {
		log.Fatalf("fetch violation rules failed: %v", err)
	}
	if len(violationRules) == 0 {
		log.Fatalf("no violation rules found for phase6 seeding")
	}

	ads := buildAds(users, media)
	adsInserted, err := seedAds(conn, ads)
	if err != nil {
		log.Fatalf("seed ads failed: %v", err)
	}

	adIDs, err := fetchAdIDs(conn, len(ads))
	if err != nil {
		log.Fatalf("fetch ads ids failed: %v", err)
	}

	analytics := buildAdAnalytics(users, adIDs)
	analyticsInserted, err := seedAdAnalytics(conn, analytics)
	if err != nil {
		log.Fatalf("seed ad analytics failed: %v", err)
	}

	notifications := buildNotifications(users, posts, comments)
	notificationsInserted, err := seedNotifications(conn, notifications)
	if err != nil {
		log.Fatalf("seed notifications failed: %v", err)
	}

	reports := buildReports(users, posts, comments, violationRules)
	reportsInserted, err := seedReports(conn, reports)
	if err != nil {
		log.Fatalf("seed reports failed: %v", err)
	}

	bans := buildBans(users)
	bansInserted, err := seedBans(conn, bans)
	if err != nil {
		log.Fatalf("seed bans failed: %v", err)
	}

	moderationLogs := buildModerationLogs(users, posts)
	moderationLogsInserted, err := seedModerationLogs(conn, moderationLogs)
	if err != nil {
		log.Fatalf("seed moderation logs failed: %v", err)
	}

	fmt.Printf("Seed phase6: success (ads=%d, ad_analytics=%d, notifications=%d, reports=%d, bans=%d, moderation_logs=%d)\n",
		adsInserted,
		analyticsInserted,
		notificationsInserted,
		reportsInserted,
		bansInserted,
		moderationLogsInserted,
	)
}

func ensurePhase6Tables(conn *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS ads (
			id INT AUTO_INCREMENT PRIMARY KEY,
			admin_id INT NOT NULL,
			title VARCHAR(255) NOT NULL,
			content TEXT,
			media_id INT NULL,
			target_url VARCHAR(500) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			budget FLOAT NOT NULL DEFAULT 0.0,
			started_at TIMESTAMP NULL,
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_ads_admin FOREIGN KEY (admin_id) REFERENCES users(id),
			CONSTRAINT fk_ads_media FOREIGN KEY (media_id) REFERENCES media(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS ad_analytics (
			id INT AUTO_INCREMENT PRIMARY KEY,
			ad_id INT NOT NULL,
			user_id INT NULL,
			action_type VARCHAR(20) NOT NULL,
			ip_address VARCHAR(45),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_ad_analytics_ad FOREIGN KEY (ad_id) REFERENCES ads(id) ON DELETE CASCADE,
			CONSTRAINT fk_ad_analytics_user FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS notifications (
			id INT AUTO_INCREMENT PRIMARY KEY,
			receiver_id INT NOT NULL,
			sender_id INT NULL,
			type VARCHAR(30) NOT NULL,
			redirect_post_id INT NULL,
			redirect_user_id INT NULL,
			redirect_comment_id INT NULL,
			content VARCHAR(500),
			is_read BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_notifications_receiver FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE,
			CONSTRAINT fk_notifications_sender FOREIGN KEY (sender_id) REFERENCES users(id),
			CONSTRAINT fk_notifications_post FOREIGN KEY (redirect_post_id) REFERENCES posts(id) ON DELETE CASCADE,
			CONSTRAINT fk_notifications_user FOREIGN KEY (redirect_user_id) REFERENCES users(id) ON DELETE CASCADE,
			CONSTRAINT fk_notifications_comment FOREIGN KEY (redirect_comment_id) REFERENCES comments(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS reports (
			id INT AUTO_INCREMENT PRIMARY KEY,
			reporter_id INT NOT NULL,
			report_type VARCHAR(20) NOT NULL,
			target_user_id INT NULL,
			target_post_id INT NULL,
			target_comment_id INT NULL,
			violation_rule_id INT NOT NULL,
			reason_detail TEXT,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_reports_reporter FOREIGN KEY (reporter_id) REFERENCES users(id),
			CONSTRAINT fk_reports_target_user FOREIGN KEY (target_user_id) REFERENCES users(id),
			CONSTRAINT fk_reports_target_post FOREIGN KEY (target_post_id) REFERENCES posts(id),
			CONSTRAINT fk_reports_target_comment FOREIGN KEY (target_comment_id) REFERENCES comments(id),
			CONSTRAINT fk_reports_violation FOREIGN KEY (violation_rule_id) REFERENCES violation_rules(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS bans (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			admin_id INT NOT NULL,
			reason VARCHAR(1000),
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_bans_user FOREIGN KEY (user_id) REFERENCES users(id),
			CONSTRAINT fk_bans_admin FOREIGN KEY (admin_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS moderation_logs (
			id INT AUTO_INCREMENT PRIMARY KEY,
			moderator_id INT NOT NULL,
			action VARCHAR(50) NOT NULL,
			target_type VARCHAR(30) NOT NULL,
			target_id INT NOT NULL,
			reason VARCHAR(1000),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_moderation_logs_moderator FOREIGN KEY (moderator_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, q := range queries {
		if _, err := conn.Exec(q); err != nil {
			return fmt.Errorf("create phase6 table: %w", err)
		}
	}

	return nil
}

func fetchUsers(conn *sql.DB) ([]userRow, error) {
	rows, err := conn.Query("SELECT id, email FROM users ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	users := make([]userRow, 0)
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.ID, &u.Email); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

func fetchMedia(conn *sql.DB) ([]mediaRow, error) {
	rows, err := conn.Query("SELECT id FROM media ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query media: %w", err)
	}
	defer rows.Close()

	media := make([]mediaRow, 0)
	for rows.Next() {
		var m mediaRow
		if err := rows.Scan(&m.ID); err != nil {
			return nil, fmt.Errorf("scan media: %w", err)
		}
		media = append(media, m)
	}

	return media, rows.Err()
}

func fetchPosts(conn *sql.DB) ([]postRow, error) {
	rows, err := conn.Query("SELECT id FROM posts ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query posts: %w", err)
	}
	defer rows.Close()

	posts := make([]postRow, 0)
	for rows.Next() {
		var p postRow
		if err := rows.Scan(&p.ID); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, p)
	}

	return posts, rows.Err()
}

func fetchComments(conn *sql.DB) ([]commentRow, error) {
	rows, err := conn.Query("SELECT id FROM comments ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query comments: %w", err)
	}
	defer rows.Close()

	comments := make([]commentRow, 0)
	for rows.Next() {
		var c commentRow
		if err := rows.Scan(&c.ID); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, c)
	}

	return comments, rows.Err()
}

func fetchViolationRules(conn *sql.DB) ([]violationRuleRow, error) {
	rows, err := conn.Query("SELECT id FROM violation_rules ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query violation rules: %w", err)
	}
	defer rows.Close()

	rules := make([]violationRuleRow, 0)
	for rows.Next() {
		var r violationRuleRow
		if err := rows.Scan(&r.ID); err != nil {
			return nil, fmt.Errorf("scan violation rule: %w", err)
		}
		rules = append(rules, r)
	}

	return rules, rows.Err()
}

func buildAds(users []userRow, media []mediaRow) []adSeed {
	items := make([]adSeed, 0, 10)
	for i := 0; i < 10; i++ {
		admin := users[i%len(users)]
		mediaID := sql.NullInt64{Valid: false}
		if len(media) > 0 {
			mediaID = sql.NullInt64{Int64: int64(media[i%len(media)].ID), Valid: true}
		}
		items = append(items, adSeed{
			AdminID:   admin.ID,
			Title:     fmt.Sprintf("Quảng cáo seed %02d", i+1),
			Content:   fmt.Sprintf("Nội dung quảng cáo thử nghiệm số %d dành cho seed data.", i+1),
			MediaID:   mediaID,
			TargetURL: fmt.Sprintf("https://example.com/landing/%02d", i+1),
			Status:    "active",
			Budget:    1000.0 + float64(i)*250.0,
			StartedAt: sql.NullString{String: "2026-01-01 00:00:00", Valid: true},
			ExpiresAt: sql.NullString{String: "2026-12-31 23:59:59", Valid: true},
		})
	}
	return items
}

func fetchAdIDs(conn *sql.DB, limit int) ([]int, error) {
	rows, err := conn.Query("SELECT id FROM ads ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("query ad ids: %w", err)
	}
	defer rows.Close()

	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan ad id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func buildAdAnalytics(users []userRow, adIDs []int) []adAnalyticsSeed {
	actions := []string{"view", "click"}
	items := make([]adAnalyticsSeed, 0)
	for i, adID := range adIDs {
		for j := 0; j < 4; j++ {
			user := users[(i+j)%len(users)]
			items = append(items, adAnalyticsSeed{
				AdID:       adID,
				UserID:     sql.NullInt64{Int64: int64(user.ID), Valid: true},
				ActionType: actions[(i+j)%len(actions)],
				IPAddress:  fmt.Sprintf("192.168.1.%d", 10+(i+j)),
			})
		}
	}
	return items
}

func buildNotifications(users []userRow, posts []postRow, comments []commentRow) []notificationSeed {
	items := make([]notificationSeed, 0)
	for i := 0; i < 25; i++ {
		receiver := users[i%len(users)]
		sender := users[(i+1)%len(users)]
		postID := sql.NullInt64{Int64: int64(posts[i%len(posts)].ID), Valid: true}
		commentID := sql.NullInt64{Int64: int64(comments[i%len(comments)].ID), Valid: true}
		items = append(items, notificationSeed{
			ReceiverID:        receiver.ID,
			SenderID:          sql.NullInt64{Int64: int64(sender.ID), Valid: true},
			Type:              "comment",
			RedirectPostID:    postID,
			RedirectUserID:    sql.NullInt64{Int64: int64(sender.ID), Valid: true},
			RedirectCommentID: commentID,
			Content:           fmt.Sprintf("%s đã bình luận vào bài viết của bạn.", sender.Email),
			IsRead:            i%3 == 0,
		})
	}
	for i := 0; i < 10; i++ {
		receiver := users[(i+2)%len(users)]
		sender := users[(i+3)%len(users)]
		items = append(items, notificationSeed{
			ReceiverID:        receiver.ID,
			SenderID:          sql.NullInt64{Int64: int64(sender.ID), Valid: true},
			Type:              "follow",
			RedirectPostID:    sql.NullInt64{Valid: false},
			RedirectUserID:    sql.NullInt64{Int64: int64(sender.ID), Valid: true},
			RedirectCommentID: sql.NullInt64{Valid: false},
			Content:           fmt.Sprintf("%s đã bắt đầu theo dõi bạn.", sender.Email),
			IsRead:            i%2 == 0,
		})
	}
	return items
}

func buildReports(users []userRow, posts []postRow, comments []commentRow, rules []violationRuleRow) []reportSeed {
	items := make([]reportSeed, 0)
	for i := 0; i < 12; i++ {
		reporter := users[i%len(users)]
		target := users[(i+1)%len(users)]
		items = append(items, reportSeed{
			ReporterID:      reporter.ID,
			ReportType:      "user",
			TargetUserID:    sql.NullInt64{Int64: int64(target.ID), Valid: true},
			TargetPostID:    sql.NullInt64{Valid: false},
			TargetCommentID: sql.NullInt64{Valid: false},
			ViolationRuleID: rules[i%len(rules)].ID,
			ReasonDetail:    "Báo cáo hành vi không phù hợp của người dùng.",
			Status:          "pending",
		})
	}
	for i := 0; i < 8; i++ {
		reporter := users[(i+2)%len(users)]
		items = append(items, reportSeed{
			ReporterID:      reporter.ID,
			ReportType:      "post",
			TargetUserID:    sql.NullInt64{Valid: false},
			TargetPostID:    sql.NullInt64{Int64: int64(posts[i%len(posts)].ID), Valid: true},
			TargetCommentID: sql.NullInt64{Valid: false},
			ViolationRuleID: rules[(i+3)%len(rules)].ID,
			ReasonDetail:    "Báo cáo bài viết vi phạm chính sách.",
			Status:          "reviewed",
		})
	}
	for i := 0; i < 5; i++ {
		reporter := users[(i+3)%len(users)]
		items = append(items, reportSeed{
			ReporterID:      reporter.ID,
			ReportType:      "comment",
			TargetUserID:    sql.NullInt64{Valid: false},
			TargetPostID:    sql.NullInt64{Valid: false},
			TargetCommentID: sql.NullInt64{Int64: int64(comments[i%len(comments)].ID), Valid: true},
			ViolationRuleID: rules[(i+5)%len(rules)].ID,
			ReasonDetail:    "Báo cáo bình luận không phù hợp.",
			Status:          "resolved",
		})
	}
	return items
}

func buildBans(users []userRow) []banSeed {
	items := make([]banSeed, 0)
	for i := 0; i < 6; i++ {
		items = append(items, banSeed{
			UserID:    users[(i+1)%len(users)].ID,
			AdminID:   users[0].ID,
			Reason:    fmt.Sprintf("Vi phạm nghiêm trọng quy định nội dung #%d.", i+1),
			ExpiresAt: sql.NullString{String: "2026-12-31 23:59:59", Valid: true},
		})
	}
	return items
}

func buildModerationLogs(users []userRow, posts []postRow) []moderationLogSeed {
	actions := []string{"BAN_USER", "REVIEW_REPORT", "DELETE_POST", "REJECT_REPORT"}
	items := make([]moderationLogSeed, 0)
	for i := 0; i < 10; i++ {
		items = append(items, moderationLogSeed{
			ModeratorID: users[0].ID,
			Action:      actions[i%len(actions)],
			TargetType:  "posts",
			TargetID:    posts[i%len(posts)].ID,
			Reason:      fmt.Sprintf("Ghi nhận hành động kiểm duyệt số %d.", i+1),
		})
	}
	return items
}

func seedAds(conn *sql.DB, items []adSeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.AdminID, item.Title, item.Content, item.MediaID, item.TargetURL, item.Status, item.Budget, item.StartedAt, item.ExpiresAt})
	}
	return bulkInsertIgnore(conn, "ads", []string{"admin_id", "title", "content", "media_id", "target_url", "status", "budget", "started_at", "expires_at"}, values)
}

func seedAdAnalytics(conn *sql.DB, items []adAnalyticsSeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.AdID, item.UserID, item.ActionType, item.IPAddress})
	}
	return bulkInsertIgnore(conn, "ad_analytics", []string{"ad_id", "user_id", "action_type", "ip_address"}, values)
}

func seedNotifications(conn *sql.DB, items []notificationSeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.ReceiverID, item.SenderID, item.Type, item.RedirectPostID, item.RedirectUserID, item.RedirectCommentID, item.Content, item.IsRead})
	}
	return bulkInsertIgnore(conn, "notifications", []string{"receiver_id", "sender_id", "type", "redirect_post_id", "redirect_user_id", "redirect_comment_id", "content", "is_read"}, values)
}

func seedReports(conn *sql.DB, items []reportSeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.ReporterID, item.ReportType, item.TargetUserID, item.TargetPostID, item.TargetCommentID, item.ViolationRuleID, item.ReasonDetail, item.Status})
	}
	return bulkInsertIgnore(conn, "reports", []string{"reporter_id", "report_type", "target_user_id", "target_post_id", "target_comment_id", "violation_rule_id", "reason_detail", "status"}, values)
}

func seedBans(conn *sql.DB, items []banSeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.UserID, item.AdminID, item.Reason, item.ExpiresAt})
	}
	return bulkInsertIgnore(conn, "bans", []string{"user_id", "admin_id", "reason", "expires_at"}, values)
}

func seedModerationLogs(conn *sql.DB, items []moderationLogSeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.ModeratorID, item.Action, item.TargetType, item.TargetID, item.Reason})
	}
	return bulkInsertIgnore(conn, "moderation_logs", []string{"moderator_id", "action", "target_type", "target_id", "reason"}, values)
}

func bulkInsertIgnore(conn *sql.DB, table string, columns []string, rows [][]any) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	var builder strings.Builder
	builder.WriteString("INSERT IGNORE INTO ")
	builder.WriteString(table)
	builder.WriteString(" (")
	builder.WriteString(strings.Join(columns, ", "))
	builder.WriteString(") VALUES ")

	args := make([]any, 0, len(rows)*len(columns))
	for i, row := range rows {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("(")
		for j := range columns {
			if j > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString("?")
		}
		builder.WriteString(")")
		args = append(args, row...)
	}

	result, err := conn.Exec(builder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("insert seed rows into %s: %w", table, err)
	}

	inserted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for %s: %w", table, err)
	}
	return inserted, nil
}
