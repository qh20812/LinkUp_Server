package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"linkup/config"
	"linkup/db"
)

type userRow struct {
	ID    int
	Email string
}

type emojiRow struct {
	ID   int
	Code string
}

type postSeed struct {
	UserID int
	Title  string
	Body   string
	Views  int
	Status string
}

type storySeed struct {
	UserID    int
	MediaURI  string
	MediaType string
	Caption   string
	ExpiresAt string
}

type mediaSeed struct {
	UserID   int
	PostID   sql.NullInt64
	FileURI  string
	FileType string
	FileSize float64
	Status   string
}

type postReactionSeed struct {
	UserID  int
	PostID  int
	EmojiID int
}

type commentSeed struct {
	UserID   int
	PostID   int
	ParentID sql.NullInt64
	Content  string
}

type bookmarkSeed struct {
	UserID int
	PostID int
}

type tagSeed struct {
	PostID       int
	CommentID    sql.NullInt64
	TagType      string
	TargetUserID sql.NullInt64
	Name         string
}

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	conn, err := db.ConnectDb(config.GetEnv())
	if err != nil {
		log.Fatalf("DB connection: failed (%v)", err)
	}
	defer conn.Close()

	if err := ensurePhase3Tables(conn); err != nil {
		log.Fatalf("ensure phase3 tables failed: %v", err)
	}

	if err := clearPhase3Data(conn); err != nil {
		log.Fatalf("clear old phase3 data failed: %v", err)
	}

	users, err := fetchUsers(conn)
	if err != nil {
		log.Fatalf("fetch users failed: %v", err)
	}
	if len(users) == 0 {
		log.Fatalf("no users found for phase3 seeding")
	}

	emojis, err := fetchEmojis(conn)
	if err != nil {
		log.Fatalf("fetch emojis failed: %v", err)
	}
	if len(emojis) == 0 {
		log.Fatalf("no emojis found for phase3 seeding")
	}

	posts := buildPosts(users, 100)
	postsInserted, err := seedPosts(conn, posts)
	if err != nil {
		log.Fatalf("seed posts failed: %v", err)
	}

	postsIDs, err := fetchPostIDs(conn, 100)
	if err != nil {
		log.Fatalf("fetch post ids failed: %v", err)
	}

	stories := buildStories(users)
	storiesInserted, err := seedStories(conn, stories)
	if err != nil {
		log.Fatalf("seed stories failed: %v", err)
	}

	media := buildMedia(users, postsIDs)
	mediaInserted, err := seedMedia(conn, media)
	if err != nil {
		log.Fatalf("seed media failed: %v", err)
	}

	postReactions := buildPostReactions(users, postsIDs, emojis)
	reactionsInserted, err := seedPostReactions(conn, postReactions)
	if err != nil {
		log.Fatalf("seed post reactions failed: %v", err)
	}

	comments := buildComments(users, postsIDs)
	commentsInserted, err := seedComments(conn, comments)
	if err != nil {
		log.Fatalf("seed comments failed: %v", err)
	}

	bookmarks := buildBookmarks(users, postsIDs)
	bookmarksInserted, err := seedBookmarks(conn, bookmarks)
	if err != nil {
		log.Fatalf("seed bookmarks failed: %v", err)
	}

	tags := buildTags(users, postsIDs)
	tagsInserted, err := seedTags(conn, tags)
	if err != nil {
		log.Fatalf("seed tags failed: %v", err)
	}

	fmt.Printf("Seed phase3: success (posts=%d, stories=%d, media=%d, reactions=%d, comments=%d, bookmarks=%d, tags=%d)\n",
		postsInserted,
		storiesInserted,
		mediaInserted,
		reactionsInserted,
		commentsInserted,
		bookmarksInserted,
		tagsInserted,
	)
}

func ensurePhase3Tables(conn *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS posts (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			title VARCHAR(255) NOT NULL,
			content VARCHAR(3000),
			views_count INT NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_posts_user FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS stories (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			media_uri VARCHAR(500) NOT NULL,
			media_type VARCHAR(20) NOT NULL DEFAULT 'video',
			caption VARCHAR(500),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NULL,
			CONSTRAINT fk_stories_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS media (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			post_id INT NULL,
			file_uri VARCHAR(500) NOT NULL,
			file_type VARCHAR(50) NOT NULL,
			file_size FLOAT,
			status VARCHAR(20) NOT NULL DEFAULT 'approved',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_media_user FOREIGN KEY (user_id) REFERENCES users(id),
			CONSTRAINT fk_media_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS post_reactions (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			post_id INT NOT NULL,
			emoji_id INT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY user_post_unique (user_id, post_id),
			CONSTRAINT fk_post_reactions_user FOREIGN KEY (user_id) REFERENCES users(id),
			CONSTRAINT fk_post_reactions_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			CONSTRAINT fk_post_reactions_emoji FOREIGN KEY (emoji_id) REFERENCES emojis(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS comments (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			post_id INT NOT NULL,
			parent_id INT NULL,
			content VARCHAR(500) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_comments_user FOREIGN KEY (user_id) REFERENCES users(id),
			CONSTRAINT fk_comments_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			CONSTRAINT fk_comments_parent FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS bookmarks (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			post_id INT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY bookmark_user_post_unique (user_id, post_id),
			CONSTRAINT fk_bookmarks_user FOREIGN KEY (user_id) REFERENCES users(id),
			CONSTRAINT fk_bookmarks_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS tags (
			id INT AUTO_INCREMENT PRIMARY KEY,
			post_id INT NOT NULL,
			comment_id INT NULL,
			tag_type VARCHAR(20),
			target_user_id INT NULL,
			name VARCHAR(100),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_tags_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			CONSTRAINT fk_tags_comment FOREIGN KEY (comment_id) REFERENCES comments(id),
			CONSTRAINT fk_tags_target_user FOREIGN KEY (target_user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, q := range queries {
		if _, err := conn.Exec(q); err != nil {
			return fmt.Errorf("create phase3 table: %w", err)
		}
	}

	return nil
}

func clearPhase3Data(conn *sql.DB) error {
	queries := []string{
		"DELETE FROM notifications WHERE redirect_post_id IS NOT NULL OR redirect_comment_id IS NOT NULL",
		"DELETE FROM reports WHERE target_post_id IS NOT NULL OR target_comment_id IS NOT NULL",
		"DELETE FROM ad_analytics",
		"DELETE FROM ads WHERE media_id IS NOT NULL",
		"DELETE FROM tags",
		"DELETE FROM bookmarks",
		"DELETE FROM post_reactions",
		"DELETE FROM comments",
		"DELETE FROM media",
		"DELETE FROM stories",
		"DELETE FROM posts",
	}

	for _, q := range queries {
		if _, err := conn.Exec(q); err != nil {
			return fmt.Errorf("clear phase3 data: %w", err)
		}
	}

	return nil
}

func fetchUsers(conn *sql.DB) ([]userRow, error) {
	rows, err := conn.Query("SELECT id, email FROM users ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	users := make([]userRow, 0)
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.ID, &u.Email); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

func fetchEmojis(conn *sql.DB) ([]emojiRow, error) {
	rows, err := conn.Query("SELECT id, code FROM emojis ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query emojis: %w", err)
	}
	defer rows.Close()

	emojis := make([]emojiRow, 0)
	for rows.Next() {
		var e emojiRow
		if err := rows.Scan(&e.ID, &e.Code); err != nil {
			return nil, fmt.Errorf("scan emoji: %w", err)
		}
		emojis = append(emojis, e)
	}

	return emojis, rows.Err()
}

func fetchPostIDs(conn *sql.DB, limit int) ([]int, error) {
	rows, err := conn.Query("SELECT id FROM posts ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("query post ids: %w", err)
	}
	defer rows.Close()

	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan post id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func buildPosts(users []userRow, total int) []postSeed {
	posts := make([]postSeed, 0, total)
	for index := 1; index <= total; index++ {
		user := users[(index-1)%len(users)]
		status := "active"
		if index%12 == 0 {
			status = "hidden"
		}

		posts = append(posts, postSeed{
			UserID: user.ID,
			Title:  fmt.Sprintf("Seed bài viết #%03d", index),
			Body:   fmt.Sprintf("Nội dung mẫu cho bài viết số %d của user %s. Đây là dữ liệu seed để kiểm tra hiển thị và tương tác.", index, user.Email),
			Views:  index * 7,
			Status: status,
		})
	}
	return posts
}

func buildStories(users []userRow) []storySeed {
	stories := make([]storySeed, 0, len(users))
	for index, user := range users {
		mediaType := "image"
		if index%3 == 0 {
			mediaType = "video"
		}

		mediaURI := fmt.Sprintf("https://picsum.photos/seed/story-%03d/640/480", index+1)
		if mediaType == "video" {
			mediaURI = fmt.Sprintf("https://picsum.photos/seed/story-video-%03d/640/480", index+1)
		}

		stories = append(stories, storySeed{
			UserID:    user.ID,
			MediaURI:  mediaURI,
			MediaType: mediaType,
			Caption:   fmt.Sprintf("Story của %s cho ngày seed thứ %d.", user.Email, index+1),
			ExpiresAt: "", // use NULL date by leaving empty
		})
	}
	return stories
}

func buildMedia(users []userRow, postIDs []int) []mediaSeed {
	items := make([]mediaSeed, 0, len(postIDs)+len(users))
	for index, postID := range postIDs {
		owner := users[index%len(users)]
		fileType := "image"
		if index%4 == 0 {
			fileType = "video"
		}

		fileURI := fmt.Sprintf("https://picsum.photos/seed/media-%03d/800/600", index+1)
		if fileType == "video" {
			fileURI = fmt.Sprintf("https://picsum.photos/seed/media-video-%03d/800/600", index+1)
		}

		items = append(items, mediaSeed{
			UserID:   owner.ID,
			PostID:   sql.NullInt64{Int64: int64(postID), Valid: true},
			FileURI:  fileURI,
			FileType: fileType,
			FileSize: 1.2 + float64(index%7)*0.3,
			Status:   "approved",
		})
	}

	for index, user := range users {
		items = append(items, mediaSeed{
			UserID:   user.ID,
			PostID:   sql.NullInt64{Valid: false},
			FileURI:  fmt.Sprintf("https://picsum.photos/seed/avatar-%03d/320/320", index+1),
			FileType: "image",
			FileSize: 0.8 + float64(index%5)*0.2,
			Status:   "approved",
		})
	}

	return items
}

func buildPostReactions(users []userRow, postIDs []int, emojis []emojiRow) []postReactionSeed {
	items := make([]postReactionSeed, 0)
	for i, user := range users {
		for j := 0; j < 3 && j < len(postIDs); j++ {
			postID := postIDs[(i+j)%len(postIDs)]
			emoji := emojis[(i+j)%len(emojis)]
			items = append(items, postReactionSeed{UserID: user.ID, PostID: postID, EmojiID: emoji.ID})
		}
	}
	return items
}

func buildComments(users []userRow, postIDs []int) []commentSeed {
	items := make([]commentSeed, 0, len(postIDs)*2)
	for i, postID := range postIDs {
		user := users[i%len(users)]
		items = append(items, commentSeed{
			UserID:   user.ID,
			PostID:   postID,
			ParentID: sql.NullInt64{Valid: false},
			Content:  fmt.Sprintf("Bình luận thử nghiệm cho bài viết %d của user %s.", postID, user.Email),
		})
		if i%5 == 0 {
			items = append(items, commentSeed{
				UserID:   users[(i+1)%len(users)].ID,
				PostID:   postID,
				ParentID: sql.NullInt64{Valid: false},
				Content:  fmt.Sprintf("Phản hồi phụ cho bài viết %d.", postID),
			})
		}
	}
	return items
}

func buildBookmarks(users []userRow, postIDs []int) []bookmarkSeed {
	items := make([]bookmarkSeed, 0)
	for i, user := range users {
		if i >= 15 {
			break
		}
		for j := 0; j < 5 && j < len(postIDs); j++ {
			items = append(items, bookmarkSeed{UserID: user.ID, PostID: postIDs[(i+j)%len(postIDs)]})
		}
	}
	return items
}

func buildTags(users []userRow, postIDs []int) []tagSeed {
	items := make([]tagSeed, 0)
	for i, postID := range postIDs {
		if i%8 != 0 {
			continue
		}
		items = append(items, tagSeed{
			PostID:       postID,
			CommentID:    sql.NullInt64{Valid: false},
			TagType:      "hashtag",
			TargetUserID: sql.NullInt64{Valid: false},
			Name:         fmt.Sprintf("#seedpost%d", postID),
		})
		items = append(items, tagSeed{
			PostID:       postID,
			CommentID:    sql.NullInt64{Valid: false},
			TagType:      "mention",
			TargetUserID: sql.NullInt64{Int64: int64(users[i%len(users)].ID), Valid: true},
			Name:         fmt.Sprintf("@%s", strings.SplitN(users[i%len(users)].Email, "@", 2)[0]),
		})
	}
	return items
}

func seedPosts(conn *sql.DB, posts []postSeed) (int64, error) {
	values := make([][]any, 0, len(posts))
	for _, item := range posts {
		values = append(values, []any{item.UserID, item.Title, item.Body, item.Views, item.Status})
	}
	return bulkInsertIgnore(conn, "posts", []string{"user_id", "title", "content", "views_count", "status"}, values)
}

func seedStories(conn *sql.DB, stories []storySeed) (int64, error) {
	values := make([][]any, 0, len(stories))
	for _, item := range stories {
		expiresAt := sql.NullString{Valid: false}
		if item.ExpiresAt != "" {
			expiresAt = sql.NullString{String: item.ExpiresAt, Valid: true}
		}
		values = append(values, []any{item.UserID, item.MediaURI, item.MediaType, item.Caption, expiresAt})
	}
	return bulkInsertIgnore(conn, "stories", []string{"user_id", "media_uri", "media_type", "caption", "expires_at"}, values)
}

func seedMedia(conn *sql.DB, mediaItems []mediaSeed) (int64, error) {
	values := make([][]any, 0, len(mediaItems))
	for _, item := range mediaItems {
		values = append(values, []any{item.UserID, item.PostID, item.FileURI, item.FileType, item.FileSize, item.Status})
	}
	return bulkInsertIgnore(conn, "media", []string{"user_id", "post_id", "file_uri", "file_type", "file_size", "status"}, values)
}

func seedPostReactions(conn *sql.DB, reactions []postReactionSeed) (int64, error) {
	values := make([][]any, 0, len(reactions))
	for _, item := range reactions {
		values = append(values, []any{item.UserID, item.PostID, item.EmojiID})
	}
	return bulkInsertIgnore(conn, "post_reactions", []string{"user_id", "post_id", "emoji_id"}, values)
}

func seedComments(conn *sql.DB, comments []commentSeed) (int64, error) {
	values := make([][]any, 0, len(comments))
	for _, item := range comments {
		values = append(values, []any{item.UserID, item.PostID, item.ParentID, item.Content})
	}
	return bulkInsertIgnore(conn, "comments", []string{"user_id", "post_id", "parent_id", "content"}, values)
}

func seedBookmarks(conn *sql.DB, bookmarks []bookmarkSeed) (int64, error) {
	values := make([][]any, 0, len(bookmarks))
	for _, item := range bookmarks {
		values = append(values, []any{item.UserID, item.PostID})
	}
	return bulkInsertIgnore(conn, "bookmarks", []string{"user_id", "post_id"}, values)
}

func seedTags(conn *sql.DB, tags []tagSeed) (int64, error) {
	values := make([][]any, 0, len(tags))
	for _, item := range tags {
		values = append(values, []any{item.PostID, item.CommentID, item.TagType, item.TargetUserID, item.Name})
	}
	return bulkInsertIgnore(conn, "tags", []string{"post_id", "comment_id", "tag_type", "target_user_id", "name"}, values)
}

func bulkInsertIgnore(conn *sql.DB, table string, columns []string, rows [][]any) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	var builder strings.Builder
	builder.WriteString("INSERT IGNORE INTO ")
	builder.WriteString(table)
	builder.WriteString(" (")
	builder.WriteString(strings.Join(columns, ", "))
	builder.WriteString(") VALUES ")

	args := make([]any, 0, len(rows)*len(columns))
	for rowIndex, row := range rows {
		if rowIndex > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("(")
		for colIndex := range columns {
			if colIndex > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString("?")
		}
		builder.WriteString(")")
		args = append(args, row...)
	}

	result, err := conn.Exec(builder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("insert seed rows into %s: %w", table, err)
	}

	inserted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for %s: %w", table, err)
	}
	return inserted, nil
}
