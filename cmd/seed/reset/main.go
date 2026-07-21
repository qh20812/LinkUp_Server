package reset

import (
	"fmt"

	"linkup/cmd/seed/internal"
	"linkup/config"
)

func Run(env config.Env) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("reset: connect: %w", err)
	}
	defer database.Close()

	tables := []string{
		"ad_media", "ad_analytics", "moderation_logs", "bans", "reports",
		"notifications", "notification_preferences", "calls", "messages", "chat_participants", "chat_invitations",
		"chats", "group_chat_bans", "group_chat_member_requests", "group_chat_member_settings",
		"group_chat_mutes", "group_chat_settings",
		"community_join_requests", "community_invitations", "community_invite_codes", "group_members", "community_rules", "communities", "tags",
		"community_policies", "member_contributions", "community_challenges", "challenge_participants",
		"post_reactions", "bookmarks", "blocks", "friends",
		"follows", "comments", "posts", "media", "stories",
		"ads", "user_roles", "profiles", "violation_rules",
		"emojis", "roles", "users",
		"call_hidden", "password_histories", "password_reset_tokens", "post_shares",
		"partner_subscriptions", "ad_packages",
		"story_interacts", "story_views",
	}

	if err := internal.Exec(database, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("reset: disable fk: %w", err)
	}

	for _, t := range tables {
		if err := internal.Exec(database, fmt.Sprintf("DROP TABLE IF EXISTS `%s`", t)); err != nil {
			return fmt.Errorf("reset: drop %s: %w", t, err)
		}
	}

	if err := internal.Exec(database, "SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		return fmt.Errorf("reset: enable fk: %w", err)
	}

	return nil
}
