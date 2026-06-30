package schema

import (
	"fmt"

	"linkup/cmd/seed/internal"
	"linkup/config"
)

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
			title VARCHAR(255) NOT NULL,
			content LONGTEXT,
			views_count INT NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
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
			created_at DATETIME NOT NULL,
			updated_at DATETIME NULL,
			FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_communities_creator (creator_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

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
			created_at DATETIME NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 6. Depends on users, posts
		`CREATE TABLE IF NOT EXISTS media (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			post_id VARCHAR(36) NULL,
			file_uri VARCHAR(512) NOT NULL,
			file_type VARCHAR(50) NOT NULL DEFAULT '',
			file_size DOUBLE NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'approved',
			created_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			INDEX idx_media_user_id (user_id),
			INDEX idx_media_post_id (post_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 7. Depends on media
		`CREATE TABLE IF NOT EXISTS ads (
			id VARCHAR(36) PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			content TEXT,
			media_id VARCHAR(36) NULL,
			target_url VARCHAR(512) NOT NULL DEFAULT '',
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			budget DOUBLE NOT NULL DEFAULT 0,
			started_at DATETIME NULL,
			expires_at DATETIME NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE SET NULL,
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
			FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_reports_reporter (reporter_id),
			INDEX idx_reports_status (status)
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
			chat_id VARCHAR(36) NULL,
			caller_id VARCHAR(36) NOT NULL,
			call_type VARCHAR(20) NOT NULL DEFAULT 'voice',
			is_group TINYINT(1) NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'completed',
			started_at DATETIME NOT NULL,
			ended_at DATETIME NULL,
			FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE SET NULL,
			FOREIGN KEY (caller_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_calls_caller (caller_id)
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

		// 29. Depends on users, posts
		`CREATE TABLE IF NOT EXISTS post_shares (
			id VARCHAR(36) PRIMARY KEY,
			post_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// 30. Chat_invite
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

		// 31. Message Security
		`ALTER TABLE messages ADD COLUMN is_encrypted BOOLEAN DEFAULT true`,
	}

	for _, stmt := range statements {
		if err := internal.Exec(database, stmt); err != nil {
			return fmt.Errorf("schema: create table: %w", err)
		}
	}

	return nil
}
