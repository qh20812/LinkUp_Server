package services

import (
	"context"
	"os"
	"testing"

	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/ws"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// newNotificationTestService builds a NotificationService against TEST_DSN
// with only the tables it needs. Skips without TEST_DSN like other suites.
func newNotificationTestService(t *testing.T) (*NotificationService, *gorm.DB) {
	t.Helper()
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		t.Skip("TEST_DSN not set; skipping DB-dependent test")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	if err := db.AutoMigrate(&models.Notification{}, &models.NotificationPreference{}); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	hub := ws.NewHub()
	go hub.Run()
	svc := NewNotificationService(
		repository.NewNotificationRepository(db),
		repository.NewNotificationPreferenceRepository(db),
		nil, hub, nil, nil,
	)
	t.Cleanup(func() {
		db.Exec("DELETE FROM notifications")
		db.Exec("DELETE FROM notification_preferences")
		sqlDB.Close()
	})
	return svc, db
}

func TestNotificationCreate_SeedsDefaultPreferences(t *testing.T) {
	svc, db := newNotificationTestService(t)
	ctx := context.Background()
	userID := utils.GenerateUUID()

	got, err := svc.Create(ctx, userID, nil, models.NotificationTypeCommunityGroupChatAdded, "hello", nil, nil, nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got == nil {
		t.Fatal("expected notification, got nil (type wrongly disabled?)")
	}

	var pref models.NotificationPreference
	if err := db.Where("user_id = ?", userID).First(&pref).Error; err != nil {
		t.Fatalf("default pref row not created: %v", err)
	}
	if !pref.CommunityEnabled || !pref.LikeEnabled || !pref.MessageEnabled {
		t.Errorf("default pref not all-enabled: %+v", pref)
	}
}

func TestNotificationCreate_RespectsDisabledPreference(t *testing.T) {
	svc, db := newNotificationTestService(t)
	ctx := context.Background()
	userID := utils.GenerateUUID()

	// Existing row with community disabled must stay silent (no regression).
	pref := models.NewDefaultNotificationPreference(userID)
	pref.CommunityEnabled = false
	if err := db.Create(pref).Error; err != nil {
		t.Fatalf("seed pref: %v", err)
	}

	got, err := svc.Create(ctx, userID, nil, models.NotificationTypeCommunityGroupChatAdded, "hello", nil, nil, nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got != nil {
		t.Error("expected silent skip for disabled type, got notification")
	}
}
