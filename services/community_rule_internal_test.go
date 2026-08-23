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

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ─── Integration infrastructure ──────────────────────────────────

func connectAndMigrateCommunityRule(t *testing.T) *gorm.DB {
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
		&models.CommunityRule{},
		&models.NotificationPreference{},
	); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM community_rules")
		db.Exec("DELETE FROM group_members")
		db.Exec("DELETE FROM user_roles")
		db.Exec("DELETE FROM communities")
		db.Exec("DELETE FROM profiles")
		db.Exec("DELETE FROM notification_preferences")
		db.Exec("DELETE FROM users")
		db.Exec("DELETE FROM roles")
	})
	return db
}

type communityRuleTestSeed struct {
	DB          *gorm.DB
	RoleIDs     map[models.RoleName]string
	AdminID     string
	MemberID    string
	CommunityID string
	Service     *CommunityRuleService
}

func newCommunityRuleTestSeed(t *testing.T) communityRuleTestSeed {
	db := connectAndMigrateCommunityRule(t)
	roleIDs := seedCommunityRoles(t, db)

	adminID := seedUser(t, db)
	memberID := seedUser(t, db)

	// Create community
	communityID := utils.GenerateUUID()
	now := time.Now().UTC()
	if err := db.Create(&models.Community{
		ID:        communityID,
		CreatorID: adminID,
		Name:      "Rule Test Community " + communityID[:8],
		CreatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create community: %v", err)
	}

	// Add admin as group member + admin role
	scopeType := models.ScopeTypeCommunity
	if err := db.Create(&models.GroupMember{
		ID:          utils.GenerateUUID(),
		CommunityID: communityID,
		UserID:      adminID,
		JoinedAt:    now,
	}).Error; err != nil {
		t.Fatalf("create group member: %v", err)
	}
	if err := db.Create(&models.UserRole{
		ID:         utils.GenerateUUID(),
		UserID:     adminID,
		RoleID:     roleIDs[models.RoleGroupAdmin],
		ScopeID:    &communityID,
		ScopeType:  &scopeType,
		AssignedAt: now,
	}).Error; err != nil {
		t.Fatalf("assign group admin role: %v", err)
	}

	// Add member as group member (no admin role)
	if err := db.Create(&models.GroupMember{
		ID:          utils.GenerateUUID(),
		CommunityID: communityID,
		UserID:      memberID,
		JoinedAt:    now,
	}).Error; err != nil {
		t.Fatalf("create group member: %v", err)
	}
	if err := db.Create(&models.UserRole{
		ID:         utils.GenerateUUID(),
		UserID:     memberID,
		RoleID:     roleIDs[models.RoleGroupMember],
		ScopeID:    &communityID,
		ScopeType:  &scopeType,
		AssignedAt: now,
	}).Error; err != nil {
		t.Fatalf("assign group member role: %v", err)
	}

	communityRepo := repository.NewCommunityRepository(db)
	ruleRepo := repository.NewCommunityRuleRepository(db)
	ruleValidation := validations.NewCommunityRuleValidation()

	svc := NewCommunityRuleService(ruleRepo, communityRepo, ruleValidation)

	return communityRuleTestSeed{
		DB:          db,
		RoleIDs:     roleIDs,
		AdminID:     adminID,
		MemberID:    memberID,
		CommunityID: communityID,
		Service:     svc,
	}
}

// ─── CreateRule integration tests ───────────────────────────────

func TestCreateRule_HappyPath(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	rule, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Be respectful", "Treat everyone with respect", nil)
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}
	if rule.ID == "" {
		t.Fatal("rule.ID is empty")
	}
	if rule.CommunityID != seed.CommunityID {
		t.Errorf("rule.CommunityID = %q, want %q", rule.CommunityID, seed.CommunityID)
	}
	if rule.Category != models.RuleConduct {
		t.Errorf("rule.Category = %q, want %q", rule.Category, models.RuleConduct)
	}
	if rule.Title != "Be respectful" {
		t.Errorf("rule.Title = %q, want %q", rule.Title, "Be respectful")
	}
	if rule.Content != "Treat everyone with respect" {
		t.Errorf("rule.Content = %q, want %q", rule.Content, "Treat everyone with respect")
	}
	if rule.Position != 1 {
		t.Errorf("rule.Position = %d, want 1 (auto-increment)", rule.Position)
	}
}

func TestCreateRule_WithPosition(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	rule, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Rule with position", "Content", intPtr(5))
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}
	if rule.Position != 5 {
		t.Errorf("rule.Position = %d, want 5", rule.Position)
	}
}

func TestCreateRule_AutoPositionIncrement(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	r1, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "First rule", "", nil)
	if err != nil {
		t.Fatalf("CreateRule 1: %v", err)
	}
	if r1.Position != 1 {
		t.Errorf("r1.Position = %d, want 1", r1.Position)
	}

	r2, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Second rule", "", nil)
	if err != nil {
		t.Fatalf("CreateRule 2: %v", err)
	}
	if r2.Position != 2 {
		t.Errorf("r2.Position = %d, want 2", r2.Position)
	}

	r3, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Third rule", "", nil)
	if err != nil {
		t.Fatalf("CreateRule 3: %v", err)
	}
	if r3.Position != 3 {
		t.Errorf("r3.Position = %d, want 3", r3.Position)
	}
}

func TestCreateRule_DifferentCategories(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	r1, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Conduct rule", "", nil)
	if err != nil {
		t.Fatalf("CreateRule conduct: %v", err)
	}
	if r1.Position != 1 {
		t.Errorf("r1.Position = %d, want 1 (first in conduct)", r1.Position)
	}

	r2, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleProhibited, "Prohibited rule", "", nil)
	if err != nil {
		t.Fatalf("CreateRule prohibited: %v", err)
	}
	if r2.Position != 1 {
		t.Errorf("r2.Position = %d, want 1 (first in prohibited)", r2.Position)
	}
}

func TestCreateRule_NotAdmin(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.CreateRule(ctx, seed.MemberID, seed.CommunityID, models.RuleConduct, "Member rule", "", nil)
	if err == nil {
		t.Fatal("expected error for non-admin, got nil")
	}
}

func TestCreateRule_TitleDuplicate(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Duplicate Title", "", nil)
	if err != nil {
		t.Fatalf("first CreateRule: %v", err)
	}

	_, err = seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Duplicate Title", "", nil)
	if err == nil {
		t.Fatal("expected error for duplicate title, got nil")
	}
	if !errors.Is(err, validations.ErrRuleTitleDuplicate) {
		t.Errorf("error = %v, want ErrRuleTitleDuplicate", err)
	}
}

func TestCreateRule_DuplicateTitleDifferentCategory(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Same Title", "", nil)
	if err != nil {
		t.Fatalf("CreateRule conduct: %v", err)
	}

	// Same title in different category should succeed
	_, err = seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleProhibited, "Same Title", "", nil)
	if err != nil {
		t.Fatalf("CreateRule prohibited (different category): %v", err)
	}
}

func TestCreateRule_EmptyTitle(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "", "", nil)
	if err == nil {
		t.Fatal("expected error for empty title, got nil")
	}
}

func TestCreateRule_TitleTooShort(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "ab", "", nil)
	if err == nil {
		t.Fatal("expected error for short title, got nil")
	}
}

func TestCreateRule_InvalidCategory(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, "invalid", "Valid title", "", nil)
	if err == nil {
		t.Fatal("expected error for invalid category, got nil")
	}
}

func TestCreateRule_NegativePosition(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Valid title", "", intPtr(-1))
	if err == nil {
		t.Fatal("expected error for negative position, got nil")
	}
}

// ─── UpdateRule integration tests ───────────────────────────────

func TestUpdateRule_HappyPath(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	rule, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Original Title", "Original content", nil)
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	updated, err := seed.Service.UpdateRule(ctx, seed.AdminID, rule.ID, "Updated Title", "Updated content", nil, nil)
	if err != nil {
		t.Fatalf("UpdateRule: %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Errorf("Title = %q, want %q", updated.Title, "Updated Title")
	}
	if updated.Content != "Updated content" {
		t.Errorf("Content = %q, want %q", updated.Content, "Updated content")
	}
	if updated.UpdatedAt == nil {
		t.Fatal("UpdatedAt is nil, want non-nil")
	}
}

func TestUpdateRule_PartialUpdateTitleOnly(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	rule, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Original Title", "Original content", nil)
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	updated, err := seed.Service.UpdateRule(ctx, seed.AdminID, rule.ID, "New Title", "", nil, nil)
	if err != nil {
		t.Fatalf("UpdateRule: %v", err)
	}
	if updated.Title != "New Title" {
		t.Errorf("Title = %q, want %q", updated.Title, "New Title")
	}
	if updated.Content != "Original content" {
		t.Errorf("Content should remain unchanged, got %q", updated.Content)
	}
}

func TestUpdateRule_PartialUpdateContentOnly(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	rule, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Original Title", "Original content", nil)
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	updated, err := seed.Service.UpdateRule(ctx, seed.AdminID, rule.ID, "", "New content", nil, nil)
	if err != nil {
		t.Fatalf("UpdateRule: %v", err)
	}
	if updated.Title != "Original Title" {
		t.Errorf("Title should remain unchanged, got %q", updated.Title)
	}
	if updated.Content != "New content" {
		t.Errorf("Content = %q, want %q", updated.Content, "New content")
	}
}

func TestUpdateRule_ChangeCategory(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	rule, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Conduct Rule", "", nil)
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	cat := models.RuleProhibited
	updated, err := seed.Service.UpdateRule(ctx, seed.AdminID, rule.ID, "", "", &cat, nil)
	if err != nil {
		t.Fatalf("UpdateRule: %v", err)
	}
	if updated.Category != models.RuleProhibited {
		t.Errorf("Category = %q, want %q", updated.Category, models.RuleProhibited)
	}
}

func TestUpdateRule_ChangePosition(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	rule, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Position Rule", "", intPtr(1))
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	pos := 10
	updated, err := seed.Service.UpdateRule(ctx, seed.AdminID, rule.ID, "", "", nil, &pos)
	if err != nil {
		t.Fatalf("UpdateRule: %v", err)
	}
	if updated.Position != 10 {
		t.Errorf("Position = %d, want 10", updated.Position)
	}
}

func TestUpdateRule_NotFound(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.UpdateRule(ctx, seed.AdminID, "non-existent-rule-id", "New Title", "", nil, nil)
	if err == nil {
		t.Fatal("expected error for non-existent rule, got nil")
	}
}

func TestUpdateRule_NotAdmin(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	rule, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Admin Rule", "", nil)
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	_, err = seed.Service.UpdateRule(ctx, seed.MemberID, rule.ID, "Hacked Title", "", nil, nil)
	if err == nil {
		t.Fatal("expected error for non-admin, got nil")
	}
}

func TestUpdateRule_TitleTooShort(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	rule, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Valid Title", "", nil)
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	_, err = seed.Service.UpdateRule(ctx, seed.AdminID, rule.ID, "ab", "", nil, nil)
	if err == nil {
		t.Fatal("expected error for short title, got nil")
	}
}

func TestUpdateRule_NegativePosition(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	rule, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Valid Title", "", nil)
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	pos := -1
	_, err = seed.Service.UpdateRule(ctx, seed.AdminID, rule.ID, "", "", nil, &pos)
	if err == nil {
		t.Fatal("expected error for negative position, got nil")
	}
	if !errors.Is(err, validations.ErrRulePositionNegative) {
		t.Errorf("error = %v, want ErrRulePositionNegative", err)
	}
}

// ─── DeleteRule integration tests ───────────────────────────────

func TestDeleteRule_HappyPath(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	rule, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Delete Me", "", nil)
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	if err := seed.Service.DeleteRule(ctx, seed.AdminID, rule.ID); err != nil {
		t.Fatalf("DeleteRule: %v", err)
	}

	// Verify rule was deleted
	rules, err := seed.Service.GetRulesByCommunity(ctx, seed.CommunityID)
	if err != nil {
		t.Fatalf("GetRulesByCommunity: %v", err)
	}
	if len(rules) != 0 {
		t.Errorf("rules count = %d, want 0", len(rules))
	}
}

func TestDeleteRule_NotFound(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	err := seed.Service.DeleteRule(ctx, seed.AdminID, "non-existent-rule-id")
	if err == nil {
		t.Fatal("expected error for non-existent rule, got nil")
	}
}

func TestDeleteRule_NotAdmin(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	rule, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Admin Rule", "", nil)
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	err = seed.Service.DeleteRule(ctx, seed.MemberID, rule.ID)
	if err == nil {
		t.Fatal("expected error for non-admin, got nil")
	}
}

// ─── GetRulesByCommunity integration tests ─────────────────────

func TestGetRulesByCommunity_HasRules(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Conduct Rule 1", "", nil)
	if err != nil {
		t.Fatalf("CreateRule 1: %v", err)
	}
	_, err = seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleConduct, "Conduct Rule 2", "", nil)
	if err != nil {
		t.Fatalf("CreateRule 2: %v", err)
	}
	_, err = seed.Service.CreateRule(ctx, seed.AdminID, seed.CommunityID, models.RuleProhibited, "Prohibited Rule 1", "", nil)
	if err != nil {
		t.Fatalf("CreateRule 3: %v", err)
	}

	rules, err := seed.Service.GetRulesByCommunity(ctx, seed.CommunityID)
	if err != nil {
		t.Fatalf("GetRulesByCommunity: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("got %d rules, want 3", len(rules))
	}

	// Rules should be sorted by category, then position
	if rules[0].Category != models.RuleConduct {
		t.Errorf("rules[0].Category = %q, want %q", rules[0].Category, models.RuleConduct)
	}
	if rules[0].Position != 1 {
		t.Errorf("rules[0].Position = %d, want 1", rules[0].Position)
	}
	if rules[1].Position != 2 {
		t.Errorf("rules[1].Position = %d, want 2", rules[1].Position)
	}
	if rules[2].Category != models.RuleProhibited {
		t.Errorf("rules[2].Category = %q, want %q", rules[2].Category, models.RuleProhibited)
	}
}

func TestGetRulesByCommunity_Empty(t *testing.T) {
	seed := newCommunityRuleTestSeed(t)
	ctx := context.Background()

	rules, err := seed.Service.GetRulesByCommunity(ctx, seed.CommunityID)
	if err != nil {
		t.Fatalf("GetRulesByCommunity: %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("got %d rules, want 0", len(rules))
	}
}

// ─── Helper ─────────────────────────────────────────────────────

func intPtr(v int) *int {
	return &v
}
