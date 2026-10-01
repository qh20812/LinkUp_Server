package services

import (
	"context"
	"os"
	"testing"
	"time"

	"linkup/models"
	"linkup/repository"
	"linkup/validations"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func connectAndMigrateInterest(t *testing.T) *gorm.DB {
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
		&models.Tag{},
		&models.UserInterest{},
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
		db.Exec("DELETE FROM posts")
	})
	return db
}

func TestInterest_AddAccumulates(t *testing.T) {
	db := connectAndMigrateInterest(t)
	ctx := context.Background()
	repo := repository.NewInterestRepository(db)

	user := "user-accumulate"
	if err := repo.AddInterestScore(ctx, user, "BongDa", 3); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := repo.AddInterestScore(ctx, user, "bongda", 2); err != nil {
		t.Fatalf("add again: %v", err)
	}

	// Tag chuẩn hóa lowercase → 1 dòng duy nhất, score cộng dồn = 5.
	var rows []models.UserInterest
	if err := db.Where("user_id = ?", user).Find(&rows).Error; err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].Tag != "bongda" || rows[0].Score != 5 {
		t.Fatalf("expected {bongda,5}, got {%s,%v}", rows[0].Tag, rows[0].Score)
	}
}

func TestInterest_TopOrdering(t *testing.T) {
	db := connectAndMigrateInterest(t)
	ctx := context.Background()
	repo := repository.NewInterestRepository(db)

	user := "user-top"
	_ = repo.AddInterestScore(ctx, user, "amnhac", 2)
	_ = repo.AddInterestScore(ctx, user, "bongda", 10)
	_ = repo.AddInterestScore(ctx, user, "dulich", 5)

	top, err := repo.GetTopInterests(ctx, user, 2, 0.1)
	if err != nil {
		t.Fatalf("top: %v", err)
	}
	if len(top) != 2 || top[0].Tag != "bongda" || top[1].Tag != "dulich" {
		t.Fatalf("unexpected order: %+v", top)
	}
}

func TestInterest_RecordedFromPostView(t *testing.T) {
	db := connectAndMigrateInterest(t)
	ctx := context.Background()

	postRepo := repository.NewPostRepository(db)
	tagSvc := NewTagService(repository.NewTagRepository(db))
	svc := NewPostService(postRepo, nil, tagSvc, &validations.PostValidation{})
	svc.SetInterestRepository(repository.NewInterestRepository(db))

	author := "author-interest"
	post := seedViewTestPost(t, db, author)
	if err := tagSvc.ProcessPostHashtags(ctx, nil, post.ID, "Trận cầu hay #BongDa quá"); err != nil {
		t.Fatalf("seed tags: %v", err)
	}

	viewer := "viewer-interest"
	if _, err := svc.TrackPostView(ctx, post.ID, viewer, models.PostViewSourceFeed); err != nil {
		t.Fatalf("TrackPostView: %v", err)
	}

	// recordInterest chạy nền → poll tối đa 5s.
	deadline := time.Now().Add(5 * time.Second)
	for {
		var row models.UserInterest
		err := db.Where("user_id = ? AND tag = ?", viewer, "bongda").First(&row).Error
		if err == nil && row.Score == interestWeightFeedView {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("interest row not recorded (last err: %v)", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
