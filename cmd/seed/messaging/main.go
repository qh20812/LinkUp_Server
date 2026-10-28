package messaging

import (
	"fmt"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
	"linkup/seeddata"
)

// Run chèn hội thoại trực tiếp và nhóm từ seeddata.Chats: chat, thành viên,
// tin nhắn và con trỏ đọc (chat_reads) để danh sách hiện đúng số tin chưa đọc.
func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("messaging: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	chatAdminRoleID := state.RoleIDs[4]
	chatMemberRoleID := state.RoleIDs[5]

	type roleKey struct{ userID, roleID, scopeID string }
	assigned := make(map[roleKey]bool)

	addUserRole := func(userID, roleID, scopeID string) error {
		key := roleKey{userID, roleID, scopeID}
		if assigned[key] {
			return nil
		}
		assigned[key] = true
		return internal.Exec(database,
			`INSERT INTO user_roles (id, user_id, role_id, scope_id, scope_type, assigned_at) VALUES (?, ?, ?, ?, 'chat', ?)`,
			internal.UUID(), userID, roleID, scopeID, now,
		)
	}

	userID := func(key string) (string, error) {
		id, ok := state.UserIDByUsername[key]
		if !ok {
			return "", fmt.Errorf("messaging: unknown user %q", key)
		}
		return id, nil
	}

	for _, c := range seeddata.Chats {
		chatID := internal.UUID()
		state.ChatIDs = append(state.ChatIDs, chatID)
		state.ChatIDByKey[c.Key] = chatID

		createdAt := now.Add(-time.Duration(7) * 24 * time.Hour)
		if err := internal.Exec(database,
			"INSERT INTO chats (id, `type`, name, avatar_uri, encryption_key, created_at) VALUES (?, ?, ?, ?, 'seed-enc-key', ?)",
			chatID, c.Type, c.Name, c.Avatar, createdAt,
		); err != nil {
			return fmt.Errorf("messaging: insert chat %s: %w", c.Key, err)
		}

		for i, member := range c.Members {
			memberID, err := userID(member)
			if err != nil {
				return err
			}
			role := "CHAT_MEMBER"
			roleID := chatMemberRoleID
			if c.Type == "group" && i == 0 {
				role = "CHAT_ADMIN"
				roleID = chatAdminRoleID
			}

			if err := internal.Exec(database,
				`INSERT INTO chat_participants (id, chat_id, user_id, role, joined_at) VALUES (?, ?, ?, ?, ?)`,
				internal.UUID(), chatID, memberID, role, createdAt,
			); err != nil {
				return fmt.Errorf("messaging: insert participant %s in %s: %w", member, c.Key, err)
			}

			if err := addUserRole(memberID, roleID, chatID); err != nil {
				return fmt.Errorf("messaging: assign role %s in %s: %w", role, c.Key, err)
			}
		}

		if c.Type == "group" {
			if err := internal.Exec(database,
				`INSERT INTO group_chat_settings (chat_id, allow_member_add, created_at, updated_at) VALUES (?, ?, ?, ?)`,
				chatID, boolToInt(c.AllowAdd), createdAt, now,
			); err != nil {
				return fmt.Errorf("messaging: insert group settings for %s: %w", c.Key, err)
			}
		}

		// Tin nhắn: seeddata xếp theo thứ tự thời gian tăng dần.
		type inserted struct {
			id        string
			createdAt time.Time
		}
		sent := make([]inserted, 0, len(c.Messages))
		for _, m := range c.Messages {
			senderID, err := userID(m.From)
			if err != nil {
				return err
			}
			sentAt := now.Add(-time.Duration(m.AgeMin) * time.Minute)
			if sentAt.Before(createdAt) {
				sentAt = createdAt.Add(time.Minute)
			}
			messageID := internal.UUID()
			if err := internal.Exec(database,
				`INSERT INTO messages (id, chat_id, sender_id, content, media_id, emoji_id, created_at)
				 VALUES (?, ?, ?, ?, NULL, NULL, ?)`,
				messageID, chatID, senderID, m.Text, sentAt,
			); err != nil {
				return fmt.Errorf("messaging: insert message in %s: %w", c.Key, err)
			}
			state.MessageIDs = append(state.MessageIDs, messageID)
			sent = append(sent, inserted{id: messageID, createdAt: sentAt})
		}

		if len(sent) == 0 {
			continue
		}

		unread := make(map[string]bool, len(c.UnreadFor))
		for _, u := range c.UnreadFor {
			unread[u] = true
		}

		for _, member := range c.Members {
			memberID, err := userID(member)
			if err != nil {
				return err
			}
			watermark := sent[len(sent)-1]
			if unread[member] && len(sent) > 2 {
				// Để lại 2 tin cuối chưa đọc.
				watermark = sent[len(sent)-3]
			}
			if err := internal.Exec(database,
				`INSERT INTO chat_reads (chat_id, user_id, last_read_at, last_message_id, updated_at)
				 VALUES (?, ?, ?, ?, ?)`,
				chatID, memberID, watermark.createdAt, watermark.id, now,
			); err != nil {
				return fmt.Errorf("messaging: insert chat_read for %s in %s: %w", member, c.Key, err)
			}
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
