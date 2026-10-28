package moderation

import (
	"fmt"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
	"linkup/seeddata"
)

// Run chèn báo cáo vi phạm, án khoá tài khoản và nhật ký kiểm duyệt từ seeddata.
func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("moderation: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	userID := func(key string) (string, error) {
		id, ok := state.UserIDByUsername[key]
		if !ok {
			return "", fmt.Errorf("moderation: unknown user %q", key)
		}
		return id, nil
	}

	for i, r := range seeddata.Reports {
		reporterID, err := userID(r.Reporter)
		if err != nil {
			return err
		}

		var targetUserID, targetPostID, targetCommunityID any
		switch {
		case r.TargetUser != "":
			id, err := userID(r.TargetUser)
			if err != nil {
				return err
			}
			targetUserID = id
		case r.TargetPost != "":
			id, ok := state.PostIDByKey[r.TargetPost]
			if !ok {
				return fmt.Errorf("moderation: report references unknown post %q", r.TargetPost)
			}
			targetPostID = id
		case r.TargetCommunity != "":
			id, ok := state.CommunityIDByKey[r.TargetCommunity]
			if !ok {
				return fmt.Errorf("moderation: report references unknown community %q", r.TargetCommunity)
			}
			targetCommunityID = id
		}

		var ruleID any
		if r.RuleTitle != "" {
			id, ok := state.ViolationRuleIDByTitle[r.RuleTitle]
			if !ok {
				return fmt.Errorf("moderation: report references unknown rule %q", r.RuleTitle)
			}
			ruleID = id
		}

		createdAt := now.Add(-time.Duration(r.AgeHours * float64(time.Hour)))
		if err := internal.Exec(database,
			`INSERT INTO reports (id, reporter_id, report_type, target_user_id, target_post_id, target_community_id,
				violation_rule_id, reason_detail, status, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			internal.UUID(), reporterID, r.Type, targetUserID, targetPostID, targetCommunityID,
			ruleID, r.Detail, r.Status, createdAt, createdAt,
		); err != nil {
			return fmt.Errorf("moderation: insert report %d: %w", i, err)
		}
	}

	for _, b := range seeddata.Bans {
		targetID, err := userID(b.User)
		if err != nil {
			return err
		}
		adminID, err := userID(b.Admin)
		if err != nil {
			return err
		}
		createdAt := now.Add(-time.Duration(b.AgeDays) * 24 * time.Hour)

		var expiresAt any
		if b.DurationDays > 0 {
			expiresAt = createdAt.Add(time.Duration(b.DurationDays) * 24 * time.Hour)
		}

		if err := internal.Exec(database,
			`INSERT INTO bans (id, user_id, admin_id, reason, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
			internal.UUID(), targetID, adminID, b.Reason, expiresAt, createdAt,
		); err != nil {
			return fmt.Errorf("moderation: insert ban for %s: %w", b.User, err)
		}
	}

	for _, l := range seeddata.ModLogs {
		moderatorID, err := userID(l.Moderator)
		if err != nil {
			return err
		}

		var targetID string
		switch l.TargetType {
		case "post":
			id, ok := state.PostIDByKey[l.TargetKey]
			if !ok {
				return fmt.Errorf("moderation: log references unknown post %q", l.TargetKey)
			}
			targetID = id
		case "user":
			id, err := userID(l.TargetKey)
			if err != nil {
				return err
			}
			targetID = id
		case "community":
			id, ok := state.CommunityIDByKey[l.TargetKey]
			if !ok {
				return fmt.Errorf("moderation: log references unknown community %q", l.TargetKey)
			}
			targetID = id
		default:
			return fmt.Errorf("moderation: log has unsupported target type %q", l.TargetType)
		}

		createdAt := now.Add(-time.Duration(l.AgeDays) * 24 * time.Hour)
		if err := internal.Exec(database,
			`INSERT INTO moderation_logs (id, moderator_id, action, target_type, target_id, reason, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			internal.UUID(), moderatorID, l.Action, l.TargetType, targetID, l.Reason, createdAt,
		); err != nil {
			return fmt.Errorf("moderation: insert log %s/%s: %w", l.Action, l.TargetKey, err)
		}
	}

	return nil
}
