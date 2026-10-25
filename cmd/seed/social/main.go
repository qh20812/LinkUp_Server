package social

import (
	"database/sql"
	"fmt"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
	"linkup/seeddata"
	"linkup/utils"
)

// Run chèn toàn bộ nội dung xã hội từ seeddata: bài viết (kể cả repost), media,
// bình luận, hashtag/mention, cảm xúc, lượt chia sẻ và quan hệ người dùng.
//
// Bài viết được chèn hai lượt: bài gốc trước, repost sau — vì repost tham chiếu
// shared_from_post_id tới bài gốc.
func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("social: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	// Bước relationships đã chèn cộng đồng; ở đây chỉ cần đảm bảo có ID cho
	// mọi cộng đồng được bài viết tham chiếu.
	for _, c := range seeddata.Communities {
		if _, ok := state.CommunityIDByKey[c.Key]; !ok {
			state.CommunityIDByKey[c.Key] = internal.UUID()
		}
	}

	userID := func(key string) (string, error) {
		id, ok := state.UserIDByUsername[key]
		if !ok {
			return "", fmt.Errorf("social: unknown user %q", key)
		}
		return id, nil
	}

	postCreatedAt := make(map[string]time.Time, len(seeddata.Posts))

	insertPost := func(p seeddata.Post) error {
		authorID, err := userID(p.Author)
		if err != nil {
			return err
		}

		createdAt := now.Add(-time.Duration(p.AgeHours * float64(time.Hour)))

		status := p.Status
		if status == "" {
			status = "public"
		}

		var communityID any
		if p.Community != "" {
			id, ok := state.CommunityIDByKey[p.Community]
			if !ok {
				return fmt.Errorf("social: post %s references unknown community %q", p.Key, p.Community)
			}
			communityID = id
		}

		var pinnedAt any
		if p.Pinned {
			pinnedAt = createdAt
		}

		postID := internal.UUID()

		title := p.Title
		content := p.Content
		var shareContent any
		var sharedFrom any

		if p.ShareOf != "" {
			originalID, ok := state.PostIDByKey[p.ShareOf]
			if !ok {
				return fmt.Errorf("social: repost %s references unknown post %q", p.Key, p.ShareOf)
			}
			// Repost giống service thật: không tiêu đề, không nội dung gốc,
			// chỉ có share_content + shared_from_post_id.
			title = ""
			content = ""
			sharedFrom = originalID
			if p.ShareText != "" {
				shareContent = p.ShareText
			}
		}

		if err := internal.Exec(database,
			`INSERT INTO posts (id, user_id, community_id, title, content, share_content, shared_from_post_id,
				views_count, status, is_pinned, pinned_at, comments_enabled, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, NULL)`,
			postID, authorID, communityID, title, content, shareContent, sharedFrom,
			p.Views, status, boolToInt(p.Pinned), pinnedAt, createdAt,
		); err != nil {
			return fmt.Errorf("social: insert post %s: %w", p.Key, err)
		}

		state.PostIDs = append(state.PostIDs, postID)
		state.PostIDByKey[p.Key] = postID
		postCreatedAt[p.Key] = createdAt

		// Media của bài — luôn ở trạng thái approved để feed hiển thị.
		for _, m := range p.Media {
			mediaID := internal.UUID()
			if err := internal.Exec(database,
				`INSERT INTO media (id, user_id, post_id, file_uri, file_type, file_size, duration_seconds, status, created_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, 'approved', ?)`,
				mediaID, authorID, postID, m.URI, m.Type, m.Size, m.Duration, createdAt,
			); err != nil {
				return fmt.Errorf("social: insert media for %s: %w", p.Key, err)
			}
			state.MediaIDs = append(state.MediaIDs, mediaID)
		}

		// Ảnh/video có thể là media của post gốc; repost không có media riêng.

		// Cảm xúc: rải đều vài phút sau khi đăng, không vượt quá hiện tại.
		reactionSeen := make(map[string]bool, len(p.Reactions))
		for i, r := range p.Reactions {
			reactorID, err := userID(r.User)
			if err != nil {
				return err
			}
			emojiID, ok := state.EmojiIDByCode[":"+r.Emoji+":"]
			if !ok {
				return fmt.Errorf("social: unknown reaction emoji %q on post %s", r.Emoji, p.Key)
			}
			dupKey := reactorID + "|" + emojiID
			if reactionSeen[dupKey] {
				continue
			}
			reactionSeen[dupKey] = true

			reactedAt := createdAt.Add(time.Duration(i+1) * time.Minute)
			if reactedAt.After(now) {
				reactedAt = now.Add(-time.Minute)
			}

			if err := internal.Exec(database,
				`INSERT INTO post_reactions (id, user_id, post_id, emoji_id, created_at) VALUES (?, ?, ?, ?, ?)`,
				internal.UUID(), reactorID, postID, emojiID, reactedAt,
			); err != nil {
				return fmt.Errorf("social: insert reaction on %s: %w", p.Key, err)
			}
		}

		// Hashtag + mention của bài (và của lời chia sẻ nếu là repost).
		tagSource := content
		if t, ok := shareContent.(string); ok {
			tagSource += " " + t
		}
		if err := insertTags(database, postID, nil, tagSource, state); err != nil {
			return fmt.Errorf("social: tags for post %s: %w", p.Key, err)
		}

		// Bình luận: cha trước, trả lời sau.
		var insertComment func(c seeddata.Comment, parentID *string, parentAt time.Time) error
		insertComment = func(c seeddata.Comment, parentID *string, parentAt time.Time) error {
			authorID, err := userID(c.Author)
			if err != nil {
				return err
			}

			createdAt := now.Add(-time.Duration(c.AgeMin) * time.Minute)
			minAt := postCreatedAt[p.Key].Add(time.Minute)
			if parentID != nil {
				minAt = parentAt.Add(time.Minute)
			}
			if createdAt.Before(minAt) {
				createdAt = minAt
			}
			if createdAt.After(now) {
				createdAt = now.Add(-time.Minute)
			}

			commentID := internal.UUID()
			if err := internal.Exec(database,
				`INSERT INTO comments (id, user_id, post_id, parent_id, content, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
				commentID, authorID, postID, parentID, c.Text, createdAt,
			); err != nil {
				return fmt.Errorf("insert comment: %w", err)
			}
			state.CommentIDs = append(state.CommentIDs, commentID)

			if err := insertTags(database, postID, internal.Ptr(commentID), c.Text, state); err != nil {
				return fmt.Errorf("tags for comment: %w", err)
			}

			for _, reply := range c.Replies {
				if err := insertComment(reply, internal.Ptr(commentID), createdAt); err != nil {
					return err
				}
			}
			return nil
		}

		for _, c := range p.Comments {
			if err := insertComment(c, nil, time.Time{}); err != nil {
				return fmt.Errorf("social: comment on %s: %w", p.Key, err)
			}
		}

		// Lượt chia sẻ: repost luôn có một dòng post_shares để đếm.
		if p.ShareOf != "" {
			originalID := state.PostIDByKey[p.ShareOf]
			if err := internal.Exec(database,
				`INSERT INTO post_shares (id, post_id, user_id, content, created_at) VALUES (?, ?, ?, ?, ?)`,
				internal.UUID(), originalID, authorID, p.ShareText, createdAt,
			); err != nil {
				return fmt.Errorf("social: insert share row for %s: %w", p.Key, err)
			}
		}

		return nil
	}

	// Lượt 1: bài gốc.
	for _, p := range seeddata.Posts {
		if p.ShareOf != "" {
			continue
		}
		if err := insertPost(p); err != nil {
			return err
		}
	}

	// Lượt 2: repost (đã có bài gốc để tham chiếu).
	for _, p := range seeddata.Posts {
		if p.ShareOf == "" {
			continue
		}
		if err := insertPost(p); err != nil {
			return err
		}
	}

	// Lượt chia sẻ bổ sung (không tạo bài repost).
	for _, s := range seeddata.Shares {
		sharerID, err := userID(s.User)
		if err != nil {
			return err
		}
		postID, ok := state.PostIDByKey[s.Post]
		if !ok {
			return fmt.Errorf("social: share references unknown post %q", s.Post)
		}
		createdAt := now.Add(-time.Duration(s.AgeMin) * time.Minute)
		if err := internal.Exec(database,
			`INSERT INTO post_shares (id, post_id, user_id, content, created_at) VALUES (?, ?, ?, ?, ?)`,
			internal.UUID(), postID, sharerID, s.Text, createdAt,
		); err != nil {
			return fmt.Errorf("social: insert extra share %s→%s: %w", s.User, s.Post, err)
		}
	}

	// Theo dõi.
	for i, f := range seeddata.Follows {
		followerID, err := userID(f.Follower)
		if err != nil {
			return err
		}
		followingID, err := userID(f.Following)
		if err != nil {
			return err
		}
		createdAt := now.Add(-time.Duration(10+i*3) * 24 * time.Hour)
		if err := internal.Exec(database,
			`INSERT INTO follows (id, follower_id, following_id, created_at) VALUES (?, ?, ?, ?)`,
			internal.UUID(), followerID, followingID, createdAt,
		); err != nil {
			return fmt.Errorf("social: insert follow %s→%s: %w", f.Follower, f.Following, err)
		}
	}

	// Bạn bè (mỗi cặp một dòng; accepted là đã kết bạn).
	for _, f := range seeddata.Friends {
		senderID, err := userID(f.Sender)
		if err != nil {
			return err
		}
		receiverID, err := userID(f.Receiver)
		if err != nil {
			return err
		}
		createdAt := now.Add(-time.Duration(f.AgeDays) * 24 * time.Hour)
		if err := internal.Exec(database,
			`INSERT INTO friends (id, sender_id, receiver_id, status, created_at) VALUES (?, ?, ?, ?, ?)`,
			internal.UUID(), senderID, receiverID, f.Status, createdAt,
		); err != nil {
			return fmt.Errorf("social: insert friend %s→%s: %w", f.Sender, f.Receiver, err)
		}
	}

	// Chặn.
	for _, b := range seeddata.Blocks {
		blockerID, err := userID(b.User)
		if err != nil {
			return err
		}
		blockedID, err := userID(b.Blocked)
		if err != nil {
			return err
		}
		createdAt := now.Add(-time.Duration(b.AgeDays) * 24 * time.Hour)
		if err := internal.Exec(database,
			`INSERT INTO blocks (id, user_id, blocked_user_id, created_at) VALUES (?, ?, ?, ?)`,
			internal.UUID(), blockerID, blockedID, createdAt,
		); err != nil {
			return fmt.Errorf("social: insert block %s→%s: %w", b.User, b.Blocked, err)
		}
	}

	// Bài đã lưu.
	for _, b := range seeddata.Bookmarks {
		bookmarkerID, err := userID(b.User)
		if err != nil {
			return err
		}
		postID, ok := state.PostIDByKey[b.Post]
		if !ok {
			return fmt.Errorf("social: bookmark references unknown post %q", b.Post)
		}
		createdAt := now.Add(-time.Duration(b.AgeDays) * 24 * time.Hour)
		if err := internal.Exec(database,
			`INSERT INTO bookmarks (id, user_id, post_id, created_at) VALUES (?, ?, ?, ?)`,
			internal.UUID(), bookmarkerID, postID, createdAt,
		); err != nil {
			return fmt.Errorf("social: insert bookmark %s→%s: %w", b.User, b.Post, err)
		}
	}

	return nil
}

// insertTags chèn hashtag và mention của một nội dung. commentID nil nghĩa là
// tag gắn ở cấp bài viết.
func insertTags(database *sql.DB, postID string, commentID *string, content string, state *internal.SeedState) error {
	for _, name := range utils.ExtractHashtags(content) {
		if err := internal.Exec(database,
			`INSERT INTO tags (id, post_id, comment_id, tag_type, target_user_id, name) VALUES (?, ?, ?, 'hashtag', NULL, ?)`,
			internal.UUID(), postID, commentID, name,
		); err != nil {
			return fmt.Errorf("insert hashtag %q: %w", name, err)
		}
	}

	for _, mention := range utils.ExtractMentions(content) {
		targetID, ok := state.UserIDByUsername[mention]
		if !ok {
			continue
		}
		if err := internal.Exec(database,
			`INSERT INTO tags (id, post_id, comment_id, tag_type, target_user_id, name) VALUES (?, ?, ?, 'mention', ?, ?)`,
			internal.UUID(), postID, commentID, targetID, "@"+mention,
		); err != nil {
			return fmt.Errorf("insert mention %q: %w", mention, err)
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
