package core

import (
	"fmt"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
)

func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("core: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	type roleEntry struct {
		id          string
		name        string
		description string
	}

	roles := []roleEntry{
		{internal.UUID(), "SUPER_ADMIN", "Full system access"},
		{internal.UUID(), "ADMIN", "Administrative access"},
		{internal.UUID(), "PARTNER", "Ads partner access"},
		{internal.UUID(), "USER", "Standard user access"},
		{internal.UUID(), "CHAT_ADMIN", "Chat administrator"},
		{internal.UUID(), "CHAT_MEMBER", "Chat member"},
		{internal.UUID(), "GROUP_ADMIN", "Group administrator"},
		{internal.UUID(), "GROUP_MOD", "Group moderator"},
		{internal.UUID(), "GROUP_MEMBER", "Group member"},
		{internal.UUID(), "COMMUNITY_ADMIN", "Community administrator"},
		{internal.UUID(), "COMMUNITY_MEMBER", "Community member"},
	}

	for _, r := range roles {
		if err := internal.Exec(database,
			`INSERT INTO roles (id, name, description) VALUES (?, ?, ?)`,
			r.id, r.name, r.description,
		); err != nil {
			return fmt.Errorf("core: insert role %s: %w", r.name, err)
		}
		state.RoleIDs = append(state.RoleIDs, r.id)
	}

	type emoji struct {
		id       string
		code     string
		imageURI string
	}

	emojis := []emoji{
		{internal.UUID(), ":like:", "https://cdn.jsdelivr.net/gh/jdecked/twemoji@15.1.0/assets/72x72/1f44d.png"},
		{internal.UUID(), ":love:", "https://cdn.jsdelivr.net/gh/jdecked/twemoji@15.1.0/assets/72x72/2764.png"},
		{internal.UUID(), ":haha:", "https://cdn.jsdelivr.net/gh/jdecked/twemoji@15.1.0/assets/72x72/1f602.png"},
		{internal.UUID(), ":wow:", "https://cdn.jsdelivr.net/gh/jdecked/twemoji@15.1.0/assets/72x72/1f62e.png"},
		{internal.UUID(), ":sad:", "https://cdn.jsdelivr.net/gh/jdecked/twemoji@15.1.0/assets/72x72/1f622.png"},
		{internal.UUID(), ":angry:", "https://cdn.jsdelivr.net/gh/jdecked/twemoji@15.1.0/assets/72x72/1f621.png"},
		{internal.UUID(), ":clap:", "https://cdn.jsdelivr.net/gh/jdecked/twemoji@15.1.0/assets/72x72/1f44f.png"},
		{internal.UUID(), ":fire:", "https://cdn.jsdelivr.net/gh/jdecked/twemoji@15.1.0/assets/72x72/1f525.png"},
		{internal.UUID(), ":heart:", "https://cdn.jsdelivr.net/gh/jdecked/twemoji@15.1.0/assets/72x72/1f496.png"},
		{internal.UUID(), ":rocket:", "https://cdn.jsdelivr.net/gh/jdecked/twemoji@15.1.0/assets/72x72/1f680.png"},
	}

	for _, e := range emojis {
		if err := internal.Exec(database,
			`INSERT INTO emojis (id, code, image_uri) VALUES (?, ?, ?)`,
			e.id, e.code, e.imageURI,
		); err != nil {
			return fmt.Errorf("core: insert emoji %s: %w", e.code, err)
		}
		state.EmojiIDs = append(state.EmojiIDs, e.id)
	}

	type violationRule struct {
		id           string
		title        string
		description  string
		applicableTo string
		severity     string
		sortOrder    int
	}

	rules := []violationRule{
		{internal.UUID(), "Spam / quảng cáo làm phiền", "Đăng nội dung quảng cáo lặp lại, link rác hoặc nội dung không liên quan.", "all", "low", 1},
		{internal.UUID(), "Quấy rối / bắt nạt", "Tấn công, đe dọa hoặc xúc phạm người khác.", "all", "high", 2},
		{internal.UUID(), "Ngôn từ thù ghét / phân biệt đối xử", "Nội dung kỳ thị chủng tộc, giới tính, tôn giáo hoặc nhóm người.", "all", "high", 3},
		{internal.UUID(), "Nội dung nhạy cảm / đồi trụy", "Hình ảnh, video hoặc mô tả mang tính khiêu dâm, bạo lực quá mức.", "all", "high", 4},
		{internal.UUID(), "Thông tin sai lệch", "Tin giả, thông tin sai sự thật gây hiểu lầm.", "post", "medium", 5},
		{internal.UUID(), "Lừa đảo / giả mạo", "Tài khoản giả mạo người khác hoặc hành vi lừa đảo.", "user", "high", 6},
		{internal.UUID(), "Vi phạm bản quyền", "Đăng lại nội dung có bản quyền mà không được phép.", "post", "medium", 7},
		{internal.UUID(), "Lý do khác", "Vi phạm không thuộc các nhóm trên (mô tả chi tiết ở lý do).", "all", "low", 99},
	}

	now = time.Now().UTC()
	for _, v := range rules {
		if err := internal.Exec(database,
			`INSERT INTO violation_rules (id, title, description, applicable_to, severity, sort_order, is_active, created_at) VALUES (?, ?, ?, ?, ?, ?, 1, ?)`,
			v.id, v.title, v.description, v.applicableTo, v.severity, v.sortOrder, now,
		); err != nil {
			return fmt.Errorf("core: insert violation_rule %s: %w", v.title, err)
		}
		state.ViolationRuleIDs = append(state.ViolationRuleIDs, v.id)
	}

	superAdminID := state.RoleIDs[0]
	adminID := state.RoleIDs[1]
	partnerID := state.RoleIDs[2]
	userRoleID := state.RoleIDs[3]

	assigned := map[string]bool{}

	addRole := func(userID, roleID string) error {
		key := userID + "|" + roleID
		if assigned[key] {
			return nil
		}
		assigned[key] = true
		return internal.Exec(database,
			`INSERT INTO user_roles (id, user_id, role_id, scope_id, scope_type, assigned_at) VALUES (?, ?, ?, NULL, NULL, ?)`,
			internal.UUID(), userID, roleID, now,
		)
	}

	for i, uid := range state.UserIDs {
		if i == 0 {
			if err := addRole(uid, superAdminID); err != nil {
				return fmt.Errorf("core: user_role super_admin for %s: %w", uid, err)
			}
			continue
		}
		if i == 1 {
			if err := addRole(uid, adminID); err != nil {
				return fmt.Errorf("core: user_role admin for %s: %w", uid, err)
			}
			continue
		}
		if i == 2 {
			if err := addRole(uid, partnerID); err != nil {
				return fmt.Errorf("core: user_role partner for %s: %w", uid, err)
			}
		}
		if err := addRole(uid, userRoleID); err != nil {
			return fmt.Errorf("core: user_role user for %s: %w", uid, err)
		}
	}

	for _, uid := range state.UserIDs {
		if err := internal.Exec(database,
			`INSERT INTO notification_preferences (user_id, like_enabled, comment_enabled, follow_enabled, message_enabled, friend_request_enabled) VALUES (?, 1, 1, 1, 1, 1)`,
			uid,
		); err != nil {
			return fmt.Errorf("core: insert notification_preference for %s: %w", uid, err)
		}
	}

	return nil
}
