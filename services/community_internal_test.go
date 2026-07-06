package services

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"linkup/ws"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ─── Integration infrastructure ──────────────────────────────────

func connectAndMigrateCommunity(t *testing.T) *gorm.DB {
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
		&models.Community{},
		&models.GroupMember{},
		&models.Chat{},
		&models.ChatParticipant{},
		&models.CommunityJoinRequest{},
		&models.Notification{},
		&models.NotificationPreference{},
		&models.CommunityInviteCode{},
		&models.CommunityInvitation{},
	); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM community_invitations")
		db.Exec("DELETE FROM community_invite_codes")
		db.Exec("DELETE FROM notifications")
		db.Exec("DELETE FROM notification_preferences")
		db.Exec("DELETE FROM community_join_requests")
		db.Exec("DELETE FROM chat_participants")
		db.Exec("DELETE FROM group_members")
		db.Exec("DELETE FROM user_roles")
		db.Exec("DELETE FROM chats")
		db.Exec("DELETE FROM communities")
		db.Exec("DELETE FROM profiles")
		db.Exec("DELETE FROM users")
		db.Exec("DELETE FROM roles")
	})
	return db
}

func seedCommunityRoles(t *testing.T, db *gorm.DB) map[models.RoleName]string {
	entries := []struct {
		name models.RoleName
		desc string
	}{
		{models.RoleSuperAdmin, "Full system access"},
		{models.RoleAdmin, "Administrative access"},
		{models.RolePartner, "Ads partner access"},
		{models.RoleUser, "Standard user access"},
		{models.RoleChatAdmin, "Chat administrator"},
		{models.RoleChatMember, "Chat member"},
		{models.RoleGroupAdmin, "Group administrator"},
		{models.RoleGroupMod, "Group moderator"},
		{models.RoleGroupMember, "Group member"},
		{models.RoleCommunityAdmin, "Community administrator"},
		{models.RoleCommunityMember, "Community member"},
	}
	ids := make(map[models.RoleName]string, len(entries))
	for _, e := range entries {
		id := utils.GenerateUUID()
		if err := db.Create(&models.Role{ID: id, Name: e.name, Description: e.desc}).Error; err != nil {
			t.Fatalf("create role %s: %v", e.name, err)
		}
		ids[e.name] = id
	}
	return ids
}

func seedUser(t *testing.T, db *gorm.DB) string {
	id := utils.GenerateUUID()
	now := time.Now().UTC()
	if err := db.Create(&models.User{
		ID:        id,
		Username:  "testuser_" + id[:8],
		Email:     "test_" + id[:8] + "@example.com",
		Status:    models.UserStatusActive,
		CreatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&models.Profile{
		ID:          utils.GenerateUUID(),
		UserID:      id,
		DisplayName: "Test User " + id[:8],
	}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if err := db.Create(&models.NotificationPreference{
		UserID:               id,
		LikeEnabled:          true,
		CommentEnabled:       true,
		FollowEnabled:        true,
		MessageEnabled:       true,
		FriendRequestEnabled: true,
	}).Error; err != nil {
		t.Fatalf("create notification preference: %v", err)
	}
	return id
}

type communityTestSeed struct {
	DB          *gorm.DB
	RoleIDs     map[models.RoleName]string
	CreatorID   string
	MemberID    string
	CommunityID string
	ChatID      string
	Service     *CommunityService
}

func newCommunityTestSeed(t *testing.T, autoApprove bool) communityTestSeed {
	db := connectAndMigrateCommunity(t)
	roleIDs := seedCommunityRoles(t, db)

	creatorID := seedUser(t, db)
	memberID := seedUser(t, db)

	now := time.Now().UTC()
	communityID := utils.GenerateUUID()
	if err := db.Create(&models.Community{
		ID:          communityID,
		CreatorID:   creatorID,
		Name:        "Test Community " + communityID[:8],
		AutoApprove: autoApprove,
		CreatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create community: %v", err)
	}

	chatID := utils.GenerateUUID()
	if err := db.Create(&models.Chat{
		ID:            chatID,
		Type:          models.ChatTypeGroup,
		Name:          "Test Community " + communityID[:8],
		EncryptionKey: "test-enc-key",
		CommunityID:   &communityID,
		CreatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create default chat: %v", err)
	}

	if err := db.Create(&models.ChatParticipant{
		ID:       utils.GenerateUUID(),
		ChatID:   chatID,
		UserID:   creatorID,
		Role:     models.ChatRoleAdmin,
		JoinedAt: now,
	}).Error; err != nil {
		t.Fatalf("create chat participant: %v", err)
	}

	if err := db.Create(&models.GroupMember{
		ID:          utils.GenerateUUID(),
		CommunityID: communityID,
		UserID:      creatorID,
		Points:      500,
		JoinedAt:    now,
	}).Error; err != nil {
		t.Fatalf("create group member: %v", err)
	}

	scopeType := models.ScopeTypeCommunity
	if err := db.Create(&models.UserRole{
		ID:         utils.GenerateUUID(),
		UserID:     creatorID,
		RoleID:     roleIDs[models.RoleCommunityAdmin],
		ScopeID:    &communityID,
		ScopeType:  &scopeType,
		AssignedAt: now,
	}).Error; err != nil {
		t.Fatalf("assign community admin role: %v", err)
	}
	if err := db.Create(&models.UserRole{
		ID:         utils.GenerateUUID(),
		UserID:     creatorID,
		RoleID:     roleIDs[models.RoleGroupAdmin],
		ScopeID:    &communityID,
		ScopeType:  &scopeType,
		AssignedAt: now,
	}).Error; err != nil {
		t.Fatalf("assign group admin role: %v", err)
	}

	communityRepo := repository.NewCommunityRepository(db)
	authRepo := repository.NewAuthRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	prefRepo := repository.NewNotificationPreferenceRepository(db)
	hub := ws.NewHub()
	go hub.Run()
	notifService := NewNotificationService(notifRepo, prefRepo, hub)
	validation := validations.NewCommunityValidation()
	groupRole := utils.NewGroupRoleChecker(communityRepo.GetUserRole)

	svc := &CommunityService{
		repo:         communityRepo,
		authRepo:     authRepo,
		profileRepo:  profileRepo,
		mediaService: nil,
		notifService: notifService,
		groupRole:    groupRole,
		validation:   validation,
	}

	return communityTestSeed{
		DB:          db,
		RoleIDs:     roleIDs,
		CreatorID:   creatorID,
		MemberID:    memberID,
		CommunityID: communityID,
		ChatID:      chatID,
		Service:     svc,
	}
}

// newCreateCommunityTestSeed sets up DB + roles + users + service
// but does NOT create a pre-existing community/chat (caller uses Service.CreateCommunity).
func newCreateCommunityTestSeed(t *testing.T) communityTestSeed {
	db := connectAndMigrateCommunity(t)
	roleIDs := seedCommunityRoles(t, db)
	creatorID := seedUser(t, db)
	memberID := seedUser(t, db)

	communityRepo := repository.NewCommunityRepository(db)
	authRepo := repository.NewAuthRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	prefRepo := repository.NewNotificationPreferenceRepository(db)
	hub := ws.NewHub()
	go hub.Run()
	notifService := NewNotificationService(notifRepo, prefRepo, hub)
	validation := validations.NewCommunityValidation()
	groupRole := utils.NewGroupRoleChecker(communityRepo.GetUserRole)

	svc := &CommunityService{
		repo:         communityRepo,
		authRepo:     authRepo,
		profileRepo:  profileRepo,
		mediaService: nil,
		notifService: notifService,
		groupRole:    groupRole,
		validation:   validation,
	}

	return communityTestSeed{
		DB:          db,
		RoleIDs:     roleIDs,
		CreatorID:   creatorID,
		MemberID:    memberID,
		CommunityID: "",
		ChatID:      "",
		Service:     svc,
	}
}

// ─── CreateCommunity integration tests ────────────────────────────

func TestCreateCommunity_HappyPath(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, chat, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Test Community", "Description", "", false)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}
	if community.ID == "" {
		t.Fatal("community.ID is empty")
	}
	if chat.ID == "" {
		t.Fatal("chat.ID is empty")
	}
	if chat.CommunityID == nil {
		t.Fatal("chat.CommunityID is nil, want pointer to community.ID")
	}
	if *chat.CommunityID != community.ID {
		t.Errorf("chat.CommunityID = %v, want %v", *chat.CommunityID, community.ID)
	}
	if chat.Name != "Test Community" {
		t.Errorf("chat.Name = %q, want %q", chat.Name, "Test Community")
	}
}

func TestCreateCommunity_AutoApprove(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Auto Approve", "", "", true)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}
	if !community.AutoApprove {
		t.Error("AutoApprove = false, want true")
	}
}

func TestCreateCommunity_AutoApproveDefault(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "No Auto Approve", "", "", false)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}
	if community.AutoApprove {
		t.Error("AutoApprove = true, want false")
	}
}

func TestCreateCommunity_VerifyDBState(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, chat, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Verify DB", "", "", false)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}

	// chats: exactly 1 row with the correct community_id
	var chatCount int64
	seed.DB.Model(&models.Chat{}).Where("community_id = ?", community.ID).Count(&chatCount)
	if chatCount != 1 {
		t.Errorf("chats count = %d, want 1", chatCount)
	}

	// chat_participants: creator as CHAT_ADMIN
	assertChatParticipantExists(t, seed.DB, chat.ID, seed.CreatorID, models.ChatRoleAdmin)

	// group_members: exactly 1 row
	var gmCount int64
	seed.DB.Model(&models.GroupMember{}).Where("community_id = ?", community.ID).Count(&gmCount)
	if gmCount != 1 {
		t.Errorf("group_members count = %d, want 1", gmCount)
	}

	// user_roles: exactly 2 rows (COMMUNITY_ADMIN + GROUP_ADMIN)
	var urCount int64
	seed.DB.Model(&models.UserRole{}).Where("user_id = ? AND scope_id = ?", seed.CreatorID, community.ID).Count(&urCount)
	if urCount != 2 {
		t.Errorf("user_roles count = %d, want 2", urCount)
	}
}

func TestCreateCommunity_EncryptionKey(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	_, chat, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Enc Key", "", "", false)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}
	if chat.EncryptionKey == "" {
		t.Fatal("EncryptionKey is empty")
	}
	if chat.EncryptionKey == "seed-enc-key" {
		t.Error("EncryptionKey should not be the seed placeholder value")
	}
}

func TestCreateCommunity_RollbackOnFKFail(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	fakeCreatorID := "non-existent-user-id"
	_, _, err := seed.Service.CreateCommunity(ctx, fakeCreatorID, "Rollback", "", "", false)
	if err == nil {
		t.Fatal("expected error for non-existent creator, got nil")
	}

	// Verify no community or chat was persisted
	var commCount int64
	seed.DB.Model(&models.Community{}).Count(&commCount)
	if commCount != 0 {
		t.Errorf("communities count = %d, want 0 (rollback)", commCount)
	}
	var chatCount int64
	seed.DB.Model(&models.Chat{}).Count(&chatCount)
	if chatCount != 0 {
		t.Errorf("chats count = %d, want 0 (rollback)", chatCount)
	}
}

func TestCreateCommunity_NameTaken(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	communityName := "Unique Community Name"

	_, _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, communityName, "", "", false)
	if err != nil {
		t.Fatalf("first CreateCommunity: %v", err)
	}

	_, _, err = seed.Service.CreateCommunity(ctx, seed.CreatorID, communityName, "", "", false)
	if err == nil {
		t.Fatal("expected error for duplicate name, got nil")
	}
	if !errors.Is(err, validations.ErrCommunityNameExists) {
		t.Errorf("error = %v, want ErrCommunityNameExists", err)
	}
}

// ─── RequestJoin integration tests ─────────────────────────────────

func TestRequestJoin_AutoApprove_Success(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, chat, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Auto Community", "", "", true)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}

	result, err := seed.Service.RequestJoin(ctx, seed.MemberID, community.ID, "", "")
	if err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}
	if !result.AutoApproved {
		t.Fatal("AutoApproved = false, want true")
	}
	if result.RequestID != "" {
		t.Errorf("RequestID = %q, want empty", result.RequestID)
	}

	// group_members: 2 rows (creator + member)
	var gmCount int64
	seed.DB.Model(&models.GroupMember{}).Where("community_id = ?", community.ID).Count(&gmCount)
	if gmCount != 2 {
		t.Errorf("group_members count = %d, want 2", gmCount)
	}

	// chat_participants: member as CHAT_MEMBER
	assertChatParticipantExists(t, seed.DB, chat.ID, seed.MemberID, models.ChatRoleMember)

	// user_roles: member has GROUP_MEMBER role
	var urCount int64
	seed.DB.Model(&models.UserRole{}).Where("user_id = ? AND scope_id = ?", seed.MemberID, community.ID).Count(&urCount)
	if urCount != 1 {
		t.Errorf("user_roles for member count = %d, want 1", urCount)
	}
}

func TestRequestJoin_AutoApprove_AlreadyMember(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Auto Community", "", "", true)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}

	_, err = seed.Service.RequestJoin(ctx, seed.MemberID, community.ID, "", "")
	if err != nil {
		t.Fatalf("first RequestJoin: %v", err)
	}

	_, err = seed.Service.RequestJoin(ctx, seed.MemberID, community.ID, "", "")
	if err == nil {
		t.Fatal("expected error for already member, got nil")
	}
	if !errors.Is(err, validations.ErrAlreadyMember) {
		t.Errorf("error = %v, want ErrAlreadyMember", err)
	}
}

func TestRequestJoin_AutoApprove_NoChat(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Auto NoChat", "", "", true)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}

	// Delete the default group chat to simulate missing chat
	seed.DB.Where("community_id = ?", community.ID).Delete(&models.Chat{})

	_, err = seed.Service.RequestJoin(ctx, seed.MemberID, community.ID, "", "")
	if err == nil {
		t.Fatal("expected error when default chat missing, got nil")
	}
}

func TestRequestJoin_JoinRequest_Success(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Join Community", "", "", false)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}

	result, err := seed.Service.RequestJoin(ctx, seed.MemberID, community.ID, "", "")
	if err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}
	if result.AutoApproved {
		t.Fatal("AutoApproved = true, want false")
	}
	if result.RequestID == "" {
		t.Fatal("RequestID is empty, want non-empty")
	}

	// DB: join_request pending
	var req models.CommunityJoinRequest
	if err := seed.DB.Where("id = ?", result.RequestID).First(&req).Error; err != nil {
		t.Fatalf("join request not found in DB: %v", err)
	}
	if req.Status != models.JoinRequestStatusPending {
		t.Errorf("join request status = %q, want %q", req.Status, models.JoinRequestStatusPending)
	}
	if req.CommunityID != community.ID {
		t.Errorf("join request community_id = %q, want %q", req.CommunityID, community.ID)
	}
	if req.UserID != seed.MemberID {
		t.Errorf("join request user_id = %q, want %q", req.UserID, seed.MemberID)
	}
}

func TestRequestJoin_JoinRequest_Duplicate(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Join Community", "", "", false)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}

	_, err = seed.Service.RequestJoin(ctx, seed.MemberID, community.ID, "", "")
	if err != nil {
		t.Fatalf("first RequestJoin: %v", err)
	}

	_, err = seed.Service.RequestJoin(ctx, seed.MemberID, community.ID, "", "")
	if err == nil {
		t.Fatal("expected error for duplicate request, got nil")
	}
	if !errors.Is(err, validations.ErrJoinRequestPending) {
		t.Errorf("error = %v, want ErrJoinRequestPending", err)
	}
}

func TestRequestJoin_CommunityNotFound(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.RequestJoin(ctx, seed.MemberID, "non-existent-community-id", "", "")
	if err == nil {
		t.Fatal("expected error for non-existent community, got nil")
	}
	if !errors.Is(err, validations.ErrCommunityNotFound) {
		t.Errorf("error = %v, want ErrCommunityNotFound", err)
	}
}

func TestRequestJoin_DeactivatedCode_Fails(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Deactivated Code", "", "", true)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}

	result, err := seed.Service.CreateInviteCode(ctx, seed.CreatorID, community.ID, 0, nil)
	if err != nil {
		t.Fatalf("CreateInviteCode: %v", err)
	}

	if err := seed.Service.DeactivateInviteCode(ctx, seed.CreatorID, result.ID); err != nil {
		t.Fatalf("DeactivateInviteCode: %v", err)
	}

	_, err = seed.Service.RequestJoin(ctx, seed.MemberID, community.ID, result.Code, "")
	if err == nil {
		t.Fatal("expected error when using deactivated code, got nil")
	}
	if !errors.Is(err, validations.ErrInviteCodeInactive) {
		t.Errorf("error = %v, want ErrInviteCodeInactive", err)
	}
}

// ─── ApproveJoinRequest integration tests ─────────────────────────

func TestApproveJoinRequest_Success(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, chat, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Approve Community", "", "", false)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}

	result, err := seed.Service.RequestJoin(ctx, seed.MemberID, community.ID, "", "")
	if err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}

	if err := seed.Service.ApproveJoinRequest(ctx, seed.CreatorID, result.RequestID); err != nil {
		t.Fatalf("ApproveJoinRequest: %v", err)
	}

	// DB: join_request status = approved
	var req models.CommunityJoinRequest
	if err := seed.DB.Where("id = ?", result.RequestID).First(&req).Error; err != nil {
		t.Fatalf("join request not found: %v", err)
	}
	if req.Status != models.JoinRequestStatusApproved {
		t.Errorf("join request status = %q, want %q", req.Status, models.JoinRequestStatusApproved)
	}
	if req.RespondedAt == nil {
		t.Fatal("RespondedAt is nil, want non-nil")
	}

	// group_members: 2 rows (creator + member)
	var gmCount int64
	seed.DB.Model(&models.GroupMember{}).Where("community_id = ?", community.ID).Count(&gmCount)
	if gmCount != 2 {
		t.Errorf("group_members count = %d, want 2", gmCount)
	}

	// chat_participants: member as CHAT_MEMBER
	assertChatParticipantExists(t, seed.DB, chat.ID, seed.MemberID, models.ChatRoleMember)

	// user_roles: member has GROUP_MEMBER role
	var urCount int64
	seed.DB.Model(&models.UserRole{}).Where("user_id = ? AND scope_id = ?", seed.MemberID, community.ID).Count(&urCount)
	if urCount != 1 {
		t.Errorf("user_roles for member count = %d, want 1", urCount)
	}
}

func TestApproveJoinRequest_AlreadyHandled(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Approve Community", "", "", false)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}

	result, err := seed.Service.RequestJoin(ctx, seed.MemberID, community.ID, "", "")
	if err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}

	if err := seed.Service.ApproveJoinRequest(ctx, seed.CreatorID, result.RequestID); err != nil {
		t.Fatalf("first ApproveJoinRequest: %v", err)
	}

	if err := seed.Service.ApproveJoinRequest(ctx, seed.CreatorID, result.RequestID); err == nil {
		t.Fatal("expected error for already handled, got nil")
	} else if !errors.Is(err, validations.ErrJoinRequestAlreadyHandled) {
		t.Errorf("error = %v, want ErrJoinRequestAlreadyHandled", err)
	}
}

func TestApproveJoinRequest_NotAdmin(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Approve Community", "", "", false)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}

	result, err := seed.Service.RequestJoin(ctx, seed.MemberID, community.ID, "", "")
	if err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}

	// member tries to approve their own request — they are not admin
	if err := seed.Service.ApproveJoinRequest(ctx, seed.MemberID, result.RequestID); err == nil {
		t.Fatal("expected error for non-admin, got nil")
	} else if !errors.Is(err, validations.ErrNotCommunityAdmin) {
		t.Errorf("error = %v, want ErrNotCommunityAdmin", err)
	}
}

func TestApproveJoinRequest_NotFound(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	if err := seed.Service.ApproveJoinRequest(ctx, seed.CreatorID, "non-existent-request-id"); err == nil {
		t.Fatal("expected error for non-existent request, got nil")
	} else if !errors.Is(err, validations.ErrJoinRequestNotFound) {
		t.Errorf("error = %v, want ErrJoinRequestNotFound", err)
	}
}

// ─── FindDefaultGroupChat edge-case tests ─────────────────────────

func TestFindDefaultGroupChat_Found(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	chat, err := seed.Service.repo.FindDefaultGroupChatByCommunity(ctx, seed.CommunityID)
	if err != nil {
		t.Fatalf("FindDefaultGroupChatByCommunity: %v", err)
	}
	if chat.ID != seed.ChatID {
		t.Errorf("chat.ID = %q, want %q", chat.ID, seed.ChatID)
	}
	if chat.CommunityID == nil || *chat.CommunityID != seed.CommunityID {
		t.Errorf("chat.CommunityID = %v, want %v", chat.CommunityID, seed.CommunityID)
	}
}

func TestFindDefaultGroupChat_NotFound(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	// Delete the sole chat — FindDefaultGroupChatByCommunity should now fail
	seed.DB.Where("id = ?", seed.ChatID).Delete(&models.Chat{})

	_, err := seed.Service.repo.FindDefaultGroupChatByCommunity(ctx, seed.CommunityID)
	if err == nil {
		t.Fatal("expected error when no default chat exists, got nil")
	}
}

// ─── Notification verification tests ──────────────────────────────

func assertNotificationExists(t *testing.T, db *gorm.DB, receiverID string, notifType models.NotificationType) *models.Notification {
	t.Helper()
	var n models.Notification
	if err := db.Where("receiver_id = ? AND type = ?", receiverID, notifType).First(&n).Error; err != nil {
		t.Fatalf("notification (receiver=%q, type=%q) not found: %v", receiverID, notifType, err)
	}
	return &n
}

func TestCreateCommunity_SendsNotification(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	_, _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Notif Community", "", "", false)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}

	n := assertNotificationExists(t, seed.DB, seed.CreatorID, models.NotificationTypeCommunityGroupChatAdded)
	if n.Content == "" {
		t.Error("notification content is empty")
	}
	if n.SenderID != nil {
		t.Errorf("notification SenderID = %v, want nil", *n.SenderID)
	}
}

func TestRequestJoin_AutoApprove_SendsNotification(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Notif Community", "", "", true)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}

	_, err = seed.Service.RequestJoin(ctx, seed.MemberID, community.ID, "", "")
	if err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}

	n := assertNotificationExists(t, seed.DB, seed.MemberID, models.NotificationTypeCommunityGroupChatAdded)
	if n.Content == "" {
		t.Error("notification content is empty")
	}
	if n.SenderID == nil {
		t.Fatal("notification SenderID is nil, want pointer")
	}
	if *n.SenderID != seed.CreatorID {
		t.Errorf("notification SenderID = %v, want %v", *n.SenderID, seed.CreatorID)
	}
}

func TestApproveJoinRequest_SendsNotification(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	community, _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Notif Community", "", "", false)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}

	result, err := seed.Service.RequestJoin(ctx, seed.MemberID, community.ID, "", "")
	if err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}

	if err := seed.Service.ApproveJoinRequest(ctx, seed.CreatorID, result.RequestID); err != nil {
		t.Fatalf("ApproveJoinRequest: %v", err)
	}

	n := assertNotificationExists(t, seed.DB, seed.MemberID, models.NotificationTypeCommunityGroupChatAdded)
	if n.Content == "" {
		t.Error("notification content is empty")
	}
	if n.SenderID == nil {
		t.Fatal("notification SenderID is nil, want pointer")
	}
	if *n.SenderID != seed.CreatorID {
		t.Errorf("notification SenderID = %v, want %v", *n.SenderID, seed.CreatorID)
	}
}

func assertChatParticipantExists(t *testing.T, db *gorm.DB, chatID, userID string, role models.ChatRole) {
	t.Helper()
	var p models.ChatParticipant
	if err := db.Where("chat_id = ? AND user_id = ?", chatID, userID).First(&p).Error; err != nil {
		t.Fatalf("chat_participant (%s, %s) not found: %v", chatID, userID, err)
	}
	if p.Role != role {
		t.Errorf("chat_participant role = %q, want %q", p.Role, role)
	}
}

// ─── CreateInviteCode integration tests ──────────────────────────────

func TestCreateInviteCode_HappyPath(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	result, err := seed.Service.CreateInviteCode(ctx, seed.CreatorID, seed.CommunityID, 0, nil)
	if err != nil {
		t.Fatalf("CreateInviteCode: %v", err)
	}
	if result.ID == "" {
		t.Fatal("result.ID is empty")
	}
	if result.Code == "" {
		t.Fatal("result.Code is empty")
	}
	if len(result.Code) != 6 {
		t.Errorf("result.Code length = %d, want 6", len(result.Code))
	}
	if !result.IsActive {
		t.Error("IsActive = false, want true")
	}
	if result.UsedCount != 0 {
		t.Errorf("UsedCount = %d, want 0", result.UsedCount)
	}

	var code models.CommunityInviteCode
	if err := seed.DB.Where("id = ?", result.ID).First(&code).Error; err != nil {
		t.Fatalf("invite code not found in DB: %v", err)
	}
	if code.CommunityID != seed.CommunityID {
		t.Errorf("code.CommunityID = %q, want %q", code.CommunityID, seed.CommunityID)
	}
	if code.CreatedBy != seed.CreatorID {
		t.Errorf("code.CreatedBy = %q, want %q", code.CreatedBy, seed.CreatorID)
	}
}

func TestCreateInviteCode_NotAdmin(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	_, err := seed.Service.CreateInviteCode(ctx, seed.MemberID, seed.CommunityID, 0, nil)
	if err == nil {
		t.Fatal("expected error for non-admin, got nil")
	}
	if !errors.Is(err, validations.ErrNotCommunityAdmin) {
		t.Errorf("error = %v, want ErrNotCommunityAdmin", err)
	}
}

// ─── ListInviteCodes integration tests ───────────────────────────────

func TestListInviteCodes_HasCodes(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	r1, err := seed.Service.CreateInviteCode(ctx, seed.CreatorID, seed.CommunityID, 0, nil)
	if err != nil {
		t.Fatalf("CreateInviteCode 1: %v", err)
	}
	r2, err := seed.Service.CreateInviteCode(ctx, seed.CreatorID, seed.CommunityID, 5, nil)
	if err != nil {
		t.Fatalf("CreateInviteCode 2: %v", err)
	}

	items, err := seed.Service.ListInviteCodes(ctx, seed.CreatorID, seed.CommunityID)
	if err != nil {
		t.Fatalf("ListInviteCodes: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != r1.ID && items[0].ID != r2.ID {
		t.Error("returned items do not match created codes")
	}
}

func TestListInviteCodes_Empty(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	items, err := seed.Service.ListInviteCodes(ctx, seed.CreatorID, seed.CommunityID)
	if err != nil {
		t.Fatalf("ListInviteCodes: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("got %d items, want 0", len(items))
	}
}

func TestListInviteCodes_NotAdmin(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	_, err := seed.Service.ListInviteCodes(ctx, seed.MemberID, seed.CommunityID)
	if err == nil {
		t.Fatal("expected error for non-admin, got nil")
	}
	if !errors.Is(err, validations.ErrNotCommunityAdmin) {
		t.Errorf("error = %v, want ErrNotCommunityAdmin", err)
	}
}

// ─── DeactivateInviteCode integration tests ──────────────────────────

func TestDeactivateInviteCode_HappyPath(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	result, err := seed.Service.CreateInviteCode(ctx, seed.CreatorID, seed.CommunityID, 0, nil)
	if err != nil {
		t.Fatalf("CreateInviteCode: %v", err)
	}

	if err := seed.Service.DeactivateInviteCode(ctx, seed.CreatorID, result.ID); err != nil {
		t.Fatalf("DeactivateInviteCode: %v", err)
	}

	var code models.CommunityInviteCode
	if err := seed.DB.Where("id = ?", result.ID).First(&code).Error; err != nil {
		t.Fatalf("invite code not found: %v", err)
	}
	if code.IsActive {
		t.Error("IsActive = true after deactivate, want false")
	}
}

func TestDeactivateInviteCode_NotFound(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	err := seed.Service.DeactivateInviteCode(ctx, seed.CreatorID, "non-existent-code-id")
	if err == nil {
		t.Fatal("expected error for non-existent code, got nil")
	}
	if !errors.Is(err, validations.ErrInviteCodeNotFound) {
		t.Errorf("error = %v, want ErrInviteCodeNotFound", err)
	}
}

func TestDeactivateInviteCode_NotAdmin(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	result, err := seed.Service.CreateInviteCode(ctx, seed.CreatorID, seed.CommunityID, 0, nil)
	if err != nil {
		t.Fatalf("CreateInviteCode: %v", err)
	}

	err = seed.Service.DeactivateInviteCode(ctx, seed.MemberID, result.ID)
	if err == nil {
		t.Fatal("expected error for non-admin, got nil")
	}
	if !errors.Is(err, validations.ErrNotCommunityAdmin) {
		t.Errorf("error = %v, want ErrNotCommunityAdmin", err)
	}
}

// ─── SendInvitation integration tests ────────────────────────────────

func TestSendInvitation_HappyPath(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	result, err := seed.Service.SendInvitation(ctx, seed.CreatorID, seed.CommunityID, seed.MemberID)
	if err != nil {
		t.Fatalf("SendInvitation: %v", err)
	}
	if result.ID == "" {
		t.Fatal("result.ID is empty")
	}
	if result.Status != string(models.InvitationStatusPending) {
		t.Errorf("Status = %q, want %q", result.Status, string(models.InvitationStatusPending))
	}

	var inv models.CommunityInvitation
	if err := seed.DB.Where("id = ?", result.ID).First(&inv).Error; err != nil {
		t.Fatalf("invitation not found in DB: %v", err)
	}
	if inv.InviteeID != seed.MemberID {
		t.Errorf("inv.InviteeID = %q, want %q", inv.InviteeID, seed.MemberID)
	}
	if inv.InviterID != seed.CreatorID {
		t.Errorf("inv.InviterID = %q, want %q", inv.InviterID, seed.CreatorID)
	}
}

func TestSendInvitation_NotAdmin(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	_, err := seed.Service.SendInvitation(ctx, seed.MemberID, seed.CommunityID, seed.CreatorID)
	if err == nil {
		t.Fatal("expected error for non-admin, got nil")
	}
	if !errors.Is(err, validations.ErrNotCommunityAdmin) {
		t.Errorf("error = %v, want ErrNotCommunityAdmin", err)
	}
}

func TestSendInvitation_Self(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	_, err := seed.Service.SendInvitation(ctx, seed.CreatorID, seed.CommunityID, seed.CreatorID)
	if err == nil {
		t.Fatal("expected error for self-invite, got nil")
	}
	if !errors.Is(err, validations.ErrCannotInviteSelf) {
		t.Errorf("error = %v, want ErrCannotInviteSelf", err)
	}
}

func TestSendInvitation_AlreadyMember(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	// Creator is already a member — try inviting them
	_, err := seed.Service.SendInvitation(ctx, seed.CreatorID, seed.CommunityID, seed.CreatorID)
	if err == nil {
		t.Fatal("expected error for already member (self-invite), got nil")
	}
}

func TestSendInvitation_DuplicatePending(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	_, err := seed.Service.SendInvitation(ctx, seed.CreatorID, seed.CommunityID, seed.MemberID)
	if err != nil {
		t.Fatalf("first SendInvitation: %v", err)
	}

	_, err = seed.Service.SendInvitation(ctx, seed.CreatorID, seed.CommunityID, seed.MemberID)
	if err == nil {
		t.Fatal("expected error for duplicate pending invitation, got nil")
	}
	if !errors.Is(err, validations.ErrJoinRequestPending) {
		t.Errorf("error = %v, want ErrJoinRequestPending", err)
	}
}

// ─── ListMyInvitations integration tests ─────────────────────────────

func TestListMyInvitations_HasInvites(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	_, err := seed.Service.SendInvitation(ctx, seed.CreatorID, seed.CommunityID, seed.MemberID)
	if err != nil {
		t.Fatalf("SendInvitation: %v", err)
	}

	items, err := seed.Service.ListMyInvitations(ctx, seed.MemberID)
	if err != nil {
		t.Fatalf("ListMyInvitations: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].CommunityID != seed.CommunityID {
		t.Errorf("items[0].CommunityID = %q, want %q", items[0].CommunityID, seed.CommunityID)
	}
}

func TestListMyInvitations_Empty(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	items, err := seed.Service.ListMyInvitations(ctx, seed.MemberID)
	if err != nil {
		t.Fatalf("ListMyInvitations: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("got %d items, want 0", len(items))
	}
}

// ─── RespondInvitation integration tests ─────────────────────────────

func TestRespondInvitation_Accept(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	result, err := seed.Service.SendInvitation(ctx, seed.CreatorID, seed.CommunityID, seed.MemberID)
	if err != nil {
		t.Fatalf("SendInvitation: %v", err)
	}

	if err := seed.Service.RespondInvitation(ctx, seed.MemberID, result.ID, true); err != nil {
		t.Fatalf("RespondInvitation(accept): %v", err)
	}

	// DB: status = accepted
	var inv models.CommunityInvitation
	if err := seed.DB.Where("id = ?", result.ID).First(&inv).Error; err != nil {
		t.Fatalf("invitation not found: %v", err)
	}
	if inv.Status != models.InvitationStatusAccepted {
		t.Errorf("inv.Status = %q, want %q", inv.Status, models.InvitationStatusAccepted)
	}
	if inv.RespondedAt == nil {
		t.Fatal("RespondedAt is nil, want non-nil")
	}

	// group_members: 2 rows (creator + member)
	var gmCount int64
	seed.DB.Model(&models.GroupMember{}).Where("community_id = ?", seed.CommunityID).Count(&gmCount)
	if gmCount != 2 {
		t.Errorf("group_members count = %d, want 2", gmCount)
	}

	// user_roles: member has GROUP_MEMBER role
	var urCount int64
	seed.DB.Model(&models.UserRole{}).Where("user_id = ? AND scope_id = ?", seed.MemberID, seed.CommunityID).Count(&urCount)
	if urCount != 1 {
		t.Errorf("user_roles for member count = %d, want 1", urCount)
	}
}

func TestRespondInvitation_Decline(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	result, err := seed.Service.SendInvitation(ctx, seed.CreatorID, seed.CommunityID, seed.MemberID)
	if err != nil {
		t.Fatalf("SendInvitation: %v", err)
	}

	if err := seed.Service.RespondInvitation(ctx, seed.MemberID, result.ID, false); err != nil {
		t.Fatalf("RespondInvitation(decline): %v", err)
	}

	var inv models.CommunityInvitation
	if err := seed.DB.Where("id = ?", result.ID).First(&inv).Error; err != nil {
		t.Fatalf("invitation not found: %v", err)
	}
	if inv.Status != models.InvitationStatusDeclined {
		t.Errorf("inv.Status = %q, want %q", inv.Status, models.InvitationStatusDeclined)
	}
	if inv.RespondedAt == nil {
		t.Fatal("RespondedAt is nil, want non-nil")
	}

	// member should NOT have been added
	var gmCount int64
	seed.DB.Model(&models.GroupMember{}).Where("community_id = ? AND user_id = ?", seed.CommunityID, seed.MemberID).Count(&gmCount)
	if gmCount != 0 {
		t.Errorf("group_members count for member = %d, want 0", gmCount)
	}
}

func TestRespondInvitation_NotInvitee(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	result, err := seed.Service.SendInvitation(ctx, seed.CreatorID, seed.CommunityID, seed.MemberID)
	if err != nil {
		t.Fatalf("SendInvitation: %v", err)
	}

	// Creator tries to respond to an invitation addressed to Member
	err = seed.Service.RespondInvitation(ctx, seed.CreatorID, result.ID, true)
	if err == nil {
		t.Fatal("expected error when wrong user responds, got nil")
	}
	if !errors.Is(err, validations.ErrInvitationNotFound) {
		t.Errorf("error = %v, want ErrInvitationNotFound", err)
	}
}

func TestRespondInvitation_AlreadyHandled(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	result, err := seed.Service.SendInvitation(ctx, seed.CreatorID, seed.CommunityID, seed.MemberID)
	if err != nil {
		t.Fatalf("SendInvitation: %v", err)
	}

	if err := seed.Service.RespondInvitation(ctx, seed.MemberID, result.ID, true); err != nil {
		t.Fatalf("first RespondInvitation: %v", err)
	}

	err = seed.Service.RespondInvitation(ctx, seed.MemberID, result.ID, false)
	if err == nil {
		t.Fatal("expected error for already handled invitation, got nil")
	}
	if !errors.Is(err, validations.ErrInvitationAlreadyHandled) {
		t.Errorf("error = %v, want ErrInvitationAlreadyHandled", err)
	}
}

func TestRespondInvitation_NotFound(t *testing.T) {
	seed := newCommunityTestSeed(t, false)
	ctx := context.Background()

	err := seed.Service.RespondInvitation(ctx, seed.MemberID, "non-existent-invitation-id", true)
	if err == nil {
		t.Fatal("expected error for non-existent invitation, got nil")
	}
	if !errors.Is(err, validations.ErrInvitationNotFound) {
		t.Errorf("error = %v, want ErrInvitationNotFound", err)
	}
}
