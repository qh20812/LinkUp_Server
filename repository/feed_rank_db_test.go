package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"linkup/models"
	"linkup/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func connectAndMigrateFeedRank(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		t.Skip("TEST_DSN not set; skipping DB-dependent test")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Profile{},
		&models.Post{},
		&models.PostView{},
		&models.PostReaction{},
		&models.Comment{},
		&models.PostShare{},
		&models.Tag{},
		&models.UserInterest{},
		&models.Follow{},
		&models.Friend{},
		&models.Block{},
		&models.Media{},
	); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM user_interests")
		db.Exec("DELETE FROM tags")
		db.Exec("DELETE FROM post_views")
		db.Exec("DELETE FROM post_shares")
		db.Exec("DELETE FROM comments")
		db.Exec("DELETE FROM post_reactions")
		db.Exec("DELETE FROM blocks")
		db.Exec("DELETE FROM friends")
		db.Exec("DELETE FROM follows")
		db.Exec("DELETE FROM posts")
		db.Exec("DELETE FROM media")
		db.Exec("DELETE FROM profiles")
		db.Exec("DELETE FROM users")
	})
	return db
}

func seedRankPost(t *testing.T, db *gorm.DB, author, status string, createdAt time.Time) models.Post {
	post := models.Post{
		ID:        utils.GenerateUUID(),
		UserID:    author,
		Title:     "rank test",
		Content:   "content",
		Status:    models.PostStatus(status),
		CreatedAt: createdAt,
	}
	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("seed post: %v", err)
	}
	return post
}

func seedFollow(t *testing.T, db *gorm.DB, follower, following string) {
	f := models.NewFollow(follower, following)
	f.ID = utils.GenerateUUID()
	f.CreatedAt = time.Now()
	if err := db.Create(&f).Error; err != nil {
		t.Fatalf("seed follow: %v", err)
	}
}

func TestFetchPersonalized_VisibilityAndExplore(t *testing.T) {
	db := connectAndMigrateFeedRank(t)
	ctx := context.Background()
	repo := NewPostRepository(db)
	now := time.Now()

	viewer := utils.GenerateUUID()
	followed := utils.GenerateUUID()
	smallAuthor := utils.GenerateUUID()
	blockedAuthor := utils.GenerateUUID()
	friendAuthor := utils.GenerateUUID()

	seedFollow(t, db, viewer, followed)

	// Bài của người follow (cũ, có engagement nền).
	oldPost := seedRankPost(t, db, followed, "public", now.Add(-48*time.Hour))
	// Bài mới của tác giả nhỏ (0 follower).
	smallPost := seedRankPost(t, db, smallAuthor, "public", now.Add(-1*time.Hour))
	// Bài của người bị block → phải vắng mặt.
	seedRankPost(t, db, blockedAuthor, "public", now.Add(-1*time.Hour))
	blk := models.NewBlock(viewer, blockedAuthor)
	blk.ID = utils.GenerateUUID()
	blk.CreatedAt = now
	if err := db.Create(&blk).Error; err != nil {
		t.Fatalf("seed block: %v", err)
	}
	// Bài friend-only của bạn bè → phải hiện.
	fr := models.NewFriend(viewer, friendAuthor, models.FriendStatusAccepted)
	fr.ID = utils.GenerateUUID()
	fr.CreatedAt = now
	if err := db.Create(&fr).Error; err != nil {
		t.Fatalf("seed friend: %v", err)
	}
	friendPost := seedRankPost(t, db, friendAuthor, "friend", now.Add(-2*time.Hour))

	merged, next, hasMore, err := repo.FetchPersonalized(ctx, viewer, 10, PersonalCursor{}, now, false)
	if err != nil {
		t.Fatalf("FetchPersonalized: %v", err)
	}
	if len(merged) == 0 {
		t.Fatalf("expected non-empty feed")
	}

	ids := map[string]bool{}
	for _, p := range merged {
		ids[p.ID] = true
	}
	if !ids[oldPost.ID] {
		t.Errorf("expected followed author's post in feed")
	}
	if !ids[smallPost.ID] {
		t.Errorf("expected small author's new post surfaced via explore")
	}
	if !ids[friendPost.ID] {
		t.Errorf("expected friend-only post visible to friend")
	}
	for _, p := range merged {
		if p.UserID == blockedAuthor {
			t.Errorf("blocked author's post must not appear")
		}
	}

	// Phân trang tiếp bằng cursor phải chạy được (không lỗi, không trùng lặp vô hạn).
	if hasMore {
		var cursorStr string
		cursorStr = next.Encode(now)
		parsed, _, ok := ParsePersonalCursor(cursorStr)
		if !ok {
			t.Fatalf("cannot re-parse emitted cursor %q", cursorStr)
		}
		page2, _, _, err := repo.FetchPersonalized(ctx, viewer, 10, parsed, now, false)
		if err != nil {
			t.Fatalf("page 2: %v", err)
		}
		for _, p := range page2 {
			if ids[p.ID] {
				t.Errorf("page 2 repeats post %s from page 1", p.ID)
			}
		}
	}
}

// Regression: bài không có media phải trả Media=[] (serialize ra []),
// không phải nil (serialize ra null khiến web crash ở media.length).
func TestFindByIDs_NormalizesEmptyMedia(t *testing.T) {
	db := connectAndMigrateFeedRank(t)
	ctx := context.Background()
	repo := NewPostRepository(db)

	post := seedRankPost(t, db, utils.GenerateUUID(), "public", time.Now())

	posts, err := repo.FindByIDs(ctx, []string{post.ID})
	if err != nil {
		t.Fatalf("FindByIDs: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if posts[0].Media == nil {
		t.Fatalf("expected non-nil Media slice (would serialize to null)")
	}
	if len(posts[0].Media) != 0 {
		t.Fatalf("expected empty Media, got %d items", len(posts[0].Media))
	}
}

// Regression trang hashtag: FetchByIDs phải trả đủ thông tin tác giả
// (username/display_name/avatar_uri), không chỉ content + thời gian.
func TestFetchByIDs_PopulatesAuthor(t *testing.T) {
	db := connectAndMigrateFeedRank(t)
	ctx := context.Background()
	repo := NewPostRepository(db)

	authorID := utils.GenerateUUID()
	if err := db.Create(&models.User{ID: authorID, Username: "tagauthor"}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	profile := models.NewProfile(authorID, "Tag Author", "", nil, "https://cdn.example.com/a.png", "")
	profile.ID = utils.GenerateUUID()
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	post := seedRankPost(t, db, authorID, "public", time.Now())

	posts, err := repo.FetchByIDs(ctx, []string{post.ID}, 10, 0)
	if err != nil {
		t.Fatalf("FetchByIDs: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	got := posts[0]
	if got.Username != "tagauthor" {
		t.Errorf("expected username tagauthor, got %q", got.Username)
	}
	if got.DisplayName != "Tag Author" {
		t.Errorf("expected display name Tag Author, got %q", got.DisplayName)
	}
	if got.AvatarURI != "https://cdn.example.com/a.png" {
		t.Errorf("expected avatar URI, got %q", got.AvatarURI)
	}
	if got.Media == nil {
		t.Errorf("expected non-nil Media slice")
	}
}
