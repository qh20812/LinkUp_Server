package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"linkup/dto"
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

// ───── Phase 3: Full-stack seed (with notificationService + profileRepo) ────

type fullTestSeed struct {
	DB          *gorm.DB
	CommunityID string
	UserID      string
	RoleIDs     map[models.RoleName]string
	Service     *ContributionService
}

func newFullTestSeed(t *testing.T) fullTestSeed {
	t.Helper()
	db := connectAndMigrate(t)

	// Add extra tables for notification + profile support.
	if err := db.AutoMigrate(
		&models.Notification{},
		&models.NotificationPreference{},
		&models.Profile{},
	); err != nil {
		t.Fatalf("auto-migrate extra tables: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM profiles")
		db.Exec("DELETE FROM notification_preferences")
		db.Exec("DELETE FROM notifications")
	})

	roleIDs := seedRoles(t, db)

	userID := utils.GenerateUUID()
	if err := db.Create(&models.User{
		ID:       userID,
		Username: "fulltestuser_" + userID[:8],
		Email:    "fulltest@example.com",
		Status:   models.UserStatusActive,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	communityID := utils.GenerateUUID()
	if err := db.Create(&models.Community{
		ID:   communityID,
		Name: "Full Test Community",
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

	contributionRepo := repository.NewContributionRepository(db)
	communityRepo := repository.NewCommunityRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	prefRepo := repository.NewNotificationPreferenceRepository(db)
	validation := validations.NewContributionValidation()
	notifService := NewNotificationService(notifRepo, prefRepo, profileRepo, nil)
	svc := NewContributionService(contributionRepo, communityRepo, profileRepo, notifService, validation)

	return fullTestSeed{
		DB:          db,
		CommunityID: communityID,
		UserID:      userID,
		RoleIDs:     roleIDs,
		Service:     svc,
	}
}

// ───── T4: GetPolicy ──────────────────────────────────────────────────────

func TestGetPolicy_CreatesDefault(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	policy, err := seed.Service.GetPolicy(ctx, seed.CommunityID, seed.UserID)
	if err != nil {
		t.Fatalf("GetPolicy: %v", err)
	}
	if policy.CommunityID != seed.CommunityID {
		t.Errorf("CommunityID = %s, want %s", policy.CommunityID, seed.CommunityID)
	}
	if policy.PostWeight != 10 || policy.CommentWeight != 5 || policy.ReactionWeight != 2 || policy.EventWeight != 20 {
		t.Errorf("default weights = [%d %d %d %d], want [10 5 2 20]",
			policy.PostWeight, policy.CommentWeight, policy.ReactionWeight, policy.EventWeight)
	}
	if policy.TopContributorThreshold != 2500 || policy.ModeratorPromotionThreshold != 5000 {
		t.Errorf("thresholds = [%d %d], want [2500 5000]",
			policy.TopContributorThreshold, policy.ModeratorPromotionThreshold)
	}

	// Second call returns the same policy (no duplicate).
	policy2, err := seed.Service.GetPolicy(ctx, seed.CommunityID, seed.UserID)
	if err != nil {
		t.Fatalf("second GetPolicy: %v", err)
	}
	if policy.ID != policy2.ID {
		t.Errorf("duplicate policy created: id %s vs %s", policy.ID, policy2.ID)
	}
}

func TestGetPolicy_WithExistingPolicy(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	// Seed an existing policy with custom weights.
	if err := seed.DB.Create(&models.CommunityPolicy{
		ID:          utils.GenerateUUID(),
		CommunityID: seed.CommunityID,
		PostWeight:  25,
		CommentWeight: 10,
		ReactionWeight: 5,
		EventWeight:   30,
		TopContributorThreshold:     5000,
		ModeratorPromotionThreshold: 10000,
		AutoPromoteEnabled:          false,
		BadgeEnabled:                false,
		CreatedAt:                   time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("create policy: %v", err)
	}

	policy, err := seed.Service.GetPolicy(ctx, seed.CommunityID, seed.UserID)
	if err != nil {
		t.Fatalf("GetPolicy: %v", err)
	}
	if policy.PostWeight != 25 {
		t.Errorf("PostWeight = %d, want 25", policy.PostWeight)
	}
}

func TestGetPolicy_CommunityNotFound(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.GetPolicy(ctx, "nonexistent", seed.UserID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, validations.ErrCommunityNotFound) {
		t.Errorf("error = %v, want ErrCommunityNotFound", err)
	}
}

func TestGetPolicy_NotMember(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	outsiderID := utils.GenerateUUID()
	if err := seed.DB.Create(&models.User{
		ID:       outsiderID,
		Username: "outsider_" + outsiderID[:8],
		Email:    "outsider@example.com",
		Status:   models.UserStatusActive,
	}).Error; err != nil {
		t.Fatalf("create outsider: %v", err)
	}

	_, err := seed.Service.GetPolicy(ctx, seed.CommunityID, outsiderID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, validations.ErrNotCommunityMember) {
		t.Errorf("error = %v, want ErrNotCommunityMember", err)
	}
}

// ───── T5: UpdatePolicy ───────────────────────────────────────────────────

func TestUpdatePolicy_HappyPath(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()
	scopeType := models.ScopeTypeCommunity

	// Make user an admin.
	if err := seed.DB.Create(&models.UserRole{
		ID:        utils.GenerateUUID(),
		UserID:    seed.UserID,
		RoleID:    seed.RoleIDs[models.RoleGroupAdmin],
		ScopeID:   &seed.CommunityID,
		ScopeType: &scopeType,
		AssignedAt: time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("assign admin role: %v", err)
	}

	err := seed.Service.UpdatePolicy(ctx, seed.UserID, seed.CommunityID, dto.UpdatePolicyInput{
		PostWeight:                  20,
		CommentWeight:               10,
		ReactionWeight:              5,
		EventWeight:                 40,
		TopContributorThreshold:     3000,
		ModeratorPromotionThreshold: 6000,
		AutoPromoteEnabled:          true,
		BadgeEnabled:                true,
	})
	if err != nil {
		t.Fatalf("UpdatePolicy: %v", err)
	}

	// Verify updated.
	policy, err := seed.Service.GetPolicy(ctx, seed.CommunityID, seed.UserID)
	if err != nil {
		t.Fatalf("GetPolicy after update: %v", err)
	}
	if policy.PostWeight != 20 || policy.CommentWeight != 10 || policy.ReactionWeight != 5 || policy.EventWeight != 40 {
		t.Errorf("weights = [%d %d %d %d], want [20 10 5 40]",
			policy.PostWeight, policy.CommentWeight, policy.ReactionWeight, policy.EventWeight)
	}
	if policy.TopContributorThreshold != 3000 {
		t.Errorf("TopContributorThreshold = %d, want 3000", policy.TopContributorThreshold)
	}
	if !policy.AutoPromoteEnabled {
		t.Error("AutoPromoteEnabled should be true")
	}
}

func TestUpdatePolicy_CommunityNotFound(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	err := seed.Service.UpdatePolicy(ctx, seed.UserID, "nonexistent", dto.UpdatePolicyInput{
		PostWeight:                  10,
		CommentWeight:               5,
		ReactionWeight:              2,
		EventWeight:                 20,
		TopContributorThreshold:     2500,
		ModeratorPromotionThreshold: 5000,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, validations.ErrCommunityNotFound) {
		t.Errorf("error = %v, want ErrCommunityNotFound", err)
	}
}

func TestUpdatePolicy_NotAdmin(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	// User is only a Member, not Admin.
	err := seed.Service.UpdatePolicy(ctx, seed.UserID, seed.CommunityID, dto.UpdatePolicyInput{
		PostWeight:                  10,
		CommentWeight:               5,
		ReactionWeight:              2,
		EventWeight:                 20,
		TopContributorThreshold:     2500,
		ModeratorPromotionThreshold: 5000,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, validations.ErrNotCommunityAdmin) {
		t.Errorf("error = %v, want ErrNotCommunityAdmin", err)
	}
}

func TestUpdatePolicy_ValidationWeightsInvalid(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()
	scopeType := models.ScopeTypeCommunity

	if err := seed.DB.Create(&models.UserRole{
		ID:        utils.GenerateUUID(),
		UserID:    seed.UserID,
		RoleID:    seed.RoleIDs[models.RoleGroupAdmin],
		ScopeID:   &seed.CommunityID,
		ScopeType: &scopeType,
		AssignedAt: time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("assign admin role: %v", err)
	}

	err := seed.Service.UpdatePolicy(ctx, seed.UserID, seed.CommunityID, dto.UpdatePolicyInput{
		PostWeight:                  -1,
		CommentWeight:               5,
		ReactionWeight:              2,
		EventWeight:                 20,
		TopContributorThreshold:     2500,
		ModeratorPromotionThreshold: 5000,
	})
	if err == nil {
		t.Fatal("expected error for negative weight, got nil")
	}
	if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.Code != errorsapp.ErrCodeContribPostWeightInvalid {
		t.Errorf("error = %v, want ErrCodeContribPostWeightInvalid", err)
	}
}

func TestUpdatePolicy_ValidationThresholdOrderInvalid(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()
	scopeType := models.ScopeTypeCommunity

	if err := seed.DB.Create(&models.UserRole{
		ID:        utils.GenerateUUID(),
		UserID:    seed.UserID,
		RoleID:    seed.RoleIDs[models.RoleGroupAdmin],
		ScopeID:   &seed.CommunityID,
		ScopeType: &scopeType,
		AssignedAt: time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("assign admin role: %v", err)
	}

	err := seed.Service.UpdatePolicy(ctx, seed.UserID, seed.CommunityID, dto.UpdatePolicyInput{
		PostWeight:                  10,
		CommentWeight:               5,
		ReactionWeight:              2,
		EventWeight:                 20,
		TopContributorThreshold:     5000,
		ModeratorPromotionThreshold: 2500,
	})
	if err == nil {
		t.Fatal("expected error for threshold order, got nil")
	}
	if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.Code != errorsapp.ErrCodeContribThresholdOrderInvalid {
		t.Errorf("error = %v, want ErrCodeContribThresholdOrderInvalid", err)
	}
}

// ───── T6: GetLeaderboard ────────────────────────────────────────────────

func TestGetLeaderboard_CommunityNotFound(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.GetLeaderboard(ctx, "nonexistent", 1, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, validations.ErrCommunityNotFound) {
		t.Errorf("error = %v, want ErrCommunityNotFound", err)
	}
}

func TestGetLeaderboard_Empty(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	items, err := seed.Service.GetLeaderboard(ctx, seed.CommunityID, 1, 10)
	if err != nil {
		t.Fatalf("GetLeaderboard: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestGetLeaderboard_WithData(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	// Create 3 contributions with different scores.
	for i, score := range []int{100, 300, 200} {
		uid := utils.GenerateUUID()
		if err := seed.DB.Create(&models.User{
			ID:       uid,
			Username: fmt.Sprintf("lbuser%d_%s", i, uid[:8]),
			Email:    fmt.Sprintf("lb%d@example.com", i),
			Status:   models.UserStatusActive,
		}).Error; err != nil {
			t.Fatalf("create user %d: %v", i, err)
		}
		if err := seed.DB.Create(&models.MemberContribution{
			ID:                  utils.GenerateUUID(),
			CommunityID:         seed.CommunityID,
			UserID:              uid,
			ContributionScore:   score,
			LastCalculatedAt:    time.Now().UTC(),
			CreatedAt:           time.Now().UTC(),
		}).Error; err != nil {
			t.Fatalf("create contribution %d: %v", i, err)
		}
	}

	items, err := seed.Service.GetLeaderboard(ctx, seed.CommunityID, 1, 10)
	if err != nil {
		t.Fatalf("GetLeaderboard: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}
	// Should be sorted by score descending.
	if items[0].ContributionScore < items[1].ContributionScore || items[1].ContributionScore < items[2].ContributionScore {
		t.Errorf("leaderboard not sorted: scores = [%d %d %d]",
			items[0].ContributionScore, items[1].ContributionScore, items[2].ContributionScore)
	}
}

func TestGetLeaderboard_Pagination(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		uid := utils.GenerateUUID()
		if err := seed.DB.Create(&models.User{
			ID:       uid,
			Username: fmt.Sprintf("pageuser%d_%s", i, uid[:8]),
			Email:    fmt.Sprintf("page%d@example.com", i),
			Status:   models.UserStatusActive,
		}).Error; err != nil {
			t.Fatalf("create user %d: %v", i, err)
		}
		if err := seed.DB.Create(&models.MemberContribution{
			ID:                  utils.GenerateUUID(),
			CommunityID:         seed.CommunityID,
			UserID:              uid,
			ContributionScore:   (i + 1) * 100,
			LastCalculatedAt:    time.Now().UTC(),
			CreatedAt:           time.Now().UTC(),
		}).Error; err != nil {
			t.Fatalf("create contribution %d: %v", i, err)
		}
	}

	// Page 1: 2 items.
	items, err := seed.Service.GetLeaderboard(ctx, seed.CommunityID, 1, 2)
	if err != nil {
		t.Fatalf("GetLeaderboard page 1: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("page 1: got %d items, want 2", len(items))
	}

	// Page 2: 2 items.
	items2, err := seed.Service.GetLeaderboard(ctx, seed.CommunityID, 2, 2)
	if err != nil {
		t.Fatalf("GetLeaderboard page 2: %v", err)
	}
	if len(items2) != 2 {
		t.Errorf("page 2: got %d items, want 2", len(items2))
	}

	// Page 3: 1 item.
	items3, err := seed.Service.GetLeaderboard(ctx, seed.CommunityID, 3, 2)
	if err != nil {
		t.Fatalf("GetLeaderboard page 3: %v", err)
	}
	if len(items3) != 1 {
		t.Errorf("page 3: got %d items, want 1", len(items3))
	}
}

// ───── T7: GetContribution ───────────────────────────────────────────────

func TestGetContribution_CreatesOnDemand(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	contribution, err := seed.Service.GetContribution(ctx, seed.CommunityID, seed.UserID)
	if err != nil {
		t.Fatalf("GetContribution: %v", err)
	}
	if contribution.CommunityID != seed.CommunityID {
		t.Errorf("CommunityID = %s, want %s", contribution.CommunityID, seed.CommunityID)
	}
	if contribution.UserID != seed.UserID {
		t.Errorf("UserID = %s, want %s", contribution.UserID, seed.UserID)
	}
	if contribution.ValidPosts != 0 || contribution.ContributionScore != 0 {
		t.Errorf("expected zeroed contribution, got posts=%d score=%d",
			contribution.ValidPosts, contribution.ContributionScore)
	}

	// Second call returns same record.
	c2, err := seed.Service.GetContribution(ctx, seed.CommunityID, seed.UserID)
	if err != nil {
		t.Fatalf("second GetContribution: %v", err)
	}
	if contribution.ID != c2.ID {
		t.Errorf("duplicate contribution: id %s vs %s", contribution.ID, c2.ID)
	}
}

func TestGetContribution_CommunityNotFound(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.GetContribution(ctx, "nonexistent", seed.UserID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, validations.ErrCommunityNotFound) {
		t.Errorf("error = %v, want ErrCommunityNotFound", err)
	}
}

// ───── T8: GetContributionResponse ────────────────────────────────────────

func TestGetContributionResponse_WithProfile(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	// Seed a profile for the user.
	if err := seed.DB.Create(&models.Profile{
		ID:          utils.GenerateUUID(),
		UserID:      seed.UserID,
		DisplayName: "Test User",
		AvatarURI:   "https://cdn.example.com/avatar.jpg",
	}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}

	resp, err := seed.Service.GetContributionResponse(ctx, seed.CommunityID, seed.UserID)
	if err != nil {
		t.Fatalf("GetContributionResponse: %v", err)
	}
	if resp.DisplayName != "Test User" {
		t.Errorf("DisplayName = %q, want %q", resp.DisplayName, "Test User")
	}
	if resp.AvatarURI != "https://cdn.example.com/avatar.jpg" {
		t.Errorf("AvatarURI = %q, want %q", resp.AvatarURI, "https://cdn.example.com/avatar.jpg")
	}
	if resp.UserID != seed.UserID {
		t.Errorf("UserID = %s, want %s", resp.UserID, seed.UserID)
	}
}

func TestGetContributionResponse_NoProfile(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	resp, err := seed.Service.GetContributionResponse(ctx, seed.CommunityID, seed.UserID)
	if err != nil {
		t.Fatalf("GetContributionResponse: %v", err)
	}
	if resp.DisplayName != "" {
		t.Errorf("DisplayName = %q, want empty", resp.DisplayName)
	}
}

func TestGetContributionResponse_CommunityNotFound(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.GetContributionResponse(ctx, "nonexistent", seed.UserID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, validations.ErrCommunityNotFound) {
		t.Errorf("error = %v, want ErrCommunityNotFound", err)
	}
}

// ───── T9: IncrementQualityComments / IncrementPositiveReactions / IncrementEventParticipations ─

func TestIncrementQualityComments(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	if err := seed.Service.IncrementQualityComments(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("IncrementQualityComments: %v", err)
	}
	if err := seed.Service.IncrementQualityComments(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("second IncrementQualityComments: %v", err)
	}
	if err := seed.Service.IncrementQualityComments(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("third IncrementQualityComments: %v", err)
	}

	c := getContribution(t, seed.DB, seed.CommunityID, seed.UserID)
	if c.QualityComments != 3 {
		t.Errorf("QualityComments = %d, want 3", c.QualityComments)
	}
	if c.ContributionScore != 15 {
		t.Errorf("ContributionScore = %d, want 15 (3 * CommentWeight=5)", c.ContributionScore)
	}
}

func TestIncrementPositiveReactions(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	if err := seed.Service.IncrementPositiveReactions(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("IncrementPositiveReactions: %v", err)
	}
	if err := seed.Service.IncrementPositiveReactions(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("second IncrementPositiveReactions: %v", err)
	}

	c := getContribution(t, seed.DB, seed.CommunityID, seed.UserID)
	if c.PositiveReactions != 2 {
		t.Errorf("PositiveReactions = %d, want 2", c.PositiveReactions)
	}
	if c.ContributionScore != 4 {
		t.Errorf("ContributionScore = %d, want 4 (2 * ReactionWeight=2)", c.ContributionScore)
	}
}

func TestIncrementEventParticipations(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	if err := seed.Service.IncrementEventParticipations(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("IncrementEventParticipations: %v", err)
	}
	if err := seed.Service.IncrementEventParticipations(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("second IncrementEventParticipations: %v", err)
	}
	if err := seed.Service.IncrementEventParticipations(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("third IncrementEventParticipations: %v", err)
	}
	if err := seed.Service.IncrementEventParticipations(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("fourth IncrementEventParticipations: %v", err)
	}

	c := getContribution(t, seed.DB, seed.CommunityID, seed.UserID)
	if c.EventParticipations != 4 {
		t.Errorf("EventParticipations = %d, want 4", c.EventParticipations)
	}
	if c.ContributionScore != 80 {
		t.Errorf("ContributionScore = %d, want 80 (4 * EventWeight=20)", c.ContributionScore)
	}
}

// ───── T10: DecrementQualityComments / DecrementPositiveReactions ──────────

func TestDecrementQualityComments(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	// Pre-increment to 3.
	for i := 0; i < 3; i++ {
		if err := seed.Service.IncrementQualityComments(ctx, seed.CommunityID, seed.UserID); err != nil {
			t.Fatalf("pre-increment %d: %v", i, err)
		}
	}

	// Decrement once.
	if err := seed.Service.DecrementQualityComments(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("DecrementQualityComments: %v", err)
	}

	c := getContribution(t, seed.DB, seed.CommunityID, seed.UserID)
	if c.QualityComments != 2 {
		t.Errorf("QualityComments = %d, want 2", c.QualityComments)
	}
	if c.ContributionScore != 10 {
		t.Errorf("ContributionScore = %d, want 10 (2 * CommentWeight=5)", c.ContributionScore)
	}
}

func TestDecrementQualityComments_BelowZeroFloor(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	// Decrement from 0 — should not go negative.
	if err := seed.Service.DecrementQualityComments(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("DecrementQualityComments: %v", err)
	}

	c := getContribution(t, seed.DB, seed.CommunityID, seed.UserID)
	if c.QualityComments != 0 {
		t.Errorf("QualityComments = %d, want 0 (floored)", c.QualityComments)
	}
}

func TestDecrementPositiveReactions(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	// Pre-increment to 4.
	for i := 0; i < 4; i++ {
		if err := seed.Service.IncrementPositiveReactions(ctx, seed.CommunityID, seed.UserID); err != nil {
			t.Fatalf("pre-increment %d: %v", i, err)
		}
	}

	// Decrement twice.
	if err := seed.Service.DecrementPositiveReactions(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("first DecrementPositiveReactions: %v", err)
	}
	if err := seed.Service.DecrementPositiveReactions(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("second DecrementPositiveReactions: %v", err)
	}

	c := getContribution(t, seed.DB, seed.CommunityID, seed.UserID)
	if c.PositiveReactions != 2 {
		t.Errorf("PositiveReactions = %d, want 2", c.PositiveReactions)
	}
	if c.ContributionScore != 4 {
		t.Errorf("ContributionScore = %d, want 4 (2 * ReactionWeight=2)", c.ContributionScore)
	}
}

func TestDecrementPositiveReactions_BelowZeroFloor(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	// Decrement from 0 — should not go negative.
	if err := seed.Service.DecrementPositiveReactions(ctx, seed.CommunityID, seed.UserID); err != nil {
		t.Fatalf("DecrementPositiveReactions: %v", err)
	}

	c := getContribution(t, seed.DB, seed.CommunityID, seed.UserID)
	if c.PositiveReactions != 0 {
		t.Errorf("PositiveReactions = %d, want 0 (floored)", c.PositiveReactions)
	}
}

// ───── T11: CreateChallenge ──────────────────────────────────────────────

func TestCreateChallenge_HappyPath(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()
	scopeType := models.ScopeTypeCommunity

	if err := seed.DB.Create(&models.UserRole{
		ID:        utils.GenerateUUID(),
		UserID:    seed.UserID,
		RoleID:    seed.RoleIDs[models.RoleGroupAdmin],
		ScopeID:   &seed.CommunityID,
		ScopeType: &scopeType,
		AssignedAt: time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("assign admin role: %v", err)
	}

	startDate := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	endDate := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	err := seed.Service.CreateChallenge(ctx, seed.UserID, seed.CommunityID, dto.CreateChallengeInput{
		Title:           "Photo Challenge",
		Description:     "Share your best photos",
		Hashtag:         "#PhotoChallenge",
		PointsPerPost:   15,
		StartDate:       startDate,
		EndDate:         endDate,
	})
	if err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}

	// Verify challenge was created.
	var count int64
	seed.DB.Model(&models.CommunityChallenge{}).
		Where("community_id = ? AND hashtag = ?", seed.CommunityID, "#PhotoChallenge").
		Count(&count)
	if count != 1 {
		t.Errorf("challenge count = %d, want 1", count)
	}
}

func TestCreateChallenge_CommunityNotFound(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	startDate := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	endDate := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	err := seed.Service.CreateChallenge(ctx, seed.UserID, "nonexistent", dto.CreateChallengeInput{
		Title:         "Test",
		Hashtag:       "#Test",
		StartDate:     startDate,
		EndDate:       endDate,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, validations.ErrCommunityNotFound) {
		t.Errorf("error = %v, want ErrCommunityNotFound", err)
	}
}

func TestCreateChallenge_NotAdmin(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	startDate := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	endDate := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	err := seed.Service.CreateChallenge(ctx, seed.UserID, seed.CommunityID, dto.CreateChallengeInput{
		Title:         "Test",
		Hashtag:       "#Test",
		StartDate:     startDate,
		EndDate:       endDate,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, validations.ErrNotCommunityAdmin) {
		t.Errorf("error = %v, want ErrNotCommunityAdmin", err)
	}
}

func TestCreateChallenge_EmptyTitle(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()
	scopeType := models.ScopeTypeCommunity

	if err := seed.DB.Create(&models.UserRole{
		ID:        utils.GenerateUUID(),
		UserID:    seed.UserID,
		RoleID:    seed.RoleIDs[models.RoleGroupAdmin],
		ScopeID:   &seed.CommunityID,
		ScopeType: &scopeType,
		AssignedAt: time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("assign admin role: %v", err)
	}

	startDate := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	endDate := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	err := seed.Service.CreateChallenge(ctx, seed.UserID, seed.CommunityID, dto.CreateChallengeInput{
		Title:         "",
		Hashtag:       "#Test",
		StartDate:     startDate,
		EndDate:       endDate,
	})
	if err == nil {
		t.Fatal("expected error for empty title, got nil")
	}
	if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.Code != errorsapp.ErrCodeContribTitleRequired {
		t.Errorf("error = %v, want ErrCodeContribTitleRequired", err)
	}
}

func TestCreateChallenge_TitleTooShort(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()
	scopeType := models.ScopeTypeCommunity

	if err := seed.DB.Create(&models.UserRole{
		ID:        utils.GenerateUUID(),
		UserID:    seed.UserID,
		RoleID:    seed.RoleIDs[models.RoleGroupAdmin],
		ScopeID:   &seed.CommunityID,
		ScopeType: &scopeType,
		AssignedAt: time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("assign admin role: %v", err)
	}

	startDate := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	endDate := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	err := seed.Service.CreateChallenge(ctx, seed.UserID, seed.CommunityID, dto.CreateChallengeInput{
		Title:         "Hi",
		Hashtag:       "#Test",
		StartDate:     startDate,
		EndDate:       endDate,
	})
	if err == nil {
		t.Fatal("expected error for short title, got nil")
	}
	if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.Code != errorsapp.ErrCodeContribTitleTooShort {
		t.Errorf("error = %v, want ErrCodeContribTitleTooShort", err)
	}
}

func TestCreateChallenge_HashtagInvalidFormat(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()
	scopeType := models.ScopeTypeCommunity

	if err := seed.DB.Create(&models.UserRole{
		ID:        utils.GenerateUUID(),
		UserID:    seed.UserID,
		RoleID:    seed.RoleIDs[models.RoleGroupAdmin],
		ScopeID:   &seed.CommunityID,
		ScopeType: &scopeType,
		AssignedAt: time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("assign admin role: %v", err)
	}

	startDate := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	endDate := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	err := seed.Service.CreateChallenge(ctx, seed.UserID, seed.CommunityID, dto.CreateChallengeInput{
		Title:         "Valid Title",
		Hashtag:       "NoHash",
		StartDate:     startDate,
		EndDate:       endDate,
	})
	if err == nil {
		t.Fatal("expected error for invalid hashtag format, got nil")
	}
	if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.Code != errorsapp.ErrCodeContribHashtagInvalidFormat {
		t.Errorf("error = %v, want ErrCodeContribHashtagInvalidFormat", err)
	}
}

func TestCreateChallenge_EndDateBeforeStart(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()
	scopeType := models.ScopeTypeCommunity

	if err := seed.DB.Create(&models.UserRole{
		ID:        utils.GenerateUUID(),
		UserID:    seed.UserID,
		RoleID:    seed.RoleIDs[models.RoleGroupAdmin],
		ScopeID:   &seed.CommunityID,
		ScopeType: &scopeType,
		AssignedAt: time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("assign admin role: %v", err)
	}

	startDate := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)
	endDate := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	err := seed.Service.CreateChallenge(ctx, seed.UserID, seed.CommunityID, dto.CreateChallengeInput{
		Title:         "Valid Title",
		Hashtag:       "#Test",
		StartDate:     startDate,
		EndDate:       endDate,
	})
	if err == nil {
		t.Fatal("expected error for end before start, got nil")
	}
	if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.Code != errorsapp.ErrCodeContribEndDateBeforeStart {
		t.Errorf("error = %v, want ErrCodeContribEndDateBeforeStart", err)
	}
}

// ───── T12: GetActiveChallenges ──────────────────────────────────────────

func TestGetActiveChallenges_CommunityNotFound(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.GetActiveChallenges(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, validations.ErrCommunityNotFound) {
		t.Errorf("error = %v, want ErrCommunityNotFound", err)
	}
}

func TestGetActiveChallenges_Empty(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	challenges, err := seed.Service.GetActiveChallenges(ctx, seed.CommunityID)
	if err != nil {
		t.Fatalf("GetActiveChallenges: %v", err)
	}
	if len(challenges) != 0 {
		t.Errorf("got %d challenges, want 0", len(challenges))
	}
}

func TestGetActiveChallenges_WithChallenges(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	seedChallenge(t, seed.DB, seed.CommunityID, "#Challenge1", nil)
	seedChallenge(t, seed.DB, seed.CommunityID, "#Challenge2", nil)

	challenges, err := seed.Service.GetActiveChallenges(ctx, seed.CommunityID)
	if err != nil {
		t.Fatalf("GetActiveChallenges: %v", err)
	}
	if len(challenges) != 2 {
		t.Errorf("got %d challenges, want 2", len(challenges))
	}
}

// ───── T13: GetChallengeParticipants ─────────────────────────────────────

func TestGetChallengeParticipants_NotFound(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.GetChallengeParticipants(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if appErr, ok := errorsapp.IsAppError(err); !ok || appErr.Code != errorsapp.ErrCodeContribChallengeNotFound {
		t.Errorf("error = %v, want ErrCodeContribChallengeNotFound", err)
	}
}

func TestGetChallengeParticipants_Empty(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	challengeID := seedChallenge(t, seed.DB, seed.CommunityID, "#Empty", nil)

	items, err := seed.Service.GetChallengeParticipants(ctx, challengeID)
	if err != nil {
		t.Fatalf("GetChallengeParticipants: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestGetChallengeParticipants_WithParticipants(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	challengeID := seedChallenge(t, seed.DB, seed.CommunityID, "#Participants", nil)

	// Join the challenge.
	if err := seed.Service.JoinChallenge(ctx, seed.UserID, challengeID); err != nil {
		t.Fatalf("JoinChallenge: %v", err)
	}

	items, err := seed.Service.GetChallengeParticipants(ctx, challengeID)
	if err != nil {
		t.Fatalf("GetChallengeParticipants: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].UserID != seed.UserID {
		t.Errorf("UserID = %s, want %s", items[0].UserID, seed.UserID)
	}
}

// ───── T14: RecalculateScore ─────────────────────────────────────────────

func TestRecalculateScore_CommunityNotFound(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.RecalculateScore(ctx, "nonexistent", seed.UserID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, validations.ErrCommunityNotFound) {
		t.Errorf("error = %v, want ErrCommunityNotFound", err)
	}
}

func TestRecalculateScore_HappyPath(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	// Manually set some contribution data.
	if err := seed.DB.Create(&models.MemberContribution{
		ID:                  utils.GenerateUUID(),
		CommunityID:         seed.CommunityID,
		UserID:              seed.UserID,
		ValidPosts:          5,
		QualityComments:     3,
		PositiveReactions:   10,
		EventParticipations: 2,
		ContributionScore:   0,
		LastCalculatedAt:    time.Now().UTC(),
		CreatedAt:           time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("create contribution: %v", err)
	}

	result, err := seed.Service.RecalculateScore(ctx, seed.CommunityID, seed.UserID)
	if err != nil {
		t.Fatalf("RecalculateScore: %v", err)
	}

	expectedScore := (5*10 + 3*5 + 10*2 + 2*20) // 50+15+20+40 = 125
	if result.ContributionScore != expectedScore {
		t.Errorf("ContributionScore = %d, want %d", result.ContributionScore, expectedScore)
	}
	if result.LastCalculatedAt.IsZero() {
		t.Error("LastCalculatedAt should be set")
	}
}

func TestRecalculateScore_BadgeAssignment(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	// Use a community with a low TopContributorThreshold so the badge triggers.
	if err := seed.DB.Create(&models.CommunityPolicy{
		ID:          utils.GenerateUUID(),
		CommunityID: seed.CommunityID,
		PostWeight:  10,
		CommentWeight: 5,
		ReactionWeight: 2,
		EventWeight:   20,
		TopContributorThreshold:     50,
		ModeratorPromotionThreshold: 999999,
		AutoPromoteEnabled:          false,
		BadgeEnabled:                true,
		CreatedAt:                   time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("create policy: %v", err)
	}

	if err := seed.DB.Create(&models.MemberContribution{
		ID:                  utils.GenerateUUID(),
		CommunityID:         seed.CommunityID,
		UserID:              seed.UserID,
		ValidPosts:          6,
		QualityComments:     0,
		PositiveReactions:   0,
		EventParticipations: 0,
		ContributionScore:   0,
		LastCalculatedAt:    time.Now().UTC(),
		CreatedAt:           time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("create contribution: %v", err)
	}

	result, err := seed.Service.RecalculateScore(ctx, seed.CommunityID, seed.UserID)
	if err != nil {
		t.Fatalf("RecalculateScore: %v", err)
	}
	// 6*10 = 60 >= threshold 50 → badge assigned.
	if result.BadgeType == nil || *result.BadgeType != "Top Contributor" {
		t.Errorf("BadgeType = %v, want %q", result.BadgeType, "Top Contributor")
	}
}

func TestRecalculateScore_BadgeDisabled(t *testing.T) {
	seed := newFullTestSeed(t)
	ctx := context.Background()

	// Policy with BadgeEnabled=false but high score.
	if err := seed.DB.Create(&models.CommunityPolicy{
		ID:          utils.GenerateUUID(),
		CommunityID: seed.CommunityID,
		PostWeight:  10,
		CommentWeight: 5,
		ReactionWeight: 2,
		EventWeight:   20,
		TopContributorThreshold:     50,
		ModeratorPromotionThreshold: 999999,
		AutoPromoteEnabled:          false,
		BadgeEnabled:                false,
		CreatedAt:                   time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("create policy: %v", err)
	}

	if err := seed.DB.Create(&models.MemberContribution{
		ID:                  utils.GenerateUUID(),
		CommunityID:         seed.CommunityID,
		UserID:              seed.UserID,
		ValidPosts:          6,
		QualityComments:     0,
		PositiveReactions:   0,
		EventParticipations: 0,
		ContributionScore:   0,
		LastCalculatedAt:    time.Now().UTC(),
		CreatedAt:           time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("create contribution: %v", err)
	}

	result, err := seed.Service.RecalculateScore(ctx, seed.CommunityID, seed.UserID)
	if err != nil {
		t.Fatalf("RecalculateScore: %v", err)
	}
	if result.BadgeType != nil {
		t.Errorf("BadgeType = %v, want nil when badge disabled", result.BadgeType)
	}
}
