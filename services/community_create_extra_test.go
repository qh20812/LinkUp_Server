package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
)

// unwrapCreateResult adapts CreateCommunityResult back to the tuple shape
// used across community_internal_test.go call sites.
func unwrapCreateResult(res *CreateCommunityResult, err error) (*models.Community, *models.Chat, error) {
	if err != nil || res == nil {
		return nil, nil, err
	}
	return res.Community, res.Chat, nil
}

// ─── No-DB unit ─────────────────────────────────────────────────────

func TestIsDuplicateEntryError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"mysql 1062", fmt.Errorf("lỗi khi lưu thông tin cộng đồng: Error 1062 (23000): Duplicate entry 'x' for key 'communities.name'"), true},
		{"wrapped duplicate key", fmt.Errorf("outer: %w", errors.New("duplicate key value violates unique constraint")), true},
		{"other db error", errors.New("connection refused"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDuplicateEntryError(tt.err); got != tt.want {
				t.Errorf("isDuplicateEntryError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// ─── DB integration (skip without TEST_DSN) ─────────────────────────

func TestCreateCommunity_QuotaTotalLimit(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	// Seed 10 old communities directly (48h ago → dodges the 24h throttle).
	old := time.Now().UTC().Add(-48 * time.Hour)
	for i := 0; i < MaxCommunitiesPerUser; i++ {
		id := utils.GenerateUUID()
		if err := seed.DB.Create(&models.Community{
			ID:        id,
			CreatorID: seed.CreatorID,
			Name:      fmt.Sprintf("Quota Filler %d %s", i, id[:8]),
			CreatedAt: old,
		}).Error; err != nil {
			t.Fatalf("seed community %d: %v", i, err)
		}
	}

	_, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "One Too Many", "", "", "", models.PrivacyPublic, false)
	if err == nil {
		t.Fatal("expected quota error, got nil")
	}
	if !errors.Is(err, validations.ErrCommunityCreateLimited) {
		t.Errorf("error = %v, want ErrCommunityCreateLimited", err)
	}
}

func TestCreateCommunity_Throttle(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	for i := 0; i < MaxCommunitiesPerDay; i++ {
		name := fmt.Sprintf("Rapid %d %s", i, utils.GenerateUUID()[:8])
		if _, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, name, "", "", "", models.PrivacyPublic, false); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	_, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Too Fast", "", "", "", models.PrivacyPublic, false)
	if err == nil {
		t.Fatal("expected throttle error, got nil")
	}
	if !errors.Is(err, validations.ErrCommunityCreateThrottled) {
		t.Errorf("error = %v, want ErrCommunityCreateThrottled", err)
	}
}

func TestCreateCommunity_TrimsName(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	res, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "  Spaced Name  ", "  Padded desc  ", "", "", models.PrivacyPublic, false)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}
	community := res.Community
	if community.Name != "Spaced Name" {
		t.Errorf("stored name = %q, want trimmed", community.Name)
	}
	if community.Description != "Padded desc" {
		t.Errorf("stored description = %q, want trimmed", community.Description)
	}
	taken, err := seed.Service.repo.IsNameTaken(ctx, "Spaced Name")
	if err != nil || !taken {
		t.Errorf("IsNameTaken(trimmed) = %v, %v; want true, nil", taken, err)
	}
}

func TestCreateCommunity_RejectsBadPrivacy(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	_, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Bad Privacy", "", "", "", models.CommunityPrivacy("publik"), false)
	if err == nil {
		t.Fatal("expected privacy error, got nil")
	}
	if !errors.Is(err, validations.ErrCommunityPrivacyInvalid) {
		t.Errorf("error = %v, want ErrCommunityPrivacyInvalid", err)
	}
}

func TestCreateCommunity_SeedsDefaults(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	res, err := seed.Service.CreateCommunity(ctx, seed.CreatorID, "Seeded Community", "", "", "", models.PrivacyPublic, false)
	if err != nil {
		t.Fatalf("CreateCommunity: %v", err)
	}
	community := res.Community

	var ruleCount int64
	if err := seed.DB.Model(&models.CommunityRule{}).Where("community_id = ?", community.ID).Count(&ruleCount).Error; err != nil {
		t.Fatalf("count rules: %v", err)
	}
	if ruleCount != 3 {
		t.Errorf("rules = %d, want 3 defaults", ruleCount)
	}

	var codes []models.CommunityInviteCode
	if err := seed.DB.Where("community_id = ?", community.ID).Find(&codes).Error; err != nil {
		t.Fatalf("list codes: %v", err)
	}
	if len(codes) != 1 || !codes[0].IsActive || codes[0].MaxUses != 0 || codes[0].ExpiresAt != nil {
		t.Fatalf("starter invite code = %+v, want 1 active unlimited code", codes)
	}
	if res.InviteCode == "" {
		t.Error("result InviteCode empty, want starter code for response payload")
	} else if res.InviteCode != codes[0].Code {
		t.Errorf("result InviteCode = %q, want %q", res.InviteCode, codes[0].Code)
	}
}

func TestCreateCommunity_DuplicateRace(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()
	name := "Race Community " + utils.GenerateUUID()[:8]

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = seed.Service.CreateCommunity(ctx, seed.CreatorID, name, "", "", "", models.PrivacyPublic, false)
		}(i)
	}
	wg.Wait()

	var success, conflict int
	for _, err := range errs {
		switch {
		case err == nil:
			success++
		case errors.Is(err, validations.ErrCommunityNameExists):
			conflict++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Errorf("success=%d conflict=%d, want exactly 1 success + 1 NAME_EXISTS", success, conflict)
	}
}

func TestAuthRepository_HasAnyRole(t *testing.T) {
	seed := newCreateCommunityTestSeed(t)
	ctx := context.Background()

	now := time.Now().UTC()
	adminRoleID, ok := seed.RoleIDs[models.RoleAdmin]
	if !ok {
		t.Fatal("admin role not seeded")
	}
	if err := seed.DB.Create(&models.UserRole{
		ID:         utils.GenerateUUID(),
		UserID:     seed.MemberID,
		RoleID:     adminRoleID,
		AssignedAt: now,
	}).Error; err != nil {
		t.Fatalf("grant admin: %v", err)
	}

	authRepo := repository.NewAuthRepository(seed.DB)
	got, err := authRepo.HasAnyRole(ctx, seed.MemberID, models.RoleAdmin, models.RoleSuperAdmin)
	if err != nil || !got {
		t.Errorf("HasAnyRole(admin) = %v, %v; want true, nil", got, err)
	}
	got, err = authRepo.HasAnyRole(ctx, seed.CreatorID, models.RoleAdmin, models.RoleSuperAdmin)
	if err != nil || got {
		t.Errorf("HasAnyRole(plain user) = %v, %v; want false, nil", got, err)
	}
	got, err = authRepo.HasAnyRole(ctx, seed.MemberID)
	if err != nil || got {
		t.Errorf("HasAnyRole(no roles) = %v, %v; want false, nil", got, err)
	}
}
