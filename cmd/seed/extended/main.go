package extended

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
	"linkup/seeddata"
)

// rng cố định hạt giống để mỗi lần seed cho ra cùng một bố cục số liệu
// (ảnh chụp báo cáo chụp lại vẫn khớp).
var rng = rand.New(rand.NewSource(20251026))

// Run chèn phần dữ liệu vận hành còn lại: gói quảng cáo, chiến dịch quảng cáo
// kèm tệp đính kèm và số liệu, story kèm lượt xem/cảm xúc, thông báo, lịch sử
// cuộc gọi, hàng chờ kiểm duyệt media và cấu hình hệ thống.
func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("extended: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	if err := seedAdPackages(database, state, now); err != nil {
		return err
	}
	if err := seedPartnerSubscription(database, state, now); err != nil {
		return err
	}
	if err := seedAds(database, state, now); err != nil {
		return err
	}
	if err := seedStories(database, state, now); err != nil {
		return err
	}
	if err := seedNotifications(database, state, now); err != nil {
		return err
	}
	if err := seedCalls(database, state, now); err != nil {
		return err
	}
	if err := seedMediaQueue(database, state, now); err != nil {
		return err
	}
	if err := seedSystemConfigs(database); err != nil {
		return err
	}

	return nil
}

func seedAdPackages(database *sql.DB, state *internal.SeedState, now time.Time) error {
	for _, p := range seeddata.AdPackages {
		id := internal.UUID()
		state.AdPackageIDByKey[p.Key] = id

		if err := internal.Exec(database,
			`INSERT INTO ad_packages (id, name, description, price_monthly, max_slots, max_duration_days,
				supports_video, supports_carousel, has_advanced_analytics, sort_order, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, p.Name, p.Description, p.PriceMonthly, p.MaxSlots, p.MaxDurationDays,
			boolToInt(p.SupportsVideo), boolToInt(p.SupportsCarousel), boolToInt(p.HasAdvancedAnalytics),
			p.SortOrder, now,
		); err != nil {
			return fmt.Errorf("extended: insert ad package %s: %w", p.Name, err)
		}
	}
	return nil
}

// seedPartnerSubscription tạo đúng một gói đang hoạt động cho tài khoản đối tác —
// đúng như luồng đăng ký thật (mỗi user chỉ có một subscription active).
func seedPartnerSubscription(database *sql.DB, state *internal.SeedState, now time.Time) error {
	packageID, ok := state.AdPackageIDByKey["business"]
	if !ok {
		return fmt.Errorf("extended: business package missing")
	}

	// Đếm chiến dịch còn chiếm slot: bỏ qua chiến dịch đã kết thúc.
	slotsUsed := 0
	for _, a := range seeddata.Ads {
		if a.Status != "completed" {
			slotsUsed++
		}
	}

	startedAt := now.Add(-40 * 24 * time.Hour)
	expiresAt := startedAt.AddDate(0, 0, 90)

	if err := internal.Exec(database,
		`INSERT INTO partner_subscriptions
			(id, user_id, package_id, slots_used, started_at, expires_at, status, auto_renew, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, 'active', 1, ?, ?)`,
		internal.UUID(), state.PartnerUserID, packageID, slotsUsed, startedAt, expiresAt, startedAt, now,
	); err != nil {
		return fmt.Errorf("extended: insert partner subscription: %w", err)
	}
	return nil
}

func seedAds(database *sql.DB, state *internal.SeedState, now time.Time) error {
	for _, a := range seeddata.Ads {
		partnerID, ok := state.UserIDByUsername[a.Partner]
		if !ok {
			return fmt.Errorf("extended: unknown ad partner %q", a.Partner)
		}
		packageID := any(nil)
		if a.Package != "" {
			id, ok := state.AdPackageIDByKey[a.Package]
			if !ok {
				return fmt.Errorf("extended: ad %s references unknown package %q", a.Key, a.Package)
			}
			packageID = id
		}

		createdAt := now.Add(-time.Duration(a.AgeDays) * 24 * time.Hour)

		var startedAt, expiresAt any
		// Chiến dịch chờ duyệt / bị từ chối thì chưa có mốc chạy.
		if a.Status != "pending" && a.Status != "rejected" {
			startedAt = createdAt
			if a.DurationDays > 0 {
				expiresAt = createdAt.Add(time.Duration(a.DurationDays) * 24 * time.Hour)
			}
		}

		// target_locations phải là JSON hợp lệ: truy vấn xếp hạng dùng
		// JSON_LENGTH(target_locations); NULL làm bài quảng cáo bị loại.
		locations := "[]"
		if len(a.TargetLocations) > 0 {
			raw, err := json.Marshal(a.TargetLocations)
			if err != nil {
				return fmt.Errorf("extended: marshal target_locations for %s: %w", a.Key, err)
			}
			locations = string(raw)
		}

		adID := internal.UUID()
		state.AdIDs = append(state.AdIDs, adID)
		state.AdIDByKey[a.Key] = adID

		if err := internal.Exec(database,
			`INSERT INTO ads (id, partner_id, package_id, title, content, format, target_url, status,
				budget, daily_budget, total_spent, cpm_price, cpc_price, max_impressions,
				target_gender, target_age_min, target_age_max, target_locations,
				started_at, expires_at, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			adID, partnerID, packageID, a.Title, a.Content, a.Format, a.TargetURL, a.Status,
			a.Budget, a.DailyBudget, a.TotalSpent, a.CPMPrice, a.CPCPrice, a.MaxImpressions,
			a.TargetGender, a.TargetAgeMin, a.TargetAgeMax, locations,
			startedAt, expiresAt, createdAt,
		); err != nil {
			return fmt.Errorf("extended: insert ad %s: %w", a.Key, err)
		}

		for _, m := range a.Media {
			if err := internal.Exec(database,
				`INSERT INTO ad_media (id, ad_id, url, media_type, sort_order, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
				internal.UUID(), adID, m.URL, m.Type, m.Order, createdAt,
			); err != nil {
				return fmt.Errorf("extended: insert ad media for %s: %w", a.Key, err)
			}
		}

		if err := seedAdAnalytics(database, state, a, adID, createdAt, now); err != nil {
			return err
		}
	}
	return nil
}

// seedAdAnalytics sinh một lượng dòng thống kê đủ để bảng điều khiển hiển thị
// các chỉ số tự nhiên (CTR ~4-6%) mà không phải chèn hàng trăm nghìn dòng.
func seedAdAnalytics(database *sql.DB, state *internal.SeedState, ad seeddata.Ad, adID string, startedAt, now time.Time) error {
	if ad.Impressions <= 0 || ad.Clicks <= 0 {
		return nil
	}

	// Quy đổi số hiển thị trong seeddata về số dòng thực tế.
	impressions := 180 + ad.Impressions/800
	clicks := ad.Clicks / 240
	if clicks < 3 {
		clicks = 3
	}

	type actionCount struct {
		action string
		count  int
	}
	plan := []actionCount{
		{"impression", impressions},
		{"view", impressions * 6 / 10},
		{"click", clicks},
	}
	switch ad.Format {
	case "video":
		starts := impressions * 3 / 10
		plan = append(plan, actionCount{"video_start", starts}, actionCount{"video_end", starts * 55 / 100})
	case "carousel":
		plan = append(plan, actionCount{"swipe", impressions * 2 / 10})
	}

	contentUsers := state.UserIDs[2:]
	ipPrefixes := []string{"113.161", "27.72", "171.244", "14.161", "42.113", "116.96"}

	window := now.Sub(startedAt)
	if window <= 0 {
		window = time.Hour
	}

	// Chèn theo lô: mỗi dòng một round-trip tới DB từ xa khiến bước extended
	// mất vài phút, trong khi dữ liệu chỉ để hiển thị biểu đồ.
	const batchSize = 400
	args := make([]any, 0, batchSize*6)
	placeholders := make([]string, 0, batchSize)

	flush := func() error {
		if len(placeholders) == 0 {
			return nil
		}
		query := `INSERT INTO ad_analytics (id, ad_id, user_id, action_type, ip_address, created_at) VALUES ` +
			strings.Join(placeholders, ", ")
		if err := internal.Exec(database, query, args...); err != nil {
			return fmt.Errorf("extended: insert ad_analytics batch for %s: %w", ad.Key, err)
		}
		args = args[:0]
		placeholders = placeholders[:0]
		return nil
	}

	for _, p := range plan {
		for i := 0; i < p.count; i++ {
			offset := time.Duration(rng.Int63n(int64(window)))
			createdAt := startedAt.Add(offset)
			if createdAt.After(now) {
				createdAt = now.Add(-time.Minute)
			}

			// Một phần hiển thị không gắn user (khách chưa đăng nhập).
			var visitorID any
			if p.action != "impression" || rng.Intn(10) < 7 {
				visitorID = contentUsers[rng.Intn(len(contentUsers))]
			}

			ip := fmt.Sprintf("%s.%d.%d", ipPrefixes[rng.Intn(len(ipPrefixes))], rng.Intn(255), rng.Intn(255))

			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?)")
			args = append(args, internal.UUID(), adID, visitorID, p.action, ip, createdAt)

			if len(placeholders) == batchSize {
				if err := flush(); err != nil {
					return err
				}
			}
		}
	}

	return flush()
}

func seedStories(database *sql.DB, state *internal.SeedState, now time.Time) error {
	for _, s := range seeddata.Stories {
		ownerID, ok := state.UserIDByUsername[s.User]
		if !ok {
			return fmt.Errorf("extended: story references unknown user %q", s.User)
		}

		createdAt := now.Add(-time.Duration(s.AgeHours * float64(time.Hour)))
		expiresAt := createdAt.Add(24 * time.Hour)

		storyID := internal.UUID()
		state.StoryIDs = append(state.StoryIDs, storyID)
		state.StoryIDByKey[s.Key] = storyID

		if err := internal.Exec(database,
			`INSERT INTO stories (id, user_id, media_uri, media_type, caption, created_at, expires_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			storyID, ownerID, s.MediaURI, s.Type, s.Caption, createdAt, expiresAt,
		); err != nil {
			return fmt.Errorf("extended: insert story %s: %w", s.Key, err)
		}

		for i, viewer := range s.Viewers {
			viewerID, ok := state.UserIDByUsername[viewer]
			if !ok {
				return fmt.Errorf("extended: story %s viewer %q not found", s.Key, viewer)
			}
			viewedAt := createdAt.Add(time.Duration(3+i*7) * time.Minute)
			if viewedAt.After(now) {
				viewedAt = now.Add(-time.Minute)
			}
			if err := internal.Exec(database,
				`INSERT INTO story_views (story_id, viewer_id, viewed_at) VALUES (?, ?, ?)`,
				storyID, viewerID, viewedAt,
			); err != nil {
				return fmt.Errorf("extended: insert story view %s: %w", s.Key, err)
			}
		}

		for _, r := range s.Reacts {
			reactorID, ok := state.UserIDByUsername[r.User]
			if !ok {
				return fmt.Errorf("extended: story %s reactor %q not found", s.Key, r.User)
			}
			emojiID, ok := state.EmojiIDByCode[":"+r.Emoji+":"]
			if !ok {
				return fmt.Errorf("extended: unknown story emoji %q", r.Emoji)
			}
			clickCount := r.Click
			if clickCount < 1 {
				clickCount = 1
			}
			if clickCount > 5 {
				clickCount = 5
			}
			reactedAt := createdAt.Add(5 * time.Minute)
			if reactedAt.After(now) {
				reactedAt = now.Add(-time.Minute)
			}
			if err := internal.Exec(database,
				`INSERT INTO story_interacts (story_id, user_id, `+"`type`"+`, emoji_id, click_count, created_at)
				 VALUES (?, ?, 'react', ?, ?, ?)`,
				storyID, reactorID, emojiID, clickCount, reactedAt,
			); err != nil {
				return fmt.Errorf("extended: insert story react %s: %w", s.Key, err)
			}
		}

		for _, r := range s.Replies {
			replierID, ok := state.UserIDByUsername[r.User]
			if !ok {
				return fmt.Errorf("extended: story %s replier %q not found", s.Key, r.User)
			}
			repliedAt := now.Add(-time.Duration(r.AgeMin) * time.Minute)
			if repliedAt.Before(createdAt) {
				repliedAt = createdAt.Add(time.Minute)
			}
			if repliedAt.After(now) {
				repliedAt = now.Add(-time.Minute)
			}
			if err := internal.Exec(database,
				`INSERT INTO story_interacts (story_id, user_id, `+"`type`"+`, content, click_count, created_at)
				 VALUES (?, ?, 'reply', ?, 0, ?)`,
				storyID, replierID, r.Content, repliedAt,
			); err != nil {
				return fmt.Errorf("extended: insert story reply %s: %w", s.Key, err)
			}
		}
	}

	for _, m := range seeddata.StoryMutes {
		muterID, ok := state.UserIDByUsername[m.User]
		if !ok {
			return fmt.Errorf("extended: story mute references unknown user %q", m.User)
		}
		mutedID, ok := state.UserIDByUsername[m.Muted]
		if !ok {
			return fmt.Errorf("extended: story mute references unknown user %q", m.Muted)
		}
		if err := internal.Exec(database,
			`INSERT INTO story_mutes (id, user_id, muted_user_id, created_at) VALUES (?, ?, ?, ?)`,
			internal.UUID(), muterID, mutedID, now.Add(-time.Duration(m.AgeDays)*24*time.Hour),
		); err != nil {
			return fmt.Errorf("extended: insert story mute %s→%s: %w", m.User, m.Muted, err)
		}
	}

	return nil
}

func seedNotifications(database *sql.DB, state *internal.SeedState, now time.Time) error {
	for _, n := range seeddata.Notifications {
		receiverID, ok := state.UserIDByUsername[n.Receiver]
		if !ok {
			return fmt.Errorf("extended: notification receiver %q not found", n.Receiver)
		}

		var senderID any
		if n.Sender != "" {
			id, ok := state.UserIDByUsername[n.Sender]
			if !ok {
				return fmt.Errorf("extended: notification sender %q not found", n.Sender)
			}
			senderID = id
		}

		var redirectPostID any
		if n.PostKey != "" {
			id, ok := state.PostIDByKey[n.PostKey]
			if !ok {
				return fmt.Errorf("extended: notification references unknown post %q", n.PostKey)
			}
			redirectPostID = id
		}

		var redirectUserID any
		if n.UserKey != "" {
			id, ok := state.UserIDByUsername[n.UserKey]
			if !ok {
				return fmt.Errorf("extended: notification references unknown user %q", n.UserKey)
			}
			redirectUserID = id
		}

		createdAt := now.Add(-time.Duration(n.AgeMin) * time.Minute)

		if err := internal.Exec(database,
			`INSERT INTO notifications (id, receiver_id, sender_id, type, redirect_post_id, redirect_user_id,
				redirect_comment_id, content, is_read, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, NULL, ?, ?, ?)`,
			internal.UUID(), receiverID, senderID, n.Type, redirectPostID, redirectUserID,
			n.Content, boolToInt(n.IsRead), createdAt,
		); err != nil {
			return fmt.Errorf("extended: insert notification %s→%s: %w", n.Sender, n.Receiver, err)
		}
	}
	return nil
}

func seedCalls(database *sql.DB, state *internal.SeedState, now time.Time) error {
	for _, c := range seeddata.Calls {
		callerID, ok := state.UserIDByUsername[c.Caller]
		if !ok {
			return fmt.Errorf("extended: call caller %q not found", c.Caller)
		}
		calleeID, ok := state.UserIDByUsername[c.Callee]
		if !ok {
			return fmt.Errorf("extended: call callee %q not found", c.Callee)
		}

		startedAt := now.Add(-time.Duration(c.AgeMin) * time.Minute)
		var endedAt any
		if c.Status == "ended" {
			endedAt = startedAt.Add(time.Duration(c.Duration) * time.Second)
		} else {
			// Cuộc gọi nhỡ / bị từ chối / huỷ kết thúc ngay khi bắt đầu.
			endedAt = startedAt
		}

		mutedCaller, mutedCallee := 0, 0
		if c.MutedOne {
			mutedCallee = 1
		}
		videoCaller, videoCallee := 0, 0
		if c.VideoOn {
			videoCaller = 1
			videoCallee = 1
		}

		if err := internal.Exec(database,
			`INSERT INTO calls (id, caller_id, callee_id, call_type, is_group, status, started_at, ended_at,
				duration, muted_caller, muted_callee, video_enabled_caller, video_enabled_callee, created_at)
			 VALUES (?, ?, ?, ?, 0, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			internal.UUID(), callerID, calleeID, c.Type, c.Status, startedAt, endedAt,
			c.Duration, mutedCaller, mutedCallee, videoCaller, videoCallee, startedAt,
		); err != nil {
			return fmt.Errorf("extended: insert call %s→%s: %w", c.Caller, c.Callee, err)
		}
	}
	return nil
}

// seedMediaQueue đưa vài tệp vào hàng chờ kiểm duyệt để trang admin có việc để làm.
func seedMediaQueue(database *sql.DB, state *internal.SeedState, now time.Time) error {
	for _, u := range seeddata.UploadQueue {
		ownerID, ok := state.UserIDByUsername[u.User]
		if !ok {
			return fmt.Errorf("extended: media upload references unknown user %q", u.User)
		}

		var reason any
		if u.Reason != "" {
			reason = u.Reason
		}

		mediaID := internal.UUID()
		if err := internal.Exec(database,
			`INSERT INTO media (id, user_id, post_id, file_uri, file_type, file_size, duration_seconds,
				status, review_reason, created_at)
			 VALUES (?, ?, NULL, ?, ?, 268435, 0, ?, ?, ?)`,
			mediaID, ownerID, u.URI, u.Type, u.Status, reason,
			now.Add(-time.Duration(u.AgeHours)*time.Hour),
		); err != nil {
			return fmt.Errorf("extended: insert media upload for %s: %w", u.User, err)
		}
		state.MediaIDs = append(state.MediaIDs, mediaID)
	}
	return nil
}

func seedSystemConfigs(database *sql.DB) error {
	configs := []struct{ key, value string }{
		{"site_name", "LinkUp"},
		{"site_description", "Mạng xã hội Việt Nam — nơi bạn bè kể nhau nghe chuyện thường ngày"},
		{"contact_email", "hotro@linkup.vn"},
		{"maintenance_mode", "false"},
		{"allow_registration", "true"},
		{"require_email_verify", "true"},
		{"password_min_length", "8"},
		{"max_login_attempts", "5"},
		{"jwt_expiry_minutes", "15"},
		{"refresh_token_expiry_days", "7"},
		{"default_user_role", "user"},
	}

	for _, c := range configs {
		if err := internal.Exec(database,
			"INSERT INTO system_configs (`key`, `value`) VALUES (?, ?) AS new ON DUPLICATE KEY UPDATE `value` = new.`value`",
			c.key, c.value,
		); err != nil {
			return fmt.Errorf("extended: insert system_configs %s: %w", c.key, err)
		}
	}
	return nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
