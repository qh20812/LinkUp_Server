package core

import (
	"fmt"
	"strings"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
	"linkup/seeddata"
)

// boolToTiny chuyển bool sang 0/1 cho cột TINYINT(1).
func boolToTiny(b bool) int {
	if b {
		return 1
	}
	return 0
}

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
		id         string
		code       string
		imageURI   string
		character  string
		name       string
		keywords   string
		category   string
		sortOrder  int
		isReaction bool
	}

	// Full Unicode set từ seeddata (10 reaction legacy đứng đầu + ~3.7k emoji).
	// Upsert theo code để chạy lại không trùng và giữ ID cũ (bảo vệ FK reaction).
	// Batch 500 dòng/lần — 3.7k insert lẻ + select lẻ khiến seed treo hàng chục phút.
	emojis := make([]emoji, 0, len(seeddata.Emojis))
	for _, e := range seeddata.Emojis {
		emojis = append(emojis, emoji{
			id:         internal.UUID(),
			code:       e.Code,
			imageURI:   "",
			character:  e.Character,
			name:       e.Name,
			keywords:   e.Keywords,
			category:   e.Category,
			sortOrder:  e.SortOrder,
			isReaction: e.IsReaction,
		})
	}

	const batchSize = 500
	for start := 0; start < len(emojis); start += batchSize {
		end := start + batchSize
		if end > len(emojis) {
			end = len(emojis)
		}
		batch := emojis[start:end]
		placeholders := make([]string, 0, len(batch))
		args := make([]any, 0, len(batch)*9)
		for _, e := range batch {
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args, e.id, e.code, e.imageURI, e.character, e.name, e.keywords, e.category, e.sortOrder, boolToTiny(e.isReaction))
		}
		if err := internal.Exec(database,
			`INSERT INTO emojis (id, code, image_uri, `+"`character`"+`, `+"`name`"+`, keywords, category, sort_order, is_reaction) VALUES `+
				strings.Join(placeholders, ", ")+
				` ON DUPLICATE KEY UPDATE `+"`character`"+` = VALUES(`+"`character`"+`), `+"`name`"+` = VALUES(`+"`name`"+`), keywords = VALUES(keywords),
			 category = VALUES(category), sort_order = VALUES(sort_order), is_reaction = VALUES(is_reaction)`,
			args...,
		); err != nil {
			return fmt.Errorf("core: insert emoji batch %d-%d: %w", start, end, err)
		}
	}

	// Lấy lại ID thật (upsert có thể giữ ID cũ) để các step sau tham chiếu đúng.
	rows, err := database.Query(`SELECT id, code FROM emojis`)
	if err != nil {
		return fmt.Errorf("core: select emojis: %w", err)
	}
	defer rows.Close()
	idByCode := make(map[string]string, len(emojis))
	for rows.Next() {
		var id, code string
		if err := rows.Scan(&id, &code); err != nil {
			return fmt.Errorf("core: scan emoji: %w", err)
		}
		idByCode[code] = id
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("core: rows emoji: %w", err)
	}
	for _, e := range emojis {
		realID, ok := idByCode[e.code]
		if !ok {
			return fmt.Errorf("core: missing emoji %s after upsert", e.code)
		}
		state.EmojiIDs = append(state.EmojiIDs, realID)
		if e.isReaction {
			state.ReactionEmojiIDs = append(state.ReactionEmojiIDs, realID)
		}
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
