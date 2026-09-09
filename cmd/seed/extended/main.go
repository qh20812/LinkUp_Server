package extended

import (
	"fmt"
	"math/rand"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func pick[T any](items []T) T {
	return items[rng.Intn(len(items))]
}

func randRange(min, max int) int {
	return rng.Intn(max-min+1) + min
}

func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("extended: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	mediaData := []struct {
		userIdx  int
		postIdx  int
		fileURI  string
		fileType string
		fileSize float64
	}{
		{2, 0, "https://picsum.photos/seed/media0/800/600", "image/jpeg", 2048576},
		{3, 1, "https://picsum.photos/seed/media1/800/600", "image/jpeg", 1543200},
		{2, 2, "https://www.w3schools.com/html/mov_bbb.mp4", "video/mp4", 52428800},
		{3, 3, "https://picsum.photos/seed/media3/800/600", "image/jpeg", 1024000},
		{4, 4, "https://picsum.photos/seed/media4/800/600", "image/png", 3850240},
		{5, 5, "https://media.w3.org/2010/05/sintel/trailer.mp4", "video/mp4", 7680000},
		{6, 6, "https://picsum.photos/seed/media6/800/600", "image/jpeg", 1254400},
		{7, 7, "https://picsum.photos/seed/media7/400/400", "image/jpeg", 512000},
		{8, 8, "https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/720/Big_Buck_Bunny_720_10s_1MB.mp4", "video/mp4", 104857600},
		{9, 9, "https://picsum.photos/seed/media9/800/600", "image/jpeg", 2891000},
		{2, 10, "https://picsum.photos/seed/media10/800/600", "image/png", 4560000},
		{5, 11, "https://picsum.photos/seed/media11/800/600", "image/jpeg", 1843200},
		{9, 12, "https://picsum.photos/seed/media12/800/600", "image/jpeg", 900000},
		{12, 13, "https://picsum.photos/seed/media13/800/600", "image/jpeg", 2048000},
		{15, 14, "https://www.w3schools.com/html/mov_bbb.mp4", "video/mp4", 33554432},
	}

	for _, m := range mediaData {
		mediaID := internal.UUID()
		postID := internal.Ptr(state.PostIDs[m.postIdx%len(state.PostIDs)])
		status := pick([]string{"pending", "approved", "approved", "approved", "rejected"})

		if err := internal.Exec(database,
			`INSERT INTO media (id, user_id, post_id, file_uri, file_type, file_size, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			mediaID, state.UserIDs[m.userIdx], postID, m.fileURI, m.fileType, m.fileSize, status, now.Add(-time.Duration(randRange(1, 720))*time.Hour),
		); err != nil {
			return fmt.Errorf("extended: insert media: %w", err)
		}
		state.MediaIDs = append(state.MediaIDs, mediaID)
	}

	ads := []struct {
		title     string
		content   string
		targetURL string
		budget    float64
		status    string
	}{
		{"Boost Your Productivity", "Try our new project management tool", "https://example.com/productivity", 5000.00, "active"},
		{"Learn Go Programming", "Master Go in 30 days", "https://example.com/learn-go", 2500.00, "active"},
		{"Cloud Hosting Sale", "50% off your first year", "https://example.com/cloud-hosting", 10000.00, "active"},
		{"AI-Powered Analytics", "Understand your users better", "https://example.com/analytics", 7500.00, "paused"},
		{"DevOps Conference 2026", "Early bird tickets available", "https://example.com/devops-conf", 3000.00, "completed"},
	}

	partnerUserID := state.UserIDs[2]

	for _, a := range ads {
		adID := internal.UUID()
		startedAt := internal.PtrTime(now.Add(-time.Duration(randRange(1, 30)) * 24 * time.Hour))
		expiresAt := internal.PtrTime(now.Add(time.Duration(randRange(15, 60)) * 24 * time.Hour))

		if err := internal.Exec(database,
			`INSERT INTO ads (id, partner_id, title, content, target_url, status, budget, started_at, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			adID, partnerUserID, a.title, a.content, a.targetURL, a.status, a.budget, startedAt, expiresAt, now.Add(-time.Duration(randRange(1, 60))*24*time.Hour),
		); err != nil {
			return fmt.Errorf("extended: insert ad %s: %w", a.title, err)
		}

		actionTypes := []string{"impression", "impression", "click", "conversion"}
		for j := 0; j < randRange(1, 4); j++ {
			userID := internal.Ptr(state.UserIDs[randRange(2, len(state.UserIDs)-1)])
			if err := internal.Exec(database,
				`INSERT INTO ad_analytics (id, ad_id, user_id, action_type, ip_address, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
				internal.UUID(), adID, userID, pick(actionTypes),
				fmt.Sprintf("192.168.%d.%d", randRange(1, 254), randRange(1, 254)),
				now.Add(-time.Duration(randRange(1, 72))*time.Hour),
			); err != nil {
				return fmt.Errorf("extended: insert ad_analytics: %w", err)
			}
		}
	}

	for i := 0; i < 10; i++ {
		userID := state.UserIDs[internal.ContentIndex(i)]
		caption := pick([]string{
			"Beautiful morning!",
			"Check out this view 🌄",
			"New project update",
			"Team lunch today!",
			"Late night coding session",
			"Weekend vibes",
			"Throwback to last conference",
			"Happy birthday to me!",
			"Just finished this book 📚",
			"Welcome to the team!",
		})
		expiresAt := internal.PtrTime(now.Add(24 * time.Hour))

		if err := internal.Exec(database,
			`INSERT INTO stories (id, user_id, media_uri, media_type, caption, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			internal.UUID(), userID,
			fmt.Sprintf("https://picsum.photos/seed/story%d/600/900", i),
			"image", caption, now.Add(-time.Duration(randRange(1, 12))*time.Hour), expiresAt,
		); err != nil {
			return fmt.Errorf("extended: insert story %d: %w", i, err)
		}
	}

	notifTypes := []string{"like", "comment", "follow", "message", "friend_request"}
	notifContents := []string{
		"đã thích bài viết của bạn",
		"đã bình luận bài viết của bạn",
		"đã theo dõi bạn",
		"đã gửi tin nhắn",
		"đã gửi lời mời kết bạn",
	}

	for i := 0; i < 20; i++ {
		// Loại trừ superadmin (index 0) và admin (index 1) khỏi receiver
		receiverID := state.UserIDs[i%18+2]
		senderID := internal.Ptr(state.UserIDs[(i+2)%18+2])
		notifType := pick(notifTypes)
		isRead := rng.Intn(3) > 0

		if err := internal.Exec(database,
			`INSERT INTO notifications (id, receiver_id, sender_id, type, redirect_post_id, redirect_user_id, redirect_comment_id, content, is_read, created_at) VALUES (?, ?, ?, ?, NULL, NULL, NULL, ?, ?, ?)`,
			internal.UUID(), receiverID, senderID, notifType, pick(notifContents), isRead, now.Add(-time.Duration(randRange(1, 168))*time.Hour),
		); err != nil {
			return fmt.Errorf("extended: insert notification %d: %w", i, err)
		}
	}

	callTypes := []string{"voice", "video"}
	callStatuses := []string{"ended", "ended", "missed", "rejected"}

	for i := 0; i < 5; i++ {
		callerID := state.UserIDs[internal.ContentIndex(i)]
		calleeID := state.UserIDs[internal.ContentIndex(i+1)]
		callType := pick(callTypes)
		status := pick(callStatuses)
		startedAt := now.Add(-time.Duration(randRange(1, 48)) * time.Hour)
		var endedAt *time.Time
		if status == "ended" {
			endedAt = internal.PtrTime(startedAt.Add(time.Duration(randRange(60, 3600)) * time.Second))
		}

		if err := internal.Exec(database,
			`INSERT INTO calls (id, caller_id, callee_id, call_type, is_group, status, started_at, ended_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			internal.UUID(), callerID, calleeID, callType, false, status, startedAt, endedAt,
		); err != nil {
			return fmt.Errorf("extended: insert call %d: %w", i, err)
		}
	}

	systemConfigs := []struct {
		key   string
		value string
	}{
		{"site_name", "LinkUp"},
		{"site_description", "Mạng xã hội Việt"},
		{"contact_email", "admin@linkup.com"},
		{"maintenance_mode", "false"},
		{"allow_registration", "true"},
		{"require_email_verify", "true"},
		{"password_min_length", "8"},
		{"max_login_attempts", "5"},
		{"jwt_expiry_minutes", "15"},
		{"default_user_role", "user"},
	}

	for _, sc := range systemConfigs {
		if err := internal.Exec(database,
			"INSERT INTO system_configs (`key`, `value`) VALUES (?, ?) AS new ON DUPLICATE KEY UPDATE `value` = new.`value`",
			sc.key, sc.value,
		); err != nil {
			return fmt.Errorf("extended: insert system_configs %s: %w", sc.key, err)
		}
	}

	return nil
}
