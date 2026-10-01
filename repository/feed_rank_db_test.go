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
