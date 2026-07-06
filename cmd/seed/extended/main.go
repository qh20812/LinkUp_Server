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
		{0, 0, "https://cdn.example.com/uploads/photo1.jpg", "image/jpeg", 2048576},
		{1, 1, "https://cdn.example.com/uploads/photo2.jpg", "image/jpeg", 1543200},
		{2, 2, "https://cdn.example.com/uploads/video1.mp4", "video/mp4", 52428800},
		{3, 3, "https://cdn.example.com/uploads/doc1.pdf", "application/pdf", 1024000},
		{4, 4, "https://cdn.example.com/uploads/photo3.png", "image/png", 3850240},
		{5, 5, "https://cdn.example.com/uploads/audio1.mp3", "audio/mpeg", 7680000},
		{6, 6, "https://cdn.example.com/uploads/photo4.jpg", "image/jpeg", 1254400},
		{7, 7, "https://cdn.example.com/uploads/photo5.gif", "image/gif", 512000},
		{8, 8, "https://cdn.example.com/uploads/video2.mp4", "video/mp4", 104857600},
		{0, 9, "https://cdn.example.com/uploads/photo6.jpg", "image/jpeg", 2891000},
		{2, 10, "https://cdn.example.com/uploads/photo7.png", "image/png", 4560000},
		{5, 11, "https://cdn.example.com/uploads/photo8.jpg", "image/jpeg", 1843200},
		{9, 12, "https://cdn.example.com/uploads/photo9.jpg", "image/jpeg", 900000},
		{12, 13, "https://cdn.example.com/uploads/doc2.pdf", "application/pdf", 2048000},
		{15, 14, "https://cdn.example.com/uploads/video3.mp4", "video/mp4", 33554432},
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
		mediaIdx  int
	}{
		{"Boost Your Productivity", "Try our new project management tool", "https://example.com/productivity", 5000.00, "active", 0},
		{"Learn Go Programming", "Master Go in 30 days", "https://example.com/learn-go", 2500.00, "active", 1},
		{"Cloud Hosting Sale", "50% off your first year", "https://example.com/cloud-hosting", 10000.00, "active", 4},
		{"AI-Powered Analytics", "Understand your users better", "https://example.com/analytics", 7500.00, "paused", 6},
		{"DevOps Conference 2026", "Early bird tickets available", "https://example.com/devops-conf", 3000.00, "completed", -1},
	}

	partnerUserID := state.UserIDs[2]

	for _, a := range ads {
		adID := internal.UUID()
		var mediaID *string
		if a.mediaIdx >= 0 && a.mediaIdx < len(state.MediaIDs) {
			mediaID = internal.Ptr(state.MediaIDs[a.mediaIdx])
		}
		startedAt := internal.PtrTime(now.Add(-time.Duration(randRange(1, 30)) * 24 * time.Hour))
		expiresAt := internal.PtrTime(now.Add(time.Duration(randRange(15, 60)) * 24 * time.Hour))

		if err := internal.Exec(database,
			`INSERT INTO ads (id, partner_id, title, content, media_id, target_url, status, budget, started_at, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			adID, partnerUserID, a.title, a.content, mediaID, a.targetURL, a.status, a.budget, startedAt, expiresAt, now.Add(-time.Duration(randRange(1, 60))*24*time.Hour),
		); err != nil {
			return fmt.Errorf("extended: insert ad %s: %w", a.title, err)
		}

		actionTypes := []string{"impression", "impression", "click", "conversion"}
		for j := 0; j < randRange(1, 4); j++ {
			userID := internal.Ptr(state.UserIDs[randRange(0, len(state.UserIDs)-1)])
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
		userID := state.UserIDs[i%len(state.UserIDs)]
		mediaType := "image"
		if i%4 == 0 {
			mediaType = "video"
		}
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
			fmt.Sprintf("https://cdn.example.com/stories/story%d.jpg", i),
			mediaType, caption, now.Add(-time.Duration(randRange(1, 12))*time.Hour), expiresAt,
		); err != nil {
			return fmt.Errorf("extended: insert story %d: %w", i, err)
		}
	}

	notifTypes := []string{"like", "comment", "follow", "message", "friend_request"}
	notifContents := []string{
		"liked your post",
		"commented on your post",
		"started following you",
		"sent you a message",
		"sent you a friend request",
	}

	for i := 0; i < 20; i++ {
		receiverID := state.UserIDs[i%len(state.UserIDs)]
		senderID := internal.Ptr(state.UserIDs[(i+2)%len(state.UserIDs)])
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
		callerID := state.UserIDs[i%len(state.UserIDs)]
		calleeID := state.UserIDs[(i+1)%len(state.UserIDs)]
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

	return nil
}
