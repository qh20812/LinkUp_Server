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
		{internal.UUID(), ":like:", "https://cdn.example.com/emojis/like.png"},
		{internal.UUID(), ":love:", "https://cdn.example.com/emojis/love.png"},
		{internal.UUID(), ":haha:", "https://cdn.example.com/emojis/haha.png"},
		{internal.UUID(), ":wow:", "https://cdn.example.com/emojis/wow.png"},
		{internal.UUID(), ":sad:", "https://cdn.example.com/emojis/sad.png"},
		{internal.UUID(), ":angry:", "https://cdn.example.com/emojis/angry.png"},
		{internal.UUID(), ":clap:", "https://cdn.example.com/emojis/clap.png"},
		{internal.UUID(), ":fire:", "https://cdn.example.com/emojis/fire.png"},
		{internal.UUID(), ":heart:", "https://cdn.example.com/emojis/heart.png"},
		{internal.UUID(), ":rocket:", "https://cdn.example.com/emojis/rocket.png"},
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
		id          string
		title       string
		description string
	}

	rules := []violationRule{
		{internal.UUID(), "Spam", "Posting repetitive or unsolicited content"},
		{internal.UUID(), "Harassment", "Bullying or threatening others"},
		{internal.UUID(), "Hate Speech", "Promoting violence or discrimination"},
		{internal.UUID(), "Nudity", "Posting explicit or adult content"},
		{internal.UUID(), "Copyright", "Posting copyrighted material without permission"},
		{internal.UUID(), "Impersonation", "Pretending to be someone else"},
		{internal.UUID(), "Misinformation", "Sharing false or misleading information"},
		{internal.UUID(), "Self-harm", "Content promoting self-harm or suicide"},
	}

	for _, v := range rules {
		if err := internal.Exec(database,
			`INSERT INTO violation_rules (id, title, description) VALUES (?, ?, ?)`,
			v.id, v.title, v.description,
		); err != nil {
			return fmt.Errorf("core: insert violation_rule %s: %w", v.title, err)
		}
		state.ViolationRuleIDs = append(state.ViolationRuleIDs, v.id)
	}

	superAdminID := state.RoleIDs[0]
	adminID := state.RoleIDs[1]
	userRoleID := state.RoleIDs[2]

	assigned := map[string]bool{}

	addRole := func(userID, roleID string) error {
		key := userID + "|" + roleID
		if assigned[key] {
			return nil
		}
		assigned[key] = true
		return internal.Exec(database,
			`INSERT INTO user_roles (id, user_id, role_id, assigned_at) VALUES (?, ?, ?, ?)`,
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
