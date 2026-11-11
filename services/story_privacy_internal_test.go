package services

import (
	"os"
	"testing"
	"time"

	"linkup/models"
	"linkup/repository"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func connectStoryDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		t.Skip("TEST_DSN not set; skipping DB-dependent test")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Story{},
		&models.Profile{},
		&models.Follow{},
		&models.Block{},
		&models.StoryMute{},
		&models.StoryView{},
		&models.StoryInteract{},
	); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	// Production có unique index này từ cmd/migrate. AutoMigrate của models.StoryView
	// chỉ tạo index thường nên phải thêm tay để LogView dedupe đúng.
	_ = db.Exec("ALTER TABLE story_views ADD UNIQUE INDEX uq_story_views_story_viewer (story_id, viewer_id)").Error

	t.Cleanup(func() {
		db.Exec("DELETE FROM story_interacts")
		db.Exec("DELETE FROM story_views")
		db.Exec("DELETE FROM story_mutes")
		db.Exec("DELETE FROM blocks")
		db.Exec("DELETE FROM follows")
		db.Exec("DELETE FROM stories")
		db.Exec("DELETE FROM profiles")
	})
	return db
}

func seedStory(t *testing.T, db *gorm.DB, id, ownerID string) {
	t.Helper()
	now := time.Now()
	expires := now.Add(time.Hour)
	st := models.Story{
		ID:        id,
		UserID:    ownerID,
		CreatedAt: now,
		ExpiresAt: &expires,
	}
	if err := db.Create(&st).Error; err != nil {
		t.Fatalf("seed story %s: %v", id, err)
	}
}

func seedProfile(t *testing.T, db *gorm.DB, userID string, private bool) {
	t.Helper()
	p := models.Profile{
		ID:               "profile-" + userID,
		UserID:           userID,
		DisplayName:      userID,
		IsPrivateProfile: private,
	}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("seed profile %s: %v", userID, err)
	}
}

func storyIDs(stories []models.Story) map[string]struct{} {
	set := make(map[string]struct{}, len(stories))
	for _, st := range stories {
		set[st.ID] = struct{}{}
	}
	return set
}

func TestStory_FeedPrivacy(t *testing.T) {
	db := connectStoryDB(t)

	const (
		publicOwner  = "owner-public"
		privateOwner = "owner-private"
		follower     = "viewer-follower"
		stranger     = "viewer-stranger"
		blocker      = "viewer-blocked"
		muter        = "viewer-muter"
	)

	seedProfile(t, db, publicOwner, false)
	seedProfile(t, db, privateOwner, true)
	seedStory(t, db, "story-public", publicOwner)
	seedStory(t, db, "story-private", privateOwner)

	if err := db.Create(&models.Follow{ID: "f1", FollowerID: follower, FollowingID: privateOwner}).Error; err != nil {
		t.Fatalf("seed follow: %v", err)
	}
	// publicOwner chặn blocker → blocker không thấy story public.
	if err := db.Create(&models.Block{ID: "b1", UserID: publicOwner, BlockedUserID: blocker}).Error; err != nil {
		t.Fatalf("seed block: %v", err)
	}
	// muter mute privateOwner → không thấy dù đang follow.
	if err := db.Create(&models.Follow{ID: "f2", FollowerID: muter, FollowingID: privateOwner}).Error; err != nil {
		t.Fatalf("seed follow muter: %v", err)
	}
	if err := db.Create(&models.StoryMute{ID: "m1", UserID: muter, MutedUserID: privateOwner}).Error; err != nil {
		t.Fatalf("seed mute: %v", err)
	}

	repo := repository.NewStoryRepository(db)

	tests := []struct {
		name    string
		viewer  string
		want    []string
		notWant []string
	}{
		{"stranger chỉ thấy public", stranger, []string{"story-public"}, []string{"story-private"}},
		{"follower thấy cả private", follower, []string{"story-public", "story-private"}, nil},
		{"guest chỉ thấy public", "", []string{"story-public"}, []string{"story-private"}},
		{"bị chặn không thấy story của người chặn", blocker, nil, []string{"story-public"}},
		{"mute thắng privacy", muter, []string{"story-public"}, []string{"story-private"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stories, err := repo.GetActiveStoriesForViewer(tc.viewer, false)
			if err != nil {
				t.Fatalf("GetActiveStoriesForViewer: %v", err)
			}
			got := storyIDs(stories)
			for _, id := range tc.want {
				if _, ok := got[id]; !ok {
					t.Errorf("mong đợi thấy %s, got %v", id, got)
				}
			}
			for _, id := range tc.notWant {
				if _, ok := got[id]; ok {
					t.Errorf("không mong đợi thấy %s, got %v", id, got)
				}
			}
		})
	}

	// followingOnly + guest → rỗng.
	stories, err := repo.GetActiveStoriesForViewer("", true)
	if err != nil {
		t.Fatalf("GetActiveStoriesForViewer guest following: %v", err)
	}
	if len(stories) != 0 {
		t.Errorf("guest + followingOnly phải rỗng, got %d", len(stories))
	}
}

func TestStory_ServicePrivacy(t *testing.T) {
	db := connectStoryDB(t)

	const (
		privateOwner = "svc-owner-private"
		follower     = "svc-follower"
		stranger     = "svc-stranger"
	)

	seedProfile(t, db, privateOwner, true)
	seedStory(t, db, "svc-story-private", privateOwner)
	if err := db.Create(&models.Follow{ID: "sf1", FollowerID: follower, FollowingID: privateOwner}).Error; err != nil {
		t.Fatalf("seed follow: %v", err)
	}

	svc := NewStoryService(
		repository.NewStoryRepository(db),
		repository.NewProfileRepository(db),
		nil, nil,
		repository.NewFollowRepository(db),
		repository.NewBlockRepository(db),
	)

	// GetUserActiveStories
	cases := []struct {
		name   string
		viewer string
		want   int
	}{
		{"self", privateOwner, 1},
		{"follower", follower, 1},
		{"stranger", stranger, 0},
		{"guest", "", 0},
	}
	for _, tc := range cases {
		t.Run("GetUserActiveStories/"+tc.name, func(t *testing.T) {
			res, err := svc.GetUserActiveStories(privateOwner, tc.viewer)
			if err != nil {
				t.Fatalf("GetUserActiveStories: %v", err)
			}
			if len(res) != tc.want {
				t.Errorf("mong đợi %d story, got %d", tc.want, len(res))
			}
		})
	}

	// HasActiveStoryForViewer
	hasCases := []struct {
		name   string
		viewer string
		want   bool
	}{
		{"self", privateOwner, true},
		{"follower", follower, true},
		{"stranger", stranger, false},
		{"guest", "", false},
	}
	for _, tc := range hasCases {
		t.Run("HasActiveStoryForViewer/"+tc.name, func(t *testing.T) {
			has, err := svc.HasActiveStoryForViewer(privateOwner, tc.viewer)
			if err != nil {
				t.Fatalf("HasActiveStoryForViewer: %v", err)
			}
			if has != tc.want {
				t.Errorf("mong đợi %v, got %v", tc.want, has)
			}
		})
	}

	// Stranger xem trực tiếp bằng storyID → NOT_FOUND.
	if _, err := svc.ViewStory("svc-story-private", stranger); err == nil {
		t.Error("stranger không được xem story private qua ViewStory")
	}
	// Follower xem được.
	if _, err := svc.ViewStory("svc-story-private", follower); err != nil {
		t.Errorf("follower phải xem được story private: %v", err)
	}
}

func TestStory_LogViewDeduplicates(t *testing.T) {
	db := connectStoryDB(t)

	repo := repository.NewStoryRepository(db)
	if err := repo.LogView("story-dedupe", "viewer-1"); err != nil {
		t.Fatalf("LogView lần 1: %v", err)
	}
	if err := repo.LogView("story-dedupe", "viewer-1"); err != nil {
		t.Fatalf("LogView lần 2: %v", err)
	}

	count, err := repo.CountViews("story-dedupe")
	if err != nil {
		t.Fatalf("CountViews: %v", err)
	}
	if count != 1 {
		t.Errorf("mong đợi 1 view (dedupe), got %d", count)
	}
}
