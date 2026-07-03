package relationships

import (
	"fmt"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
)

func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("relationships: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	superAdminID := state.UserIDs[0]

	communityAdminRoleID := state.RoleIDs[8]
	communityMemberRoleID := state.RoleIDs[9]
	groupAdminRoleID := state.RoleIDs[5]
	groupModRoleID := state.RoleIDs[6]
	groupMemberRoleID := state.RoleIDs[7]

	userRoleAssigned := map[string]bool{}

	addUserRole := func(userID, roleID, scopeID string) error {
		key := userID + "|" + roleID + "|" + scopeID
		if userRoleAssigned[key] {
			return nil
		}
		userRoleAssigned[key] = true
		return internal.Exec(database,
			`INSERT INTO user_roles (id, user_id, role_id, scope_id, scope_type, assigned_at) VALUES (?, ?, ?, ?, 'community', ?)`,
			internal.UUID(), userID, roleID, scopeID, now,
		)
	}

	type communityData struct {
		id          string
		name        string
		description string
		creatorIdx  int
	}

	communities := []communityData{
		{internal.UUID(), "Go Developers", "A community for Go programming enthusiasts", 2},
		{internal.UUID(), "Cloud Architects", "Discussing cloud infrastructure and architecture", 4},
		{internal.UUID(), "AI Enthusiasts", "Exploring artificial intelligence and machine learning", 3},
		{internal.UUID(), "DevOps United", "Sharing DevOps best practices and tools", 5},
		{internal.UUID(), "Open Source Contributors", "Collaborating on open source projects", 8},
	}

	for i, c := range communities {
		if err := internal.Exec(database,
			`INSERT INTO communities (id, creator_id, name, description, avatar_uri, background_uri, created_at, updated_at) VALUES (?, ?, ?, ?, ?, '', ?, NULL)`,
			c.id, state.UserIDs[c.creatorIdx], c.name, c.description,
			fmt.Sprintf("https://api.dicebear.com/7.x/identicon/svg?seed=community%d", i),
			now,
		); err != nil {
			return fmt.Errorf("relationships: insert community %s: %w", c.name, err)
		}
		state.CommunityIDs = append(state.CommunityIDs, c.id)

		if err := internal.Exec(database,
			`INSERT INTO community_policies (
				id, community_id, post_weight, comment_weight, reaction_weight, event_weight,
				top_contributor_threshold, moderator_promotion_threshold, auto_promote_enabled, badge_enabled,
				created_at, updated_at
			) VALUES (?, ?, 10, 5, 2, 20, 2500, 5000, 1, 1, ?, NULL)`,
			internal.UUID(), c.id, now,
		); err != nil {
			return fmt.Errorf("relationships: insert community policy for %s: %w", c.name, err)
		}

		if err := addUserRole(state.UserIDs[c.creatorIdx], communityAdminRoleID, c.id); err != nil {
			return fmt.Errorf("relationships: community_admin user_role for %s: %w", state.UserIDs[c.creatorIdx], err)
		}

		rules := []struct {
			category string
			title    string
			content  string
			position int
		}{
			{"conduct", "Tôn trọng lẫn nhau", "Hãy luôn giữ thái độ tôn trọng với các thành viên khác. Không công kích cá nhân, quấy rối hoặc phân biệt đối xử.", 1},
			{"conduct", "Giữ văn minh", "Thảo luận lành mạnh, không spam, không flood tin nhắn.", 2},
			{"prohibited", "Nội dung bất hợp pháp", "Không đăng tải nội dung vi phạm pháp luật, bao gồm bản quyền và sở hữu trí tuệ.", 1},
			{"prohibited", "Quảng cáo trái phép", "Không đăng quảng cáo hoặc link giới thiệu khi chưa được cho phép.", 2},
			{"guidelines", "Cách đăng bài", "Đăng bài đúng chủ đề, tiêu đề rõ ràng, sử dụng tag phù hợp.", 1},
			{"guidelines", "Đóng góp tích cực", "Chia sẻ kiến thức, hỗ trợ thành viên mới, báo cáo nội dung vi phạm.", 2},
		}
		for _, rule := range rules {
			if err := internal.Exec(database,
				`INSERT INTO community_rules (id, community_id, category, title, content, position, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, NULL)`,
				internal.UUID(), c.id, rule.category, rule.title, rule.content, rule.position, now,
			); err != nil {
				return fmt.Errorf("relationships: insert rule %s for community %s: %w", rule.title, c.name, err)
			}
		}
	}

	type memberData struct {
		communityIdx int
		userIdx      int
		role         string
	}

	members := []memberData{}
	for ci := range communities {
		adminIdx := ci*2 + 2
		modIdx := ci*2 + 3
		members = append(members, memberData{communityIdx: ci, userIdx: adminIdx, role: "GROUP_ADMIN"})
		members = append(members, memberData{communityIdx: ci, userIdx: modIdx, role: "GROUP_MOD"})
		for m := 0; m < 3; m++ {
			idx := (ci*5+m+4)%19 + 1
			members = append(members, memberData{communityIdx: ci, userIdx: idx, role: "GROUP_MEMBER"})
		}
	}

	for _, m := range members {
		userID := state.UserIDs[m.userIdx]
		if userID == superAdminID {
			continue
		}

		points := 0
		if m.role == "GROUP_ADMIN" {
			points = 500
		} else if m.role == "GROUP_MOD" {
			points = 200
		} else {
			points = 10
		}

		if err := internal.Exec(database,
			`INSERT INTO group_members (id, community_id, user_id, points, joined_at) VALUES (?, ?, ?, ?, ?)`,
			internal.UUID(), state.CommunityIDs[m.communityIdx], userID, points, now,
		); err != nil {
			return fmt.Errorf("relationships: insert member for community %d user %d: %w", m.communityIdx, m.userIdx, err)
		}

		var roleID string
		switch m.role {
		case "GROUP_ADMIN":
			roleID = groupAdminRoleID
		case "GROUP_MOD":
			roleID = groupModRoleID
		default:
			roleID = groupMemberRoleID
		}
		communityID := state.CommunityIDs[m.communityIdx]

		if err := addUserRole(userID, roleID, communityID); err != nil {
			return fmt.Errorf("relationships: user_role %s for user %s: %w", m.role, userID, err)
		}

		if err := addUserRole(userID, communityMemberRoleID, communityID); err != nil {
			return fmt.Errorf("relationships: community_member user_role for %s: %w", userID, err)
		}
	}

	return nil
}
