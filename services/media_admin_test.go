package services

import (
	"context"
	"os"
	"testing"
	"time"

	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ─── Integration infrastructure ──────────────────────────────────

func connectAndMigrateMediaAdmin(t *testing.T) *gorm.DB {
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
		&models.Role{},
		&models.UserRole{},
		&models.Media{},
		&models.ModerationLog{},
		&models.Notification{},
		&models.NotificationPreference{},
	); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM notifications")
		db.Exec("DELETE FROM notification_preferences")
		db.Exec("DELETE FROM moderation_logs")
		db.Exec("DELETE FROM media")
		db.Exec("DELETE FROM user_roles")
		db.Exec("DELETE FROM profiles")
		db.Exec("DELETE FROM users")
		db.Exec("DELETE FROM roles")
	})
	return db
}

func seedMediaAdminRoles(t *testing.T, db *gorm.DB) (adminRoleID string, superAdminRoleID string, userRoleID string) {
	entries := []struct {
		name        models.RoleName
		description string
	}{
		{models.RoleSuperAdmin, "Super admin"},
		{models.RoleAdmin, "Admin"},
		{models.RoleUser, "Standard user"},
	}
	roleIDs := make(map[models.RoleName]string)
	for _, e := range entries {
		role := models.Role{
			ID:          utils.GenerateUUID(),
			Name:        e.name,
			Description: e.description,
		}
		if err := db.Create(&role).Error; err != nil {
			t.Fatalf("seed role %s: %v", e.name, err)
		}
		roleIDs[e.name] = role.ID
	}
	return roleIDs[models.RoleAdmin], roleIDs[models.RoleSuperAdmin], roleIDs[models.RoleUser]
}

func seedMediaAdminUser(t *testing.T, db *gorm.DB, roleID string) (userID string) {
	userID = utils.GenerateUUID()
	now := time.Now().UTC()
	user := models.User{
		ID:                userID,
		Username:          "admin",
		Email:             "admin@media-test.local",
		PasswordHash:      "hash",
		Status:            models.UserStatusActive,
		StorageQuotaBytes: 100 * 1024 * 1024,
		StorageUsedBytes:  0,
		CreatedAt:         now,
		UpdatedAt:         &now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	profile := models.Profile{
		ID:          utils.GenerateUUID(),
		UserID:      userID,
		DisplayName: "Admin User",
		AvatarURI:   "",
		Bio:         "",
		UpdatedAt:   &now,
	}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	userRole := models.UserRole{
		ID:         utils.GenerateUUID(),
		UserID:     userID,
		RoleID:     roleID,
		ScopeID:    nil,
		ScopeType:  nil,
		AssignedAt: now,
	}
	if err := db.Create(&userRole).Error; err != nil {
		t.Fatalf("seed user_role: %v", err)
	}
	return userID
}

func seedMediaOwner(t *testing.T, db *gorm.DB, userRoleID string) (userID string) {
	userID = utils.GenerateUUID()
	now := time.Now().UTC()
	user := models.User{
		ID:                userID,
		Username:          "owner",
		Email:             "owner@media-test.local",
		PasswordHash:      "hash",
		Status:            models.UserStatusActive,
		StorageQuotaBytes: 100 * 1024 * 1024,
		StorageUsedBytes:  0,
		CreatedAt:         now,
		UpdatedAt:         &now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed owner: %v", err)
	}
	profile := models.Profile{
		ID:          utils.GenerateUUID(),
		UserID:      userID,
		DisplayName: "Media Owner",
		AvatarURI:   "",
		Bio:         "",
		UpdatedAt:   &now,
	}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("seed owner profile: %v", err)
	}
	userRole := models.UserRole{
		ID:         utils.GenerateUUID(),
		UserID:     userID,
		RoleID:     userRoleID,
		ScopeID:    nil,
		ScopeType:  nil,
		AssignedAt: now,
	}
	if err := db.Create(&userRole).Error; err != nil {
		t.Fatalf("seed owner user_role: %v", err)
	}
	pref := models.NotificationPreference{
		UserID:               userID,
		LikeEnabled:          true,
		CommentEnabled:       true,
		FollowEnabled:        true,
		MessageEnabled:       true,
		FriendRequestEnabled: true,
	}
	if err := db.Create(&pref).Error; err != nil {
		t.Fatalf("seed notification pref: %v", err)
	}
	return userID
}

func seedFlaggedMedia(t *testing.T, db *gorm.DB, ownerID string, status models.MediaStatus, createdAt time.Time) string {
	media := models.NewMedia(ownerID, nil, "https://res.cloudinary.com/test/image/upload/v1/test.jpg", "image/jpeg", 1024)
	media.ID = utils.GenerateUUID()
	media.Status = status
	media.CreatedAt = createdAt
	if err := db.Create(&media).Error; err != nil {
		t.Fatalf("seed media: %v", err)
	}
	return media.ID
}

func buildAdminService(db *gorm.DB) *AdminService {
	authRepo := repository.NewAuthRepository(db)
	banRepo := repository.NewBanRepository(db)
	postRepo := repository.NewPostRepository(db)
	reportRepo := repository.NewReportRepository(db)
	moderationRepo := repository.NewModerationRepository(db)
	chatRepo := repository.NewChatRepository(db)
	communityRepo := repository.NewCommunityRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	groupChatRepo := repository.NewGroupChatRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	mediaRepo := repository.NewMediaRepository(db)
	notifPrefRepo := repository.NewNotificationPreferenceRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	notifService := NewNotificationService(notifRepo, notifPrefRepo, nil)

	return NewAdminService(
		authRepo, banRepo, postRepo, reportRepo, moderationRepo,
		chatRepo, communityRepo, profileRepo, groupChatRepo,
		adminRepo, mediaRepo, notifService,
	)
}

// ─── Tests ───────────────────────────────────────────────────────

func TestAdminService_ReviewMedia_Approve(t *testing.T) {
	db := connectAndMigrateMediaAdmin(t)
	_, _, userRoleID := seedMediaAdminRoles(t, db)
	adminID := seedMediaAdminUser(t, db, userRoleID)
	ownerID := seedMediaOwner(t, db, userRoleID)
	mediaID := seedFlaggedMedia(t, db, ownerID, models.MediaStatusFlagged, time.Now().UTC())

	svc := buildAdminService(db)
	err := svc.ReviewMedia(context.Background(), adminID, mediaID, dto.AdminReviewMediaInput{
		Action: "approve",
		Reason: "Looks good",
	})
	if err != nil {
		t.Fatalf("ReviewMedia approve: %v", err)
	}

	var got models.Media
	if err := db.Where("id = ?", mediaID).First(&got).Error; err != nil {
		t.Fatalf("get media: %v", err)
	}
	if got.Status != models.MediaStatusApproved {
		t.Errorf("expected approved, got %s", got.Status)
	}

	var logCount int64
	db.Model(&models.ModerationLog{}).Where("target_id = ?", mediaID).Count(&logCount)
	if logCount != 1 {
		t.Errorf("expected 1 moderation log, got %d", logCount)
	}
}

func TestAdminService_ReviewMedia_Reject(t *testing.T) {
	db := connectAndMigrateMediaAdmin(t)
	_, _, userRoleID := seedMediaAdminRoles(t, db)
	adminID := seedMediaAdminUser(t, db, userRoleID)
	ownerID := seedMediaOwner(t, db, userRoleID)
	mediaID := seedFlaggedMedia(t, db, ownerID, models.MediaStatusFlagged, time.Now().UTC())

	svc := buildAdminService(db)
	err := svc.ReviewMedia(context.Background(), adminID, mediaID, dto.AdminReviewMediaInput{
		Action: "reject",
		Reason: "Violates guidelines",
	})
	if err != nil {
		t.Fatalf("ReviewMedia reject: %v", err)
	}

	var got models.Media
	if err := db.Where("id = ?", mediaID).First(&got).Error; err != nil {
		t.Fatalf("get media: %v", err)
	}
	if got.Status != models.MediaStatusRejected {
		t.Errorf("expected rejected, got %s", got.Status)
	}

	var logCount int64
	db.Model(&models.ModerationLog{}).Where("target_id = ?", mediaID).Count(&logCount)
	if logCount != 1 {
		t.Errorf("expected 1 moderation log, got %d", logCount)
	}
}

func TestAdminService_ReviewMedia_InvalidAction(t *testing.T) {
	db := connectAndMigrateMediaAdmin(t)
	_, _, userRoleID := seedMediaAdminRoles(t, db)
	adminID := seedMediaAdminUser(t, db, userRoleID)
	ownerID := seedMediaOwner(t, db, userRoleID)
	mediaID := seedFlaggedMedia(t, db, ownerID, models.MediaStatusFlagged, time.Now().UTC())

	svc := buildAdminService(db)
	err := svc.ReviewMedia(context.Background(), adminID, mediaID, dto.AdminReviewMediaInput{
		Action: "ban",
		Reason: "nope",
	})
	if err == nil {
		t.Fatal("expected error for invalid action, got nil")
	}
}

func TestAdminService_ReviewMedia_InvalidTransition(t *testing.T) {
	db := connectAndMigrateMediaAdmin(t)
	_, _, userRoleID := seedMediaAdminRoles(t, db)
	adminID := seedMediaAdminUser(t, db, userRoleID)
	ownerID := seedMediaOwner(t, db, userRoleID)
	// Seed media with approved status — not reviewable
	mediaID := seedFlaggedMedia(t, db, ownerID, models.MediaStatusApproved, time.Now().UTC())

	svc := buildAdminService(db)
	err := svc.ReviewMedia(context.Background(), adminID, mediaID, dto.AdminReviewMediaInput{
		Action: "reject",
		Reason: "nope",
	})
	if err == nil {
		t.Fatal("expected error for non-flagged media, got nil")
	}
}

func TestAdminService_ListFlaggedMedia_Defaults(t *testing.T) {
	db := connectAndMigrateMediaAdmin(t)
	_, _, userRoleID := seedMediaAdminRoles(t, db)
	adminID := seedMediaAdminUser(t, db, userRoleID)
	ownerID := seedMediaOwner(t, db, userRoleID)

	for i := 0; i < 5; i++ {
		seedFlaggedMedia(t, db, ownerID, models.MediaStatusFlagged, time.Now().UTC())
	}

	svc := buildAdminService(db)
	resp, err := svc.ListFlaggedMedia(context.Background(), adminID, dto.AdminMediaFilterInput{})
	if err != nil {
		t.Fatalf("ListFlaggedMedia: %v", err)
	}
	if len(resp.Items) != 5 {
		t.Errorf("expected 5 items, got %d", len(resp.Items))
	}
	if resp.Total != 5 {
		t.Errorf("expected total 5, got %d", resp.Total)
	}
	if resp.Page != 1 {
		t.Errorf("expected page 1, got %d", resp.Page)
	}
}

func TestAdminService_ListFlaggedMedia_FilterByStatus(t *testing.T) {
	db := connectAndMigrateMediaAdmin(t)
	_, _, userRoleID := seedMediaAdminRoles(t, db)
	adminID := seedMediaAdminUser(t, db, userRoleID)
	ownerID := seedMediaOwner(t, db, userRoleID)

	seedFlaggedMedia(t, db, ownerID, models.MediaStatusFlagged, time.Now().UTC())
	seedFlaggedMedia(t, db, ownerID, models.MediaStatusApproved, time.Now().UTC())

	svc := buildAdminService(db)
	// Filter by "rejected" — no results
	resp, err := svc.ListFlaggedMedia(context.Background(), adminID, dto.AdminMediaFilterInput{
		Status: "rejected",
	})
	if err != nil {
		t.Fatalf("ListFlaggedMedia: %v", err)
	}
	if len(resp.Items) != 0 {
		t.Errorf("expected 0 items for rejected filter, got %d", len(resp.Items))
	}
}

func TestAdminService_CleanupRejectedMedia(t *testing.T) {
	db := connectAndMigrateMediaAdmin(t)
	_, _, userRoleID := seedMediaAdminRoles(t, db)
	adminID := seedMediaAdminUser(t, db, userRoleID)
	ownerID := seedMediaOwner(t, db, userRoleID)

	// Create rejected media older than 7 days and recent rejected + flagged
	old := time.Now().UTC().AddDate(0, 0, -10)
	seedFlaggedMedia(t, db, ownerID, models.MediaStatusRejected, old)
	seedFlaggedMedia(t, db, ownerID, models.MediaStatusRejected, time.Now().UTC())
	seedFlaggedMedia(t, db, ownerID, models.MediaStatusFlagged, old)

	svc := buildAdminService(db)
	cleaned, err := svc.CleanupRejectedMedia(context.Background(), adminID)
	if err != nil {
		t.Fatalf("CleanupRejectedMedia: %v", err)
	}
	// Only the old rejected media should be cleaned (1 item).
	// The recent rejected and old flagged are skipped.
	if cleaned != 1 {
		t.Errorf("expected 1 cleaned, got %d", cleaned)
	}
}

func TestAdminService_ReviewMedia_NotAdmin(t *testing.T) {
	db := connectAndMigrateMediaAdmin(t)
	_, _, userRoleID := seedMediaAdminRoles(t, db)
	// Regular user, not admin
	ownerID := seedMediaOwner(t, db, userRoleID)
	mediaID := seedFlaggedMedia(t, db, ownerID, models.MediaStatusFlagged, time.Now().UTC())

	svc := buildAdminService(db)
	err := svc.ReviewMedia(context.Background(), ownerID, mediaID, dto.AdminReviewMediaInput{
		Action: "approve",
		Reason: "ok",
	})
	if err == nil {
		t.Fatal("expected error for non-admin, got nil")
	}
}
