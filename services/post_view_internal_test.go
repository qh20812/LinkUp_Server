package services

import (
	"context"
	"os"
	"testing"
	"time"

	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ─── Integration infrastructure (require TEST_DSN, skip nếu thiếu) ───

func connectAndMigratePostView(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		t.Skip("TEST_DSN not set; skipping DB-dependent test")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Post{},
		&models.PostView{},
		&models.PostReaction{},
		&models.Comment{},
		&models.PostShare{},
	); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM post_views")
		db.Exec("DELETE FROM post_shares")
		db.Exec("DELETE FROM comments")
		db.Exec("DELETE FROM post_reactions")
		db.Exec("DELETE FROM posts")
	})
	return db
}

func seedViewTestPost(t *testing.T, db *gorm.DB, authorID string) models.Post {
	post := models.Post{
		ID:        utils.GenerateUUID(),
		UserID:    authorID,
		Title:     "view test",
		Content:   "content",
		Status:    models.PostStatusPublic,
		CreatedAt: time.Now(),
	}
	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("seed post: %v", err)
	}
	return post
}

func newViewTestService(db *gorm.DB) PostService {
	repo := repository.NewPostRepository(db)
	return NewPostService(repo, nil, nil, &validations.PostValidation{})
}

func TestTrackPostView_CountsFirstView(t *testing.T) {
	db := connectAndMigratePostView(t)
	svc := newViewTestService(db)
	ctx := context.Background()

	author := utils.GenerateUUID()
	viewer := utils.GenerateUUID()
	post := seedViewTestPost(t, db, author)

	counted, err := svc.TrackPostView(ctx, post.ID, viewer, models.PostViewSourceFeed)
	if err != nil {
		t.Fatalf("TrackPostView: %v", err)
	}
	if !counted {
		t.Fatalf("expected first view to be counted")
	}

	var dbPost models.Post
	if err := db.Table("posts").Where("id = ?", post.ID).First(&dbPost).Error; err != nil {
		t.Fatalf("read post: %v", err)
	}
	if dbPost.ViewsCount != 1 {
		t.Fatalf("expected views_count=1, got %d", dbPost.ViewsCount)
	}
}

func TestTrackPostView_DedupsSameUserSameDay(t *testing.T) {
	db := connectAndMigratePostView(t)
	svc := newViewTestService(db)
	ctx := context.Background()

	author := utils.GenerateUUID()
	viewer := utils.GenerateUUID()
	post := seedViewTestPost(t, db, author)

	if _, err := svc.TrackPostView(ctx, post.ID, viewer, models.PostViewSourceFeed); err != nil {
		t.Fatalf("first TrackPostView: %v", err)
	}
	counted, err := svc.TrackPostView(ctx, post.ID, viewer, models.PostViewSourceDetail)
	if err != nil {
		t.Fatalf("second TrackPostView: %v", err)
	}
	if counted {
		t.Fatalf("expected second view same day to NOT be counted")
	}

	var dbPost models.Post
	if err := db.Table("posts").Where("id = ?", post.ID).First(&dbPost).Error; err != nil {
		t.Fatalf("read post: %v", err)
	}
	if dbPost.ViewsCount != 1 {
		t.Fatalf("expected views_count=1, got %d", dbPost.ViewsCount)
	}
}

func TestTrackPostView_SkipsSelfView(t *testing.T) {
	db := connectAndMigratePostView(t)
	svc := newViewTestService(db)
	ctx := context.Background()

	author := utils.GenerateUUID()
	post := seedViewTestPost(t, db, author)

	counted, err := svc.TrackPostView(ctx, post.ID, author, models.PostViewSourceDetail)
	if err != nil {
		t.Fatalf("TrackPostView: %v", err)
	}
	if counted {
		t.Fatalf("expected self view to NOT be counted")
	}

	// Self view vẫn được log để phân tích.
	var logs int64
	if err := db.Model(&models.PostView{}).Where("post_id = ?", post.ID).Count(&logs).Error; err != nil {
		t.Fatalf("count logs: %v", err)
	}
	if logs != 1 {
		t.Fatalf("expected 1 log row, got %d", logs)
	}
}

func TestTrackPostView_PostNotFound(t *testing.T) {
	db := connectAndMigratePostView(t)
	svc := newViewTestService(db)

	if _, err := svc.TrackPostView(context.Background(), "no-such-post", utils.GenerateUUID(), models.PostViewSourceFeed); err == nil {
		t.Fatalf("expected error for missing post")
	}
}
