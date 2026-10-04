package moderation

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

func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("moderation: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	// Dùng rule đã seed ở step core (state.ViolationRuleIDs), thứ tự:
	// 0 spam(all), 1 harassment(all), 2 hate(all), 3 nudity(all),
	// 4 misinfo(post), 5 scam(user), 6 copyright(post), 7 other(all).
	ruleIDs := state.ViolationRuleIDs

	reportTypes := []string{"spam", "harassment", "hate_speech", "nudity", "copyright"}
	reportDetails := []string{
		"User keeps posting the same link in every thread",
		"Offensive comments directed at another user",
		"Discriminatory language used in multiple posts",
		"Inappropriate image shared publicly",
		"Reposting content without attribution",
		"Threatening behavior in DMs",
		"Fake account impersonating a real person",
		"Misleading information shared as fact",
	}

	for i := 0; i < 8; i++ {
		reporterID := state.UserIDs[i%18+2]

		var targetUserID *string
		var targetPostID *string
		var targetCommentID *string

		switch i % 3 {
		case 0:
			targetUserID = internal.Ptr(state.UserIDs[(i+3)%len(state.UserIDs)])
		case 1:
			if len(state.PostIDs) > 0 {
				targetPostID = internal.Ptr(state.PostIDs[i%len(state.PostIDs)])
			}
		case 2:
			if len(state.CommentIDs) > 0 {
				targetCommentID = internal.Ptr(state.CommentIDs[i%len(state.CommentIDs)])
			}
		}

		status := pick([]string{"pending", "reviewed", "resolved"})

		// Gán ngẫu nhiên rule phù hợp target (70% có rule, 30% NULL để test optional).
		var violationRuleID *string
		if rng.Intn(10) < 7 && len(ruleIDs) > 0 {
			target := []string{"user", "post", "comment"}[i%3]
			candidates := ruleIDs
			if target == "post" {
				candidates = []string{ruleIDs[0], ruleIDs[1], ruleIDs[2], ruleIDs[3], ruleIDs[4], ruleIDs[6], ruleIDs[7]}
			} else if target == "user" {
				candidates = []string{ruleIDs[0], ruleIDs[1], ruleIDs[2], ruleIDs[3], ruleIDs[5], ruleIDs[7]}
			} else {
				candidates = []string{ruleIDs[0], ruleIDs[1], ruleIDs[2], ruleIDs[3], ruleIDs[7]}
			}
			violationRuleID = &candidates[rng.Intn(len(candidates))]
		}

		if err := internal.Exec(database,
			`INSERT INTO reports (id, reporter_id, report_type, target_user_id, target_post_id, target_comment_id, violation_rule_id, reason_detail, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			internal.UUID(), reporterID, pick(reportTypes), targetUserID, targetPostID, targetCommentID, violationRuleID, pick(reportDetails), status, now.Add(-time.Duration(rng.Intn(168))*time.Hour),
		); err != nil {
			return fmt.Errorf("moderation: insert report %d: %w", i, err)
		}
	}

	banReasons := []string{
		"Repeated violations of community guidelines",
		"Hate speech and harassment",
		"Spamming across multiple channels",
		"Impersonating an admin",
		"Sharing illegal content",
	}

	for i := 0; i < 5; i++ {
		userID := state.UserIDs[(i+6)%len(state.UserIDs)]
		adminID := state.UserIDs[0]
		expiresAt := internal.PtrTime(now.Add(30 * 24 * time.Hour))

		if err := internal.Exec(database,
			`INSERT INTO bans (id, user_id, admin_id, reason, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
			internal.UUID(), userID, adminID, banReasons[i], expiresAt, now.Add(-time.Duration(rng.Intn(72))*time.Hour),
		); err != nil {
			return fmt.Errorf("moderation: insert ban %d: %w", i, err)
		}
	}

	actions := []string{"warn", "ban", "mute", "suspend", "delete"}
	targetTypes := []string{"user", "post", "comment", "ad", "report"}

	for i := 0; i < 8; i++ {
		moderatorID := state.UserIDs[0]
		if i%2 == 0 {
			moderatorID = state.UserIDs[1]
		}
		action := pick(actions)
		targetType := pick(targetTypes)
		targetID := state.UserIDs[i%18+2]
		reason := pick([]string{
			"Violation of terms of service",
			"Reported by multiple users",
			"Automated detection flagged content",
			"Manual review by moderation team",
		})

		if err := internal.Exec(database,
			`INSERT INTO moderation_logs (id, moderator_id, action, target_type, target_id, reason, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			internal.UUID(), moderatorID, action, targetType, targetID, reason, now.Add(-time.Duration(rng.Intn(168))*time.Hour),
		); err != nil {
			return fmt.Errorf("moderation: insert log %d: %w", i, err)
		}
	}

	return nil
}
