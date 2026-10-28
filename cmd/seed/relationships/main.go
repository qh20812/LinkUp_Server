package relationships

import (
	"fmt"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
	"linkup/seeddata"
)

// Run chèn cộng đồng, nội quy, chính sách đóng góp, thành viên, yêu cầu tham gia
// và chat nhóm mặc định của từng cộng đồng từ seeddata.Communities.
func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("relationships: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	communityAdminRoleID := state.RoleIDs[9]
	communityMemberRoleID := state.RoleIDs[10]

	type assignedRole struct{ userID, roleID, scopeID string }
	assigned := make(map[assignedRole]bool)

	addUserRole := func(userID, roleID, scopeID string) error {
		key := assignedRole{userID, roleID, scopeID}
		if assigned[key] {
			return nil
		}
		assigned[key] = true
		return internal.Exec(database,
			`INSERT INTO user_roles (id, user_id, role_id, scope_id, scope_type, assigned_at) VALUES (?, ?, ?, ?, 'community', ?)`,
			internal.UUID(), userID, roleID, scopeID, now,
		)
	}

	userID := func(key string) (string, error) {
		id, ok := state.UserIDByUsername[key]
		if !ok {
			return "", fmt.Errorf("relationships: unknown user %q", key)
		}
		return id, nil
	}

	for i, c := range seeddata.Communities {
		communityID, ok := state.CommunityIDByKey[c.Key]
		if !ok {
			// Bước này chạy trước social nên thường tự cấp ID; nếu đã có thì dùng lại.
			communityID = internal.UUID()
			state.CommunityIDByKey[c.Key] = communityID
		}
		state.CommunityIDs = append(state.CommunityIDs, communityID)

		creatorID, err := userID(c.Creator)
		if err != nil {
			return err
		}

		createdAt := now.Add(-time.Duration(120+i*90) * 24 * time.Hour)

		if err := internal.Exec(database,
			`INSERT INTO communities (id, creator_id, name, description, avatar_uri, background_uri, auto_approve, privacy, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)`,
			communityID, creatorID, c.Name, c.Description, c.Avatar, c.Background,
			boolToInt(c.AutoApprove), c.Privacy, createdAt,
		); err != nil {
			return fmt.Errorf("relationships: insert community %s: %w", c.Name, err)
		}

		if err := internal.Exec(database,
			`INSERT INTO community_policies (
				id, community_id, post_weight, comment_weight, reaction_weight, event_weight,
				top_contributor_threshold, moderator_promotion_threshold, auto_promote_enabled, badge_enabled,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)`,
			internal.UUID(), communityID,
			c.Policy.PostWeight, c.Policy.CommentWeight, c.Policy.ReactionWeight, c.Policy.EventWeight,
			c.Policy.TopThreshold, c.Policy.ModThreshold,
			boolToInt(c.Policy.AutoPromote), boolToInt(c.Policy.BadgeEnabled), createdAt,
		); err != nil {
			return fmt.Errorf("relationships: insert policy for %s: %w", c.Name, err)
		}

		// Nội quy: position đánh lại từ 1 trong từng nhóm.
		categoryPosition := make(map[string]int)
		for _, rule := range c.Rules {
			categoryPosition[rule.Category]++
			if err := internal.Exec(database,
				`INSERT INTO community_rules (id, community_id, category, title, content, position, created_at, updated_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, NULL)`,
				internal.UUID(), communityID, rule.Category, rule.Title, rule.Content,
				categoryPosition[rule.Category], createdAt,
			); err != nil {
				return fmt.Errorf("relationships: insert rule %q for %s: %w", rule.Title, c.Name, err)
			}
		}

		// Thành viên. Members[0] là quản trị viên cộng đồng.
		for m, member := range c.Members {
			memberID, err := userID(member.User)
			if err != nil {
				return err
			}
			joinedAt := now.Add(-time.Duration(60+m*22) * 24 * time.Hour)
			if joinedAt.Before(createdAt) {
				joinedAt = createdAt.Add(time.Hour)
			}

			if err := internal.Exec(database,
				`INSERT INTO group_members (id, community_id, user_id, points, joined_at) VALUES (?, ?, ?, ?, ?)`,
				internal.UUID(), communityID, memberID, member.Points, joinedAt,
			); err != nil {
				return fmt.Errorf("relationships: insert member %s in %s: %w", member.User, c.Name, err)
			}

			roleID := communityMemberRoleID
			if m == 0 {
				roleID = communityAdminRoleID
			}
			if err := addUserRole(memberID, roleID, communityID); err != nil {
				return fmt.Errorf("relationships: assign role %s in %s: %w", member.User, c.Name, err)
			}
		}

		if c.CreateChat {
			chatID := internal.UUID()
			state.CommunityChatIDs = append(state.CommunityChatIDs, chatID)

			if err := internal.Exec(database,
				`INSERT INTO chats (id, type, name, avatar_uri, encryption_key, community_id, created_at)
				 VALUES (?, 'group', ?, ?, 'seed-enc-key', ?, ?)`,
				chatID, c.ChatName, c.ChatAvatar, communityID, createdAt,
			); err != nil {
				return fmt.Errorf("relationships: insert community chat for %s: %w", c.Name, err)
			}

			for m, member := range c.Members {
				memberID, err := userID(member.User)
				if err != nil {
					return err
				}
				chatRole := "CHAT_MEMBER"
				if m == 0 {
					chatRole = "CHAT_ADMIN"
				}
				if err := internal.Exec(database,
					`INSERT INTO chat_participants (id, chat_id, user_id, role, joined_at) VALUES (?, ?, ?, ?, ?)`,
					internal.UUID(), chatID, memberID, chatRole, createdAt.Add(time.Minute),
				); err != nil {
					return fmt.Errorf("relationships: insert chat participant %s for %s: %w", member.User, c.Name, err)
				}
			}
		}
	}

	// Yêu cầu tham gia đang chờ duyệt.
	for _, r := range seeddata.JoinRequests {
		communityID, ok := state.CommunityIDByKey[r.Community]
		if !ok {
			return fmt.Errorf("relationships: join request references unknown community %q", r.Community)
		}
		requesterID, err := userID(r.User)
		if err != nil {
			return err
		}
		createdAt := now.Add(-time.Duration(r.AgeDays) * 24 * time.Hour)
		if err := internal.Exec(database,
			`INSERT INTO community_join_requests (id, community_id, user_id, status, created_at, responded_at)
			 VALUES (?, ?, ?, ?, ?, NULL)`,
			internal.UUID(), communityID, requesterID, r.Status, createdAt,
		); err != nil {
			return fmt.Errorf("relationships: insert join request %s→%s: %w", r.User, r.Community, err)
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
