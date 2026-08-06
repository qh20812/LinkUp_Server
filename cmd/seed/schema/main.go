package schema

import (
	"database/sql"
	"fmt"
	"regexp"

	"linkup/cmd/seed/internal"
	"linkup/config"
)

// validIdentifier validates that a name contains only alphanumeric characters
// and underscores — safe for use in DDL statements without parameterization.
// This prevents SQL injection via table/column names in the idempotent helpers.
var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// validateIdentifier checks that a DDL identifier (table or column name) is
// safe to embed in a SQL string. DDL does not support parameterized queries,
// so we enforce a whitelist of allowed characters instead.
func validateIdentifier(kind, name string) error {
	if !validIdentifier.MatchString(name) {
		return fmt.Errorf("invalid %s name %q: must contain only letters, digits, and underscores, and start with a letter or underscore", kind, name)
	}
	return nil
}

// addColumnIfMissing adds a column only if it does not exist,
// so re-running seed does not error on duplicate column.
func addColumnIfMissing(database *sql.DB, table, column, definition string) error {
	if err := validateIdentifier("table", table); err != nil {
		return err
	}
	if err := validateIdentifier("column", column); err != nil {
		return err
	}
	var count int
	if err := database.QueryRow(
		"SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		table, column,
	).Scan(&count); err != nil {
		return fmt.Errorf("check column %s.%s: %w", table, column, err)
	}
	if count == 0 {
		return internal.Exec(database, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition))
	}
	return nil
}

// addIndexIfMissing adds an index only if it does not exist,
// so re-running seed does not error on duplicate index.
func addIndexIfMissing(database *sql.DB, table, indexName, indexDef string) error {
	if err := validateIdentifier("table", table); err != nil {
		return err
	}
	if err := validateIdentifier("index", indexName); err != nil {
		return err
	}
	var count int
	if err := database.QueryRow(
		"SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?",
		table, indexName,
	).Scan(&count); err != nil {
		return fmt.Errorf("check index %s.%s: %w", table, indexName, err)
	}
	if count == 0 {
		return internal.Exec(database, fmt.Sprintf("ALTER TABLE %s ADD %s", table, indexDef))
	}
	return nil
}

// addForeignKeyIfMissing adds a foreign key constraint only if it does not exist,
// so re-running seed does not error on duplicate constraint.
func addForeignKeyIfMissing(database *sql.DB, table, constraintName, constraintDef string) error {
	if err := validateIdentifier("table", table); err != nil {
		return err
	}
	if err := validateIdentifier("constraint", constraintName); err != nil {
		return err
	}
	var count int
	if err := database.QueryRow(
		"SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND CONSTRAINT_NAME = ? AND CONSTRAINT_TYPE = 'FOREIGN KEY'",
		table, constraintName,
	).Scan(&count); err != nil {
		return fmt.Errorf("check fk %s.%s: %w", table, constraintName, err)
	}
	if count == 0 {
		return internal.Exec(database, fmt.Sprintf("ALTER TABLE %s ADD %s", table, constraintDef))
	}
	return nil
}

func Run(env config.Env) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("schema: connect: %w", err)
	}
	defer database.Close()

	statements := []string{
		// 1. No FK dependencies
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			username VARCHAR(50) NOT NULL,
			email VARCHAR(255) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			storage_quota_bytes DOUBLE NOT NULL DEFAULT 2147483648,
			storage_used_bytes DOUBLE NOT NULL DEFAULT 0,
			token_version INT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NULL,
			UNIQUE INDEX idx_users_username (username),
			UNIQUE INDEX idx_users_email (email)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS roles (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(50) NOT NULL UNIQUE,
			description VARCHAR(255) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS emojis (
			id VARCHAR(36) PRIMARY KEY,
			code VARCHAR(50) NOT NULL UNIQUE,
			image_uri VARCHAR(512) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS violation_rules (
			id VARCHAR(36) PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 2. Depends on users
		`CREATE TABLE IF NOT EXISTS profiles (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL UNIQUE,
			display_name VARCHAR(55) NOT NULL DEFAULT '',
			phone_number VARCHAR(20) NOT NULL DEFAULT '',
			date_of_birth DATE NULL,
			avatar_uri VARCHAR(512) NOT NULL DEFAULT '',
			bio TEXT,
			is_private_profile TINYINT(1) NOT NULL DEFAULT 0,
			is_private_posts TINYINT(1) NOT NULL DEFAULT 0,
			allow_stranger_friend_request TINYINT(1) NOT NULL DEFAULT 1,
			updated_at DATETIME NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 3. Depends on users
		`CREATE TABLE IF NOT EXISTS posts (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			community_id VARCHAR(36) NULL,
			title VARCHAR(255) NOT NULL,
			content LONGTEXT,
			views_count INT NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'public',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_posts_user_id (user_id),
			INDEX idx_posts_status (status),
			INDEX idx_posts_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 4. Depends on users
		`CREATE TABLE IF NOT EXISTS communities (
			id VARCHAR(36) PRIMARY KEY,
			creator_id VARCHAR(36) NOT NULL,
			name VARCHAR(255) NOT NULL UNIQUE,
			description TEXT,
			avatar_uri VARCHAR(512) NOT NULL DEFAULT '',
			background_uri VARCHAR(512) NOT NULL DEFAULT '',
			auto_approve TINYINT(1) NOT NULL DEFAULT 0,
			privacy VARCHAR(20) NOT NULL DEFAULT 'public',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NULL,
			FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_communities_creator (creator_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 4b. FK từ posts đến communities — moved to idempotent helpers below
		// (raw ALTER TABLE is not idempotent; re-running would fail on duplicate index/constraint)

		// 5. Depends on communities
		`CREATE TABLE IF NOT EXISTS community_rules (
			id VARCHAR(36) PRIMARY KEY,
			community_id VARCHAR(36) NOT NULL,
			category VARCHAR(50) NOT NULL,
			title VARCHAR(255) NOT NULL,
			content TEXT,
			position INT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NULL,
			FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
			INDEX idx_community_rules_community (community_id),
			INDEX idx_community_rules_category (community_id, category, position)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 6. No FK
		`CREATE TABLE IF NOT EXISTS chats (
			id VARCHAR(36) PRIMARY KEY,
			` + "`type`" + ` VARCHAR(20) NOT NULL DEFAULT 'direct',
			name VARCHAR(255) NOT NULL DEFAULT '',
			avatar_uri VARCHAR(512) NOT NULL DEFAULT '',
			encryption_key VARCHAR(255) NOT NULL DEFAULT '',
			community_id VARCHAR(36) NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
			UNIQUE INDEX idx_chats_community (community_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 6. Depends on users, posts
		`CREATE TABLE IF NOT EXISTS media (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			post_id VARCHAR(36) NULL,
			file_uri VARCHAR(512) NOT NULL,
			file_type VARCHAR(50) NOT NULL DEFAULT '',
			file_size DOUBLE NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			review_reason TEXT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			INDEX idx_media_user_id (user_id),
			INDEX idx_media_post_id (post_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 7. Depends on media
		`CREATE TABLE IF NOT EXISTS ads (
            id VARCHAR(36) PRIMARY KEY,
            partner_id VARCHAR(36) NOT NULL,
            title VARCHAR(255) NOT NULL,
            content TEXT,
            media_id VARCHAR(36) NULL,
            target_url VARCHAR(512) NOT NULL DEFAULT '',
		status VARCHAR(20) NOT NULL DEFAULT 'public',
            budget DOUBLE NOT NULL DEFAULT 0,
            started_at DATETIME NULL,
            expires_at DATETIME NULL,
            created_at DATETIME NOT NULL,
            FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE SET NULL,
            FOREIGN KEY (partner_id) REFERENCES users(id) ON DELETE CASCADE,
            INDEX idx_ads_partner_id (partner_id),
            INDEX idx_ads_media_id (media_id),
            INDEX idx_ads_status (status)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 8. Depends on users, posts
		`CREATE TABLE IF NOT EXISTS comments (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			post_id VARCHAR(36) NOT NULL,
			parent_id VARCHAR(36) NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE,
			INDEX idx_comments_post_id (post_id),
			INDEX idx_comments_user_id (user_id),
			INDEX idx_comments_parent_id (parent_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 9. Depends on posts, comments
		`CREATE TABLE IF NOT EXISTS tags (
			id VARCHAR(36) PRIMARY KEY,
			post_id VARCHAR(36) NOT NULL,
			comment_id VARCHAR(36) NULL,
			tag_type VARCHAR(20) NOT NULL DEFAULT 'hashtag',
			target_user_id VARCHAR(36) NULL,
			name VARCHAR(255) NOT NULL,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
			INDEX idx_tags_post_id (post_id),
			INDEX idx_tags_name (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 10. Depends on users, posts, emojis
		`CREATE TABLE IF NOT EXISTS post_reactions (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			post_id VARCHAR(36) NOT NULL,
			emoji_id VARCHAR(36) NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (emoji_id) REFERENCES emojis(id) ON DELETE CASCADE,
			UNIQUE INDEX idx_reactions_user_post (user_id, post_id, emoji_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 11. Depends on users
		`CREATE TABLE IF NOT EXISTS follows (
			id VARCHAR(36) PRIMARY KEY,
			follower_id VARCHAR(36) NOT NULL,
			following_id VARCHAR(36) NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE INDEX idx_follows_pair (follower_id, following_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 12. Depends on users
		`CREATE TABLE IF NOT EXISTS friends (
			id VARCHAR(36) PRIMARY KEY,
			sender_id VARCHAR(36) NOT NULL,
			receiver_id VARCHAR(36) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL,
			FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE INDEX idx_friends_pair (sender_id, receiver_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 13. Depends on users
		`CREATE TABLE IF NOT EXISTS blocks (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			blocked_user_id VARCHAR(36) NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (blocked_user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE INDEX idx_blocks_pair (user_id, blocked_user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 14. Depends on users, posts
		`CREATE TABLE IF NOT EXISTS bookmarks (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			post_id VARCHAR(36) NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			UNIQUE INDEX idx_bookmarks_user_post (user_id, post_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 15. Depends on communities, users
		`CREATE TABLE IF NOT EXISTS group_members (
			id VARCHAR(36) PRIMARY KEY,
			community_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			points INT NOT NULL DEFAULT 0,
			joined_at DATETIME NOT NULL,
			FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE INDEX idx_group_members_pair (community_id, user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 15b. Depends on communities, users
		`CREATE TABLE IF NOT EXISTS community_join_requests (
			id VARCHAR(36) PRIMARY KEY,
			community_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL,
			responded_at DATETIME NULL,
			FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE INDEX idx_community_join_requests_pair (community_id, user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 16. Depends on chats, users
		`CREATE TABLE IF NOT EXISTS chat_participants (
			id VARCHAR(36) PRIMARY KEY,
			chat_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'CHAT_MEMBER',
			joined_at DATETIME NOT NULL,
			FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE INDEX idx_chat_participants_pair (chat_id, user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 17. Depends on chats, users, media, emojis
		`CREATE TABLE IF NOT EXISTS messages (
			id VARCHAR(36) PRIMARY KEY,
			chat_id VARCHAR(36) NOT NULL,
			sender_id VARCHAR(36) NOT NULL,
			content TEXT NOT NULL,
			media_id VARCHAR(36) NULL,
			emoji_id VARCHAR(36) NULL,
			deleted_for_sender TINYINT(1) NOT NULL DEFAULT 0,
			deleted_for_receiver TINYINT(1) NOT NULL DEFAULT 0,
			deleted_at DATETIME NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
			FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE SET NULL,
			FOREIGN KEY (emoji_id) REFERENCES emojis(id) ON DELETE SET NULL,
			INDEX idx_messages_chat_id (chat_id),
			INDEX idx_messages_sender_id (sender_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 18. Depends on users
		`CREATE TABLE IF NOT EXISTS stories (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			media_uri VARCHAR(512) NOT NULL,
			media_type VARCHAR(20) NOT NULL DEFAULT 'image',
			caption TEXT,
			created_at DATETIME NOT NULL,
			expires_at DATETIME NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_stories_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 19. Depends on users
		`CREATE TABLE IF NOT EXISTS bans (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			admin_id VARCHAR(36) NOT NULL,
			reason TEXT,
			expires_at DATETIME NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (admin_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_bans_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 20. Depends on users
		`CREATE TABLE IF NOT EXISTS reports (
			id VARCHAR(36) PRIMARY KEY,
			reporter_id VARCHAR(36) NOT NULL,
			report_type VARCHAR(50) NOT NULL DEFAULT '',
			target_user_id VARCHAR(36) NULL,
			target_post_id VARCHAR(36) NULL,
			target_comment_id VARCHAR(36) NULL,
			violation_rule_id VARCHAR(36) NULL,
			reason_detail TEXT,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NULL,
			FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_reports_reporter (reporter_id),
			INDEX idx_reports_status (status),
			INDEX idx_reports_reporter_status (reporter_id, status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 21. Depends on users
		`CREATE TABLE IF NOT EXISTS moderation_logs (
			id VARCHAR(36) PRIMARY KEY,
			moderator_id VARCHAR(36) NOT NULL,
			action VARCHAR(50) NOT NULL,
			target_type VARCHAR(50) NOT NULL,
			target_id VARCHAR(36) NOT NULL,
			reason TEXT,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (moderator_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_moderation_logs_moderator (moderator_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 22. Depends on users
		`CREATE TABLE IF NOT EXISTS notifications (
			id VARCHAR(36) PRIMARY KEY,
			receiver_id VARCHAR(36) NOT NULL,
			sender_id VARCHAR(36) NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'like',
			redirect_post_id VARCHAR(36) NULL,
			redirect_user_id VARCHAR(36) NULL,
			redirect_comment_id VARCHAR(36) NULL,
			content TEXT,
			is_read TINYINT(1) NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_notifications_receiver (receiver_id),
			INDEX idx_notifications_is_read (is_read)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 23. Depends on chats, users
		`CREATE TABLE IF NOT EXISTS calls (
			id VARCHAR(36) PRIMARY KEY,
			caller_id VARCHAR(36) NOT NULL,
			callee_id VARCHAR(36) NOT NULL,
			call_type VARCHAR(20) NOT NULL DEFAULT 'voice',
			is_group TINYINT(1) NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'calling',
			started_at DATETIME NULL,
			ended_at DATETIME NULL,
			duration INT NOT NULL DEFAULT 0,
			muted_caller TINYINT(1) NOT NULL DEFAULT 0,
			muted_callee TINYINT(1) NOT NULL DEFAULT 0,
			video_enabled_caller TINYINT(1) NOT NULL DEFAULT 0,
			video_enabled_callee TINYINT(1) NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (caller_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (callee_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_calls_caller (caller_id),
			INDEX idx_calls_callee (callee_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 24. Depends on ads, users
		`CREATE TABLE IF NOT EXISTS ad_analytics (
			id VARCHAR(36) PRIMARY KEY,
			ad_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NULL,
			action_type VARCHAR(50) NOT NULL,
			ip_address VARCHAR(45) NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL,
			FOREIGN KEY (ad_id) REFERENCES ads(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
			INDEX idx_ad_analytics_ad_id (ad_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 25. Depends on users, roles
		`CREATE TABLE IF NOT EXISTS user_roles (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			role_id VARCHAR(36) NOT NULL,
			scope_id VARCHAR(36) NULL,
			scope_type VARCHAR(20) NULL,
			assigned_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
			UNIQUE INDEX idx_user_roles_pair (user_id, role_id, scope_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 26. Depends on users
		`CREATE TABLE IF NOT EXISTS notification_preferences (
			user_id VARCHAR(36) PRIMARY KEY,
			like_enabled TINYINT(1) NOT NULL DEFAULT 1,
			comment_enabled TINYINT(1) NOT NULL DEFAULT 1,
			follow_enabled TINYINT(1) NOT NULL DEFAULT 1,
			message_enabled TINYINT(1) NOT NULL DEFAULT 1,
			friend_request_enabled TINYINT(1) NOT NULL DEFAULT 1,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 27. Depends on users
		`CREATE TABLE IF NOT EXISTS password_histories (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_password_histories_user_id (user_id),
			INDEX idx_password_histories_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 28. Depends on users
		`CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			token VARCHAR(255) NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			used_at DATETIME NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_password_reset_tokens_token (token),
			INDEX idx_password_reset_tokens_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 29. Depends on users
		`CREATE TABLE IF NOT EXISTS email_verification_tokens (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			token VARCHAR(255) NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			used_at DATETIME NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_email_verification_tokens_token (token),
			INDEX idx_email_verification_tokens_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 30. Depends on users, posts
		`CREATE TABLE IF NOT EXISTS post_shares (
			id VARCHAR(36) PRIMARY KEY,
			post_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 31. Chat_invite
		`CREATE TABLE IF NOT EXISTS chat_invitations (
			id VARCHAR(36) PRIMARY KEY,
			requester_id VARCHAR(36) NOT NULL,
			target_id VARCHAR(36) NOT NULL,
			chat_id VARCHAR(36) NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (requester_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (target_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 32. Message Security — moved to idempotent addColumnIfMissing below

		// 33. group_chat_settings
		`CREATE TABLE IF NOT EXISTS group_chat_settings (
			chat_id VARCHAR(64) PRIMARY KEY,
			allow_member_add BOOLEAN NOT NULL DEFAULT TRUE,
			last_admin_transfer_at DATETIME NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 34. group_chat_member_settings
		`CREATE TABLE IF NOT EXISTS group_chat_member_settings (
			chat_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			notifications_enabled TINYINT(1) NOT NULL DEFAULT 1,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (chat_id, user_id),
			FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 35. group_chat_mutes
		`CREATE TABLE IF NOT EXISTS group_chat_mutes (
			id VARCHAR(36) PRIMARY KEY,
			chat_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			muted_by VARCHAR(36) NOT NULL,
			reason TEXT,
			expires_at DATETIME NULL,
			created_at DATETIME NOT NULL,
			INDEX idx_group_chat_mutes_chat_user (chat_id, user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 36. Depends on communities - Community contribution policy
		`CREATE TABLE IF NOT EXISTS community_policies (
			id VARCHAR(36) PRIMARY KEY,
			community_id VARCHAR(36) NOT NULL UNIQUE,
			post_weight INT NOT NULL DEFAULT 10,
			comment_weight INT NOT NULL DEFAULT 5,
			reaction_weight INT NOT NULL DEFAULT 2,
			event_weight INT NOT NULL DEFAULT 20,
			top_contributor_threshold INT NOT NULL DEFAULT 2500,
			moderator_promotion_threshold INT NOT NULL DEFAULT 5000,
			auto_promote_enabled TINYINT(1) NOT NULL DEFAULT 1,
			badge_enabled TINYINT(1) NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NULL,
			FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
			INDEX idx_community_policies_community (community_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 37. Depends on communities, users - Member contribution tracking
		`CREATE TABLE IF NOT EXISTS member_contributions (
			id VARCHAR(36) PRIMARY KEY,
			community_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			valid_posts INT NOT NULL DEFAULT 0,
			quality_comments INT NOT NULL DEFAULT 0,
			positive_reactions INT NOT NULL DEFAULT 0,
			event_participations INT NOT NULL DEFAULT 0,
			contribution_score INT NOT NULL DEFAULT 0,
			badge_type VARCHAR(50) NULL,
			promoted_to_mod TINYINT(1) NOT NULL DEFAULT 0,
			last_calculated_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NULL,
			FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE INDEX idx_member_contributions_pair (community_id, user_id),
			INDEX idx_member_contributions_score (community_id, contribution_score DESC)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 38. Depends on communities, users - Community challenges
		// 37b. Depends on communities, users — Invite codes
		`CREATE TABLE IF NOT EXISTS community_invite_codes (
			id VARCHAR(36) PRIMARY KEY,
			community_id VARCHAR(36) NOT NULL,
			code VARCHAR(6) NOT NULL,
			created_by VARCHAR(36) NOT NULL,
			max_uses INT NOT NULL DEFAULT 0,
			used_count INT NOT NULL DEFAULT 0,
			expires_at DATETIME NULL,
			is_active TINYINT(1) NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
			FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE INDEX idx_invite_codes_code (code),
			INDEX idx_invite_codes_community (community_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 37c. Depends on communities, users — Direct invitations
		`CREATE TABLE IF NOT EXISTS community_invitations (
			id VARCHAR(36) PRIMARY KEY,
			community_id VARCHAR(36) NOT NULL,
			inviter_id VARCHAR(36) NOT NULL,
			invitee_id VARCHAR(36) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL,
			responded_at DATETIME NULL,
			FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
			FOREIGN KEY (inviter_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (invitee_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_invitations_community (community_id),
			INDEX idx_invitations_invitee (invitee_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 38. Depends on community_policies, users — Community challenges
		`CREATE TABLE IF NOT EXISTS community_challenges (
			id VARCHAR(36) PRIMARY KEY,
			community_id VARCHAR(36) NOT NULL,
			creator_id VARCHAR(36) NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			hashtag VARCHAR(100) NOT NULL,
			points_per_post INT NOT NULL DEFAULT 15,
			start_date DATETIME NOT NULL,
			end_date DATETIME NOT NULL,
			max_participants INT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at DATETIME NOT NULL,
			FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
			FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_community_challenges_community (community_id),
			INDEX idx_community_challenges_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 39. Depends on community_challenges, users - Challenge participants
		`CREATE TABLE IF NOT EXISTS challenge_participants (
			id VARCHAR(36) PRIMARY KEY,
			challenge_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			posts_count INT NOT NULL DEFAULT 0,
			total_points_earned INT NOT NULL DEFAULT 0,
			joined_at DATETIME NOT NULL,
			FOREIGN KEY (challenge_id) REFERENCES community_challenges(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE INDEX idx_challenge_participants_pair (challenge_id, user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 40. Soft-delete for call history (user-level hide, row stays for the other party)
		`CREATE TABLE IF NOT EXISTS call_hidden (
			call_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (call_id, user_id),
			FOREIGN KEY (call_id) REFERENCES calls(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		// 41. group_chat_member_request:
		`CREATE TABLE IF NOT EXISTS group_chat_member_requests (
			id VARCHAR(36) PRIMARY KEY,
			chat_id VARCHAR(36) NOT NULL,
			requester_id VARCHAR(36) NOT NULL,
			target_user_id VARCHAR(36) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL,
			responded_at DATETIME NULL,
			INDEX idx_group_chat_member_requests_chat_id (chat_id),
			INDEX idx_group_chat_member_requests_requester_id (requester_id),
			INDEX idx_group_chat_member_requests_target_user_id (target_user_id),
			INDEX idx_group_chat_member_requests_status (status),
			FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
			FOREIGN KEY (requester_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 42. Add 2 columns to messages
		`ALTER TABLE messages
		ADD COLUMN is_anonymized TINYINT(1) NOT NULL DEFAULT 0,
		ADD COLUMN anonymous_name VARCHAR(255) NULL`,

		// 43. Reply Messges
		`ALTER TABLE messages ADD COLUMN reply_to_message_id VARCHAR(36) NULL`,

		// 44. Depends on chats, users — group chat bans
		`CREATE TABLE IF NOT EXISTS group_chat_bans (
			id VARCHAR(36) PRIMARY KEY,
			chat_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			banned_by VARCHAR(36) NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (banned_by) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_group_chat_bans_chat_user (chat_id, user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 45. system_configs — cài đặt hệ thống (không dependency)
		`CREATE TABLE IF NOT EXISTS system_configs (
			` + "`key`" + ` VARCHAR(100) PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 46. user_settings — cài đặt quyền riêng tư của user (1-1 với users)
		`CREATE TABLE IF NOT EXISTS user_settings (
			user_id VARCHAR(36) PRIMARY KEY,
			discoverable_in_search TINYINT(1) NOT NULL DEFAULT 1,
			allow_stranger_messages TINYINT(1) NOT NULL DEFAULT 0,
			theme VARCHAR(10) NOT NULL DEFAULT 'light',
			language VARCHAR(5) NOT NULL DEFAULT 'vi',
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 47. user_sessions — phiên đăng nhập (id = JWT jti)
		`CREATE TABLE IF NOT EXISTS user_sessions (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			device_name VARCHAR(255) NOT NULL DEFAULT '',
			ip_address VARCHAR(45) NOT NULL DEFAULT '',
			user_agent VARCHAR(512) NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL,
			expires_at DATETIME NOT NULL,
			last_active_at DATETIME NOT NULL,
			revoked_at DATETIME NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_sessions_user_id (user_id),
			INDEX idx_user_sessions_expires_at (expires_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for _, stmt := range statements {
		if err := internal.Exec(database, stmt); err != nil {
			return fmt.Errorf("schema: create table: %w", err)
		}
	}

	// Add last_read_missed_at to profiles (idempotent — skips if already exists)
	if err := addColumnIfMissing(database, "profiles", "last_read_missed_at", "DATETIME NULL"); err != nil {
		return fmt.Errorf("schema: add last_read_missed_at: %w", err)
	}

	// Phase 4 fix: Idempotent ALTER TABLE statements (previously raw ALTER TABLE
	// that failed on re-run). Now using helpers that check existence first.

	// 4b. Index + FK from posts to communities
	if err := addIndexIfMissing(database, "posts", "idx_posts_community_id",
		"INDEX idx_posts_community_id (community_id)"); err != nil {
		return fmt.Errorf("schema: add idx_posts_community_id: %w", err)
	}
	if err := addForeignKeyIfMissing(database, "posts", "fk_posts_community",
		"CONSTRAINT fk_posts_community FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE SET NULL"); err != nil {
		return fmt.Errorf("schema: add fk_posts_community: %w", err)
	}

	// Bookmark cursor metadata columns on posts (read-only in the Post model via gorm:"->").
	// Populated by the saved-posts listing (FetchSaved JOINs bookmarks); harmless NULLs elsewhere.
	if err := addColumnIfMissing(database, "posts", "bookmark_id", "VARCHAR(36) NULL"); err != nil {
		return fmt.Errorf("schema: add posts.bookmark_id: %w", err)
	}
	if err := addColumnIfMissing(database, "posts", "saved_at", "DATETIME NULL"); err != nil {
		return fmt.Errorf("schema: add posts.saved_at: %w", err)
	}

	// 31. Message encryption column
	if err := addColumnIfMissing(database, "messages", "is_encrypted", "BOOLEAN DEFAULT true"); err != nil {
		return fmt.Errorf("schema: add is_encrypted: %w", err)
	}

	// Phase 1: Admin Manage Groups/Communities — idempotent column additions

	// chats.creator_id: who created the group (NULL for direct chats)
	if err := addColumnIfMissing(database, "chats", "creator_id", "VARCHAR(36) NULL"); err != nil {
		return fmt.Errorf("schema: add chats.creator_id: %w", err)
	}
	// chats.status: moderation state for group chats
	if err := addColumnIfMissing(database, "chats", "status", "VARCHAR(20) NOT NULL DEFAULT 'active'"); err != nil {
		return fmt.Errorf("schema: add chats.status: %w", err)
	}
	// communities.status: moderation state for communities
	if err := addColumnIfMissing(database, "communities", "status", "VARCHAR(20) NOT NULL DEFAULT 'active'"); err != nil {
		return fmt.Errorf("schema: add communities.status: %w", err)
	}

	// FK from chats.creator_id to users.id
	if err := addForeignKeyIfMissing(database, "chats", "fk_chats_creator",
		"CONSTRAINT fk_chats_creator FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE SET NULL"); err != nil {
		return fmt.Errorf("schema: add fk_chats_creator: %w", err)
	}
	// index for chats.creator_id
	if err := addIndexIfMissing(database, "chats", "idx_chats_creator_id",
		"INDEX idx_chats_creator_id (creator_id)"); err != nil {
		return fmt.Errorf("schema: add idx_chats_creator_id: %w", err)
	}
	// index for communities.status (faster admin filtering)
	if err := addIndexIfMissing(database, "communities", "idx_communities_status",
		"INDEX idx_communities_status (status)"); err != nil {
		return fmt.Errorf("schema: add idx_communities_status: %w", err)
	}

	// Step D: created_at indexes for analytics queries (admin dashboard chart + UNION ALL counts)
	tables := []struct {
		Name      string
		IndexName string
		Def       string
	}{
		{"users", "idx_users_created_at", "INDEX idx_users_created_at (created_at)"},
		{"reports", "idx_reports_created_at", "INDEX idx_reports_created_at (created_at)"},
		{"comments", "idx_comments_created_at", "INDEX idx_comments_created_at (created_at)"},
		{"media", "idx_media_created_at", "INDEX idx_media_created_at (created_at)"},
		{"chats", "idx_chats_created_at", "INDEX idx_chats_created_at (created_at)"},
		{"communities", "idx_communities_created_at", "INDEX idx_communities_created_at (created_at)"},
	}
	for _, t := range tables {
		if err := addIndexIfMissing(database, t.Name, t.IndexName, t.Def); err != nil {
			return fmt.Errorf("schema: add %s: %w", t.IndexName, err)
		}
	}

	// Phase 1: composite index for notifications query (WHERE receiver_id = ? ORDER BY created_at DESC)
	if err := addIndexIfMissing(database, "notifications", "idx_notifications_receiver_created",
		"INDEX idx_notifications_receiver_created (receiver_id, created_at)"); err != nil {
		return fmt.Errorf("schema: add idx_notifications_receiver_created: %w", err)
	}

	// Phase 2: new preference columns for community and voice call notifications
	if err := addColumnIfMissing(database, "notification_preferences", "community_enabled", "TINYINT(1) NOT NULL DEFAULT 1"); err != nil {
		return fmt.Errorf("schema: add notification_preferences.community_enabled: %w", err)
	}
	if err := addColumnIfMissing(database, "notification_preferences", "voice_call_enabled", "TINYINT(1) NOT NULL DEFAULT 1"); err != nil {
		return fmt.Errorf("schema: add notification_preferences.voice_call_enabled: %w", err)
	}

	// Phase 3: indexes for admin ListPosts performance (correlated subqueries in ListPosts)
	if err := addIndexIfMissing(database, "post_reactions", "idx_post_reactions_post_id",
		"INDEX idx_post_reactions_post_id (post_id)"); err != nil {
		return fmt.Errorf("schema: add idx_post_reactions_post_id: %w", err)
	}
	if err := addIndexIfMissing(database, "post_shares", "idx_post_shares_post_id",
		"INDEX idx_post_shares_post_id (post_id)"); err != nil {
		return fmt.Errorf("schema: add idx_post_shares_post_id: %w", err)
	}

	// Phase 4: comment moderation columns (status + review_reason for report handling)
	if err := addColumnIfMissing(database, "comments", "status", "VARCHAR(20) NOT NULL DEFAULT 'active'"); err != nil {
		return fmt.Errorf("schema: add comments.status: %w", err)
	}
	if err := addColumnIfMissing(database, "comments", "review_reason", "TEXT NULL"); err != nil {
		return fmt.Errorf("schema: add comments.review_reason: %w", err)
	}
	if err := addColumnIfMissing(database, "comments", "updated_at", "DATETIME NULL"); err != nil {
		return fmt.Errorf("schema: add comments.updated_at: %w", err)
	}
	if err := addIndexIfMissing(database, "comments", "idx_comments_status",
		"INDEX idx_comments_status (status)"); err != nil {
		return fmt.Errorf("schema: add idx_comments_status: %w", err)
	}

	// Login attempt tracking columns (max_login_attempts lockout)
	if err := addColumnIfMissing(database, "users", "login_attempts", "INT NOT NULL DEFAULT 0"); err != nil {
		return fmt.Errorf("schema: add users.login_attempts: %w", err)
	}
	if err := addColumnIfMissing(database, "users", "locked_until", "DATETIME NULL"); err != nil {
		return fmt.Errorf("schema: add users.locked_until: %w", err)
	}

	// Email verification column
	if err := addColumnIfMissing(database, "users", "email_verified_at", "DATETIME NULL"); err != nil {
		return fmt.Errorf("schema: add users.email_verified_at: %w", err)
	}

	// Self deactivation column (user deactivates own account; reactivates on login)
	if err := addColumnIfMissing(database, "users", "self_deactivated_at", "DATETIME NULL"); err != nil {
		return fmt.Errorf("schema: add users.self_deactivated_at: %w", err)
	}

	return nil
}
