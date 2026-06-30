package messaging

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
		return fmt.Errorf("messaging: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	superAdminID := state.UserIDs[0]
	chatAdminRoleID := state.RoleIDs[3]
	chatMemberRoleID := state.RoleIDs[4]

	userRoleAssigned := map[string]bool{}

	addUserRole := func(userID, roleID, scopeID string) error {
		key := userID + "|" + roleID + "|" + scopeID
		if userRoleAssigned[key] {
			return nil
		}
		userRoleAssigned[key] = true
		return internal.Exec(database,
			`INSERT INTO user_roles (id, user_id, role_id, scope_id, scope_type, assigned_at) VALUES (?, ?, ?, ?, 'chat', ?)`,
			internal.UUID(), userID, roleID, scopeID, now,
		)
	}

	type chatData struct {
		id   string
		typ  string
		name string
	}

	chats := []chatData{
		{internal.UUID(), "direct", ""},
		{internal.UUID(), "direct", ""},
		{internal.UUID(), "direct", ""},
		{internal.UUID(), "group", "Go Dev Team"},
		{internal.UUID(), "group", "Project Alpha"},
		{internal.UUID(), "group", "Coffee Chat"},
		{internal.UUID(), "direct", ""},
		{internal.UUID(), "group", "Weekend Hikers"},
	}

	for _, c := range chats {
		avatarURI := fmt.Sprintf("https://api.dicebear.com/7.x/identicon/svg?seed=chat%s", c.id[:8])
		if err := internal.Exec(database,
			"INSERT INTO chats (id, `type`, name, avatar_uri, created_at) VALUES (?, ?, ?, ?, ?)",
			c.id, c.typ, c.name, avatarURI, now,
		); err != nil {
			return fmt.Errorf("messaging: insert chat %s: %w", c.id[:8], err)
		}
		state.ChatIDs = append(state.ChatIDs, c.id)
	}

	type participantData struct {
		chatIdx int
		userIdx int
		role    string
	}

	participants := []participantData{
		{0, 2, "CHAT_ADMIN"}, {0, 3, "CHAT_MEMBER"},
		{1, 4, "CHAT_ADMIN"}, {1, 5, "CHAT_MEMBER"},
		{2, 6, "CHAT_ADMIN"}, {2, 7, "CHAT_MEMBER"},
		{3, 2, "CHAT_ADMIN"}, {3, 3, "CHAT_MEMBER"}, {3, 4, "CHAT_MEMBER"}, {3, 5, "CHAT_MEMBER"},
		{4, 6, "CHAT_ADMIN"}, {4, 7, "CHAT_MEMBER"}, {4, 8, "CHAT_MEMBER"},
		{5, 9, "CHAT_ADMIN"}, {5, 10, "CHAT_MEMBER"}, {5, 11, "CHAT_MEMBER"}, {5, 12, "CHAT_MEMBER"},
		{6, 13, "CHAT_ADMIN"}, {6, 14, "CHAT_MEMBER"},
		{7, 15, "CHAT_ADMIN"}, {7, 16, "CHAT_MEMBER"}, {7, 17, "CHAT_MEMBER"}, {7, 18, "CHAT_MEMBER"},
	}

	for _, p := range participants {
		userID := state.UserIDs[p.userIdx]
		if userID == superAdminID {
			continue
		}

		if err := internal.Exec(database,
			`INSERT INTO chat_participants (id, chat_id, user_id, role, joined_at) VALUES (?, ?, ?, ?, ?)`,
			internal.UUID(), state.ChatIDs[p.chatIdx], userID, p.role, now,
		); err != nil {
			return fmt.Errorf("messaging: insert participant chat %d user %d: %w", p.chatIdx, p.userIdx, err)
		}

		var roleID string
		if p.role == "CHAT_ADMIN" {
			roleID = chatAdminRoleID
		} else {
			roleID = chatMemberRoleID
		}
		chatID := state.ChatIDs[p.chatIdx]
		if err := addUserRole(userID, roleID, chatID); err != nil {
			return fmt.Errorf("messaging: user_role %s for %s: %w", p.role, userID, err)
		}
	}

	messages := []string{
		"Hey, how are you?",
		"Did you see the latest update?",
		"We need to discuss the project timeline.",
		"Can you review my pull request?",
		"The deployment went smoothly.",
		"Meeting at 3 PM today?",
		"I pushed some changes to the repo.",
		"Great work on the last sprint!",
		"Anyone free for a quick call?",
		"Let's schedule a code review session.",
		"The CI pipeline is failing on main.",
		"I fixed the bug you reported.",
		"Can someone help me with this issue?",
		"Documentation has been updated.",
		"We should refactor the auth module.",
		"New feature request: dark mode.",
		"Performance improved by 40%!",
		"Who wants to grab lunch?",
		"The API rate limit is too low.",
		"I'll handle the database migration.",
		"Check out this cool library I found.",
		"We need more test coverage.",
		"Deploying to staging now.",
		"The UI looks much better now.",
		"Any updates on the security audit?",
	}

	for i := 0; i < 50; i++ {
		chatIdx := i % len(state.ChatIDs)
		var senderIdx int
		for _, p := range participants {
			if p.chatIdx == chatIdx {
				senderIdx = p.userIdx
				break
			}
		}
		senderID := state.UserIDs[senderIdx%len(state.UserIDs)]
		content := pick(messages)

		if err := internal.Exec(database,
			`INSERT INTO messages (id, chat_id, sender_id, content, media_id, emoji_id, created_at) VALUES (?, ?, ?, ?, NULL, NULL, ?)`,
			internal.UUID(), state.ChatIDs[chatIdx], senderID, content, now.Add(-time.Duration(50-i)*time.Minute),
		); err != nil {
			return fmt.Errorf("messaging: insert message %d: %w", i, err)
		}
	}

	return nil
}
