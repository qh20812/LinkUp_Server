package services

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ─── Pure-logic tests (no DB needed) ────────────────────────────────────────

func TestContributionScoreAndBadgeHelpers(t *testing.T) {
	svc := &ContributionService{}
	policy := &models.CommunityPolicy{
		PostWeight:                  10,
		CommentWeight:               5,
		ReactionWeight:              2,
		EventWeight:                 20,
		TopContributorThreshold:     2500,
		ModeratorPromotionThreshold: 5000,
		AutoPromoteEnabled:          true,
		BadgeEnabled:                true,
	}

	contribution := &models.MemberContribution{
		ValidPosts:          100,
		QualityComments:     50,
		PositiveReactions:   25,
		EventParticipations: 10,
	}

	score := svc.calculateContributionScore(contribution, policy)
	if score != 1000+250+50+200 {
		t.Fatalf("score = %d, want %d", score, 1500)
	}

	contribution.ContributionScore = 2500
	svc.checkAndAssignBadge(contribution, policy)
	if contribution.BadgeType == nil || *contribution.BadgeType != "Top Contributor" {
		t.Fatalf("badge = %v, want Top Contributor", contribution.BadgeType)
	}

	contribution.ContributionScore = 5000
	err := svc.checkAndPromoteToMod(context.Background(), contribution, policy)
	if err != nil {
		t.Fatalf("checkAndPromoteToMod failed: %v", err)
	}
	if !contribution.PromotedToMod {
		t.Fatal("expected promoted_to_mod to be true")
	}
}

func TestExtractHashtags(t *testing.T) {
	content := "Joining the #LinkUpPhoto challenge! also #golang and #LinkUpPhoto."
	hashtags := extractHashtags(content)
	if len(hashtags) != 2 {
		t.Fatalf("len = %d, want 2; value = %#v", len(hashtags), hashtags)
	}
	if hashtags[0] != "#LinkUpPhoto" || hashtags[1] != "#golang" {
		t.Fatalf("hashtags = %#v", hashtags)
	}
}

func TestCalculateContributionScore_Zero(t *testing.T) {
	svc := &ContributionService{}
	score := svc.calculateContributionScore(&models.MemberContribution{}, &models.CommunityPolicy{})
	if score != 0 {
		t.Fatalf("score = %d, want 0", score)
	}
}

func TestCheckAndAssignBadge_Disabled(t *testing.T) {
	svc := &ContributionService{}
	contribution := &models.MemberContribution{ContributionScore: 99999}
	policy := &models.CommunityPolicy{BadgeEnabled: false, TopContributorThreshold: 100}
	svc.checkAndAssignBadge(contribution, policy)
	if contribution.BadgeType != nil {
		t.Fatalf("badge = %v, want nil when badge disabled", contribution.BadgeType)
	}
}

func TestCheckAndPromoteToMod_Disabled(t *testing.T) {
	svc := &ContributionService{}
	contribution := &models.MemberContribution{ContributionScore: 99999}
	policy := &models.CommunityPolicy{AutoPromoteEnabled: false}
	err := svc.checkAndPromoteToMod(context.Background(), contribution, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contribution.PromotedToMod {
		t.Fatal("expected promoted_to_mod false when auto-promote disabled")
	}
}

func TestCheckAndPromoteToMod_NilCommunityRepo(t *testing.T) {
	svc := &ContributionService{} // communityRepo is nil
	contribution := &models.MemberContribution{ContributionScore: 5000}
	policy := &models.CommunityPolicy{
		AutoPromoteEnabled:          true,
		ModeratorPromotionThreshold: 100,
	}
	err := svc.checkAndPromoteToMod(context.Background(), contribution, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contribution.PromotedToMod {
		t.Fatal("expected promoted_to_mod true when communityRepo is nil")
	}
}

func TestExtractHashtags_Empty(t *testing.T) {
	if h := extractHashtags(""); len(h) != 0 {
		t.Fatalf("expected empty, got %#v", h)
	}
	if h := extractHashtags("no hashtags here"); len(h) != 0 {
		t.Fatalf("expected empty, got %#v", h)
	}
}

func TestExtractHashtags_OnlyHash(t *testing.T) {
	// A bare # with no word should not count as a hashtag.
	h := extractHashtags("this is a # post")
	if len(h) != 0 {
		t.Fatalf("expected empty for bare #, got %#v", h)
	}
}

// ─── Integration tests (require TEST_DSN) ──────────────────────────────────

func connectAndMigrate(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		t.Skip("TEST_DSN not set; skipping DB-dependent test")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Community{},
		&models.Role{},
		&models.UserRole{},
		&models.User{},
		&models.MemberContribution{},
		&models.CommunityPolicy{},
		&models.CommunityChallenge{},
		&models.ChallengeParticipant{},
	); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM challenge_participants")
		db.Exec("DELETE FROM community_challenges")
		db.Exec("DELETE FROM community_policies")
		db.Exec("DELETE FROM member_contributions")
		db.Exec("DELETE FROM user_roles")
		db.Exec("DELETE FROM users")
		db.Exec("DELETE FROM roles")
		db.Exec("DELETE FROM communities")
	})
	return db
}

func seedRoles(t *testing.T, db *gorm.DB) map[models.RoleName]string {
	entries := []struct {
		name models.RoleName
		desc string
	}{
		{models.RoleGroupMember, "Group member"},
		{models.RoleGroupMod, "Group moderator"},
		{models.RoleGroupAdmin, "Group admin"},
		{models.RoleCommunityAdmin, "Community admin"},
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

type testSeed struct {
	DB          *gorm.DB
	CommunityID string
	UserID      string
	RoleIDs     map[models.RoleName]string
	Service     *ContributionService
}

func newTestSeed(t *testing.T) testSeed {
	db := connectAndMigrate(t)
	roleIDs := seedRoles(t, db)

	userID := utils.GenerateUUID()
	if err := db.Create(&models.User{
		ID:       userID,
		Username: "testuser_" + userID[:8],
		Email:    "test@example.com",
		Status:   models.UserStatusActive,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	communityID := utils.GenerateUUID()
	if err := db.Create(&models.Community{
		ID:   communityID,
		Name: "Test Community",
	}).Error; err != nil {
		t.Fatalf("create community: %v", err)
	}

	scopeType := models.ScopeTypeCommunity
	if err := db.Create(&models.UserRole{
		ID:        utils.GenerateUUID(),
		UserID:    userID,
		RoleID:    roleIDs[models.RoleGroupMember],
		ScopeID:   &communityID,
		ScopeType: &scopeType,
		AssignedAt: time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("assign user role: %v", err)
	}

	// Pre-create contribution + policy so ensure* calls are no-ops.
	if err := db.Create(&models.MemberContribution{
		ID:          utils.GenerateUUID(),
		CommunityID: communityID,
		UserID:      userID,
		LastCalculatedAt: time.Now().UTC(),
		CreatedAt:   time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("create contribution: %v", err)
	}
	if err := db.Create(&models.CommunityPolicy{
		ID:          utils.GenerateUUID(),
		CommunityID: communityID,
		PostWeight:  10,
		CommentWeight: 5,
		ReactionWeight: 2,
		EventWeight: 20,
		TopContributorThreshold:     2500,
		ModeratorPromotionThreshold: 999999,
		AutoPromoteEnabled:          false,
		BadgeEnabled:                false,
	}).Error; err != nil {
		t.Fatalf("create policy: %v", err)
	}

	contributionRepo := repository.NewContributionRepository(db)
	communityRepo := repository.NewCommunityRepository(db)
	validation := validations.NewContributionValidation()
	groupRole := utils.NewGroupRoleChecker(communityRepo.GetUserRole)
	svc := &ContributionService{
		contributionRepo: contributionRepo,
		communityRepo:    communityRepo,
		validation:       validation,
		groupRole:        groupRole,
	}

	return testSeed{
		DB:          db,
		CommunityID: communityID,
		UserID:      userID,
		RoleIDs:     roleIDs,
		Service:     svc,
	}
}

func seedChallenge(t *testing.T, db *gorm.DB, communityID, hashtag string, maxParticipants *int) string {
	challengeID := utils.GenerateUUID()
	now := time.Now().UTC()
	if err := db.Create(&models.CommunityChallenge{
		ID:              challengeID,
		CommunityID:     communityID,
		CreatorID:       "creator-id",
		Title:           "Test Challenge",
		Description:     "A test challenge",
		Hashtag:         hashtag,
		PointsPerPost:   10,
		StartDate:       now.Add(-24 * time.Hour),
		EndDate:         now.Add(24 * time.Hour),
		MaxParticipants: maxParticipants,
		Status:          models.ChallengeStatusActive,
		CreatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	return challengeID
}

func getContribution(t *testing.T, db *gorm.DB, communityID, userID string) *models.MemberContribution {
	var c models.MemberContribution
	if err := db.Where("community_id = ? AND user_id = ?", communityID, userID).First(&c).Error; err != nil {
		t.Fatalf("get contribution: %v", err)
	}
	return &c
}

// ───── T2: IncrementValidPosts ─────────────────────────────────────────────

func TestIncrementValidPosts(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	if err := seed.Service.IncrementValidPosts(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("IncrementValidPosts: %v", err)
	}

	c := getContribution(t, seed.DB, seed.CommunityID, seed.UserID)
	if c.ValidPosts != 1 {
		t.Errorf("ValidPosts = %d, want 1", c.ValidPosts)
	}
	if c.ContributionScore != 10 {
		t.Errorf("ContributionScore = %d, want 10 (PostWeight=10)", c.ContributionScore)
	}
}

func TestIncrementValidPosts_Multiple(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if err := seed.Service.IncrementValidPosts(ctx, seed.CommunityID, seed.UserID); err != nil {
			t.Fatalf("IncrementValidPosts #%d: %v", i+1, err)
		}
	}

	c := getContribution(t, seed.DB, seed.CommunityID, seed.UserID)
	if c.ValidPosts != 5 {
		t.Errorf("ValidPosts = %d, want 5", c.ValidPosts)
	}
	if c.ContributionScore != 50 {
		t.Errorf("ContributionScore = %d, want 50", c.ContributionScore)
	}
}

// ───── T3: JoinChallenge ───────────────────────────────────────────────────

func TestJoinChallenge_Success(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	challengeID := seedChallenge(t, seed.DB, seed.CommunityID, "#Test", nil)

	if err := seed.Service.JoinChallenge(ctx, seed.UserID, challengeID); err != nil {
		t.Fatalf("JoinChallenge: %v", err)
	}

	// Verify participant exists.
	var count int64
	seed.DB.Model(&models.ChallengeParticipant{}).
		Where("challenge_id = ? AND user_id = ?", challengeID, seed.UserID).
		Count(&count)
	if count != 1 {
		t.Errorf("participant count = %d, want 1", count)
	}
}

func TestJoinChallenge_AlreadyJoined(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	challengeID := seedChallenge(t, seed.DB, seed.CommunityID, "#Test", nil)

	if err := seed.Service.JoinChallenge(ctx, seed.UserID, challengeID); err != nil {
		t.Fatalf("first join: %v", err)
	}
	if err := seed.Service.JoinChallenge(ctx, seed.UserID, challengeID); err == nil {
		t.Fatal("expected error on second join, got nil")
	} else {
		if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.Code != errorsapp.ErrCodeContribAlreadyJoined {
			t.Errorf("error = %v, want ErrCodeContribAlreadyJoined", err)
		}
	}
}

func TestJoinChallenge_NotFound(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	err := seed.Service.JoinChallenge(ctx, seed.UserID, "nonexistent-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.Code != errorsapp.ErrCodeContribChallengeNotFound {
		t.Errorf("error = %v, want ErrCodeContribChallengeNotFound", err)
	}
}

func TestJoinChallenge_Inactive(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	challengeID := utils.GenerateUUID()
	now := time.Now().UTC()
	if err := seed.DB.Create(&models.CommunityChallenge{
		ID:          challengeID,
		CommunityID: seed.CommunityID,
		Title:       "Inactive Challenge",
		Hashtag:     "#Inactive",
		StartDate:   now.Add(-48 * time.Hour),
		EndDate:     now.Add(-24 * time.Hour),
		Status:      models.ChallengeStatusCompleted,
		CreatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create completed challenge: %v", err)
	}

	err := seed.Service.JoinChallenge(ctx, seed.UserID, challengeID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.Code != errorsapp.ErrCodeContribChallengeInactive {
		t.Errorf("error = %v, want ErrCodeContribChallengeInactive", err)
	}
}

func TestJoinChallenge_NotStarted(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	challengeID := utils.GenerateUUID()
	now := time.Now().UTC()
	if err := seed.DB.Create(&models.CommunityChallenge{
		ID:          challengeID,
		CommunityID: seed.CommunityID,
		Title:       "Future Challenge",
		Hashtag:     "#Future",
		StartDate:   now.Add(24 * time.Hour),
		EndDate:     now.Add(48 * time.Hour),
		Status:      models.ChallengeStatusActive,
		CreatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create future challenge: %v", err)
	}

	err := seed.Service.JoinChallenge(ctx, seed.UserID, challengeID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.Code != errorsapp.ErrCodeContribChallengeNotStarted {
		t.Errorf("error = %v, want ErrCodeContribChallengeNotStarted", err)
	}
}

func TestJoinChallenge_Ended(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	challengeID := utils.GenerateUUID()
	now := time.Now().UTC()
	if err := seed.DB.Create(&models.CommunityChallenge{
		ID:          challengeID,
		CommunityID: seed.CommunityID,
		Title:       "Expired Challenge",
		Hashtag:     "#Expired",
		StartDate:   now.Add(-48 * time.Hour),
		EndDate:     now.Add(-1 * time.Hour),
		Status:      models.ChallengeStatusActive,
		CreatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create expired challenge: %v", err)
	}

	err := seed.Service.JoinChallenge(ctx, seed.UserID, challengeID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.Code != errorsapp.ErrCodeContribChallengeEnded {
		t.Errorf("error = %v, want ErrCodeContribChallengeEnded", err)
	}
}

func TestJoinChallenge_ParticipantLimitHit(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	maxP := 1
	challengeID := seedChallenge(t, seed.DB, seed.CommunityID, "#Limited", &maxP)

	// First user joins — succeeds.
	otherUserID := utils.GenerateUUID()
	if err := seed.DB.Create(&models.User{
		ID:       otherUserID,
		Username: "other_" + otherUserID[:8],
		Email:    "other@example.com",
		Status:   models.UserStatusActive,
	}).Error; err != nil {
		t.Fatalf("create other user: %v", err)
	}
	scopeType := models.ScopeTypeCommunity
	if err := seed.DB.Create(&models.UserRole{
		ID:         utils.GenerateUUID(),
		UserID:     otherUserID,
		RoleID:     seed.RoleIDs[models.RoleGroupMember],
		ScopeID:    &seed.CommunityID,
		ScopeType:  &scopeType,
		AssignedAt: time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("assign other user role: %v", err)
	}

	if err := seed.Service.JoinChallenge(ctx, otherUserID, challengeID); err != nil {
		t.Fatalf("first user join: %v", err)
	}

	// Second user (seed.UserID) tries to join — limit hit.
	err := seed.Service.JoinChallenge(ctx, seed.UserID, challengeID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.Code != errorsapp.ErrCodeContribParticipantLimitHit {
		t.Errorf("error = %v, want ErrCodeContribParticipantLimitHit", err)
	}
}

func TestJoinChallenge_NotMember(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	challengeID := seedChallenge(t, seed.DB, seed.CommunityID, "#Test", nil)

	nonMemberID := utils.GenerateUUID()
	if err := seed.DB.Create(&models.User{
		ID:       nonMemberID,
		Username: "nonmember_" + nonMemberID[:8],
		Email:    "nonmember@example.com",
		Status:   models.UserStatusActive,
	}).Error; err != nil {
		t.Fatalf("create non-member user: %v", err)
	}

	err := seed.Service.JoinChallenge(ctx, nonMemberID, challengeID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, validations.ErrNotCommunityMember) {
		t.Errorf("error = %v, want ErrNotCommunityMember", err)
	}
}

// ───── T1: ProcessChallengePost ────────────────────────────────────────────

func TestProcessChallengePost_HappyPath(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	challengeID := seedChallenge(t, seed.DB, seed.CommunityID, "#PhotoChallenge", nil)

	err := seed.Service.ProcessChallengePost(ctx, seed.CommunityID, seed.UserID,
		"Check out my #PhotoChallenge entry!")
	if err != nil {
		t.Fatalf("ProcessChallengePost: %v", err)
	}

	var participant models.ChallengeParticipant
	if err := seed.DB.Where("challenge_id = ? AND user_id = ?", challengeID, seed.UserID).
		First(&participant).Error; err != nil {
		t.Fatalf("get participant: %v", err)
	}
	if participant.PostsCount != 1 {
		t.Errorf("PostsCount = %d, want 1", participant.PostsCount)
	}
	if participant.TotalPointsEarned != 10 {
		t.Errorf("TotalPointsEarned = %d, want 10", participant.TotalPointsEarned)
	}

	c := getContribution(t, seed.DB, seed.CommunityID, seed.UserID)
	if c.EventParticipations != 1 {
		t.Errorf("EventParticipations = %d, want 1", c.EventParticipations)
	}
}

func TestProcessChallengePost_EmptyContent(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	if err := seed.Service.ProcessChallengePost(ctx, seed.CommunityID, seed.UserID, ""); err != nil {
		t.Fatalf("expected nil for empty content, got %v", err)
	}
	if err := seed.Service.ProcessChallengePost(ctx, seed.CommunityID, seed.UserID, "   "); err != nil {
		t.Fatalf("expected nil for whitespace content, got %v", err)
	}
}

func TestProcessChallengePost_NoHashtags(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	if err := seed.Service.ProcessChallengePost(ctx, seed.CommunityID, seed.UserID,
		"Just a regular post without hashtags"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestProcessChallengePost_NoMatchingChallenge(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	seedChallenge(t, seed.DB, seed.CommunityID, "#Golang", nil)

	err := seed.Service.ProcessChallengePost(ctx, seed.CommunityID, seed.UserID,
		"I love #Python!")
	if err != nil {
		t.Fatalf("expected nil for non-matching hashtag, got %v", err)
	}
}

func TestProcessChallengePost_AlreadyJoined(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	seedChallenge(t, seed.DB, seed.CommunityID, "#PhotoChallenge", nil)

	// First call joins the challenge.
	if err := seed.Service.ProcessChallengePost(ctx, seed.CommunityID, seed.UserID,
		"My first #PhotoChallenge post!"); err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Second call should increment posts_count but not error.
	if err := seed.Service.ProcessChallengePost(ctx, seed.CommunityID, seed.UserID,
		"My second #PhotoChallenge post!"); err != nil {
		t.Fatalf("second call: %v", err)
	}

	var participant models.ChallengeParticipant
	if err := seed.DB.Where("challenge_id = ? AND user_id = ?", seed.CommunityID+"_challenge", seed.UserID).
		First(&participant).Error; err != nil {
		// Need actual challenge ID, let me look it up differently.
	}
	// Re-fetch properly.
	seed.DB.Where("hashtag = ? AND community_id = ?", "#PhotoChallenge", seed.CommunityID).First(&models.CommunityChallenge{})
	var chall models.CommunityChallenge
	if err := seed.DB.Where("hashtag = ? AND community_id = ?", "#PhotoChallenge", seed.CommunityID).First(&chall).Error; err != nil {
		t.Fatalf("find challenge: %v", err)
	}
	if err := seed.DB.Where("challenge_id = ? AND user_id = ?", chall.ID, seed.UserID).
		First(&participant).Error; err != nil {
		t.Fatalf("get participant: %v", err)
	}
	if participant.PostsCount != 2 {
		t.Errorf("PostsCount = %d, want 2", participant.PostsCount)
	}

	c := getContribution(t, seed.DB, seed.CommunityID, seed.UserID)
	if c.EventParticipations != 2 {
		t.Errorf("EventParticipations = %d, want 2", c.EventParticipations)
	}
}

func TestProcessChallengePost_ChallengeFull(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	maxP := 0
	challengeID := seedChallenge(t, seed.DB, seed.CommunityID, "#Full", &maxP)

	// Manually fill the challenge by inserting a participant.
	if err := seed.DB.Create(&models.ChallengeParticipant{
		ID:          utils.GenerateUUID(),
		ChallengeID: challengeID,
		UserID:      "filler-user",
		JoinedAt:    time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("seed participant: %v", err)
	}

	// Now the challenge is full (maxParticipants=0, but there's already 1 participant).
	// Actually, JoinChallengeAtomic checks count >= maxParticipants.
	// With maxParticipants=0, count(1) >= 0 → limit hit.
	err := seed.Service.ProcessChallengePost(ctx, seed.CommunityID, seed.UserID,
		"Trying to join #Full challenge")
	if err != nil {
		t.Fatalf("expected nil (silently skipped), got %v", err)
	}

	// Verify user did NOT become a participant.
	var count int64
	seed.DB.Model(&models.ChallengeParticipant{}).
		Where("challenge_id = ? AND user_id = ?", challengeID, seed.UserID).
		Count(&count)
	if count != 0 {
		t.Errorf("participant count = %d, want 0 (silently skipped)", count)
	}
}

func TestProcessChallengePost_ExpiredChallenge(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	now := time.Now().UTC()
	challengeID := utils.GenerateUUID()
	if err := seed.DB.Create(&models.CommunityChallenge{
		ID:            challengeID,
		CommunityID:   seed.CommunityID,
		Title:         "Expired",
		Hashtag:       "#Expired",
		PointsPerPost: 10,
		StartDate:     now.Add(-48 * time.Hour),
		EndDate:       now.Add(-1 * time.Hour),
		Status:        models.ChallengeStatusActive,
		CreatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create expired challenge: %v", err)
	}

	err := seed.Service.ProcessChallengePost(ctx, seed.CommunityID, seed.UserID,
		"Late post for #Expired challenge")
	if err != nil {
		t.Fatalf("expected nil for expired challenge, got %v", err)
	}

	var count int64
	seed.DB.Model(&models.ChallengeParticipant{}).
		Where("challenge_id = ? AND user_id = ?", challengeID, seed.UserID).
		Count(&count)
	if count != 0 {
		t.Errorf("participant count = %d, want 0 (expired challenge skipped)", count)
	}
}

// ───── Race condition C2 ────────────────────────────────────────────────────

func TestConcurrentIncrementValidPosts(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()
	const goroutines = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			seed.Service.IncrementValidPosts(ctx, seed.CommunityID, seed.UserID)
		}()
	}
	wg.Wait()

	c := getContribution(t, seed.DB, seed.CommunityID, seed.UserID)
	if c.ValidPosts != goroutines {
		t.Errorf("ValidPosts = %d, want %d (lost updates: %d)", c.ValidPosts, goroutines, goroutines-c.ValidPosts)
	}
}

func TestConcurrentIncrementAndChallenge(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	seedChallenge(t, seed.DB, seed.CommunityID, "#LinkUp", nil)
	const goroutines = 10

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			seed.Service.IncrementValidPosts(ctx, seed.CommunityID, seed.UserID)
		}()
		go func() {
			defer wg.Done()
			seed.Service.ProcessChallengePost(ctx, seed.CommunityID, seed.UserID,
				"Check out my #LinkUp photo!")
		}()
	}
	wg.Wait()

	c := getContribution(t, seed.DB, seed.CommunityID, seed.UserID)
	if c.ValidPosts != goroutines {
		t.Errorf("ValidPosts = %d, want %d (lost: %d)", c.ValidPosts, goroutines, goroutines-c.ValidPosts)
	}
	if c.EventParticipations != goroutines {
		t.Errorf("EventParticipations = %d, want %d (lost: %d)", c.EventParticipations, goroutines, goroutines-c.EventParticipations)
	}
}

func TestConcurrentJoinChallenge_LimitRace(t *testing.T) {
	seed := newTestSeed(t)
	ctx := context.Background()

	maxP := 2
	challengeID := seedChallenge(t, seed.DB, seed.CommunityID, "#Race", &maxP)

	// Create 5 members, all trying to join a challenge with maxParticipants=2.
	memberIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		uid := utils.GenerateUUID()
		if err := seed.DB.Create(&models.User{
			ID:       uid,
			Username: "racer_" + uid[:8],
			Email:    "racer" + uid[:8] + "@example.com",
			Status:   models.UserStatusActive,
		}).Error; err != nil {
			t.Fatalf("create racer %d: %v", i, err)
		}
		scopeType := models.ScopeTypeCommunity
		if err := seed.DB.Create(&models.UserRole{
			ID:         utils.GenerateUUID(),
			UserID:     uid,
			RoleID:     seed.RoleIDs[models.RoleGroupMember],
			ScopeID:    &seed.CommunityID,
			ScopeType:  &scopeType,
			AssignedAt: time.Now().UTC(),
		}).Error; err != nil {
			t.Fatalf("assign racer role %d: %v", i, err)
		}
		memberIDs[i] = uid
	}

	var wg sync.WaitGroup
	wg.Add(5)
	for _, uid := range memberIDs {
		uid := uid
		go func() {
			defer wg.Done()
			seed.Service.JoinChallenge(ctx, uid, challengeID)
		}()
	}
	wg.Wait()

	var count int64
	seed.DB.Model(&models.ChallengeParticipant{}).
		Where("challenge_id = ?", challengeID).
		Count(&count)
	if count != int64(maxP) {
		t.Errorf("participants = %d, want %d (over-limit: %d)", count, maxP, count-int64(maxP))
	}
}
