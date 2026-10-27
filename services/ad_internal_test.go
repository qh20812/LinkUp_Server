package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gin-gonic/gin"

	"linkup/dto"
	"linkup/models"
	"linkup/repository"
)

// ─── fakes ────────────────────────────────────────────────────────────────

type fakePackageRepo struct {
	repository.PackageRepository
	sub       *models.PartnerSubscription
	err       error
	slotsUsed int
	lastSlots string
}

func (f *fakePackageRepo) GetActiveSubscription(_ string) (*models.PartnerSubscription, error) {
	return f.sub, f.err
}

func (f *fakePackageRepo) IncrementSlotsUsed(id string) error {
	f.slotsUsed++
	f.lastSlots = id
	return nil
}

func (f *fakePackageRepo) DecrementSlotsUsed(id string) error {
	f.slotsUsed++
	f.lastSlots = id
	return nil
}

type fakeAdRepo struct {
	repository.AdRepository
	created   *models.Ad
	createErr error

	existing   *models.Ad
	findErr    error
	isOwner    bool
	ownerErr   error
	deleted    string
	deleteErr  error
	mediaPurge string
}

func (f *fakeAdRepo) Create(ad *models.Ad) error {
	f.created = ad
	return f.createErr
}

func (f *fakeAdRepo) FindByID(_ string) (*models.Ad, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.existing, nil
}

func (f *fakeAdRepo) CheckAdOwnership(_, _ string) (bool, error) {
	return f.isOwner, f.ownerErr
}

func (f *fakeAdRepo) Delete(_ context.Context, id string) error {
	f.deleted = id
	return f.deleteErr
}

func (f *fakeAdRepo) DeleteAdMedia(id string) error {
	f.mediaPurge = id
	return nil
}

type fakeAdMediaService struct {
	MediaService
}

func newActiveSubscription() *models.PartnerSubscription {
	return &models.PartnerSubscription{
		ID:        "sub-1",
		PackageID: "pkg-1",
		SlotsUsed: 0,
		Package: &models.AdPackage{
			MaxSlots:        8,
			SupportsVideo:   true,
			SupportsCarousel: true,
		},
	}
}

func newTestAdService(repo *fakeAdRepo, pkg *fakePackageRepo) AdService {
	return NewAdService(repo, pkg, &fakeAdMediaService{})
}

func validCreateInput() dto.CreateAdInput {
	return dto.CreateAdInput{
		Title:      "Khuyến mãi",
		Content:    "Nội dung",
		TargetURL:  "https://linkup.vn",
		Format:     "image",
		Budget:     100000,
		TargetGender: "all",
	}
}

// ─── targetLocationsJSON ──────────────────────────────────────────────────

func TestTargetLocationsJSON(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want string
	}{
		{name: "nil thành mảng rỗng", in: nil, want: "[]"},
		{name: "mảng rỗng giữ nguyên", in: []string{}, want: "[]"},
		{name: "danh sách địa điểm", in: []string{"Hà Nội", "TP.HCM"}, want: `["Hà Nội","TP.HCM"]`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := targetLocationsJSON(tc.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, got)
			}
			// Phải parse lại được — MySQL từ chối mọi chuỗi không phải JSON.
			var back []string
			if err := json.Unmarshal([]byte(got), &back); err != nil {
				t.Fatalf("kết quả không phải JSON hợp lệ: %v", err)
			}
		})
	}
}

// ─── CreateAdWithMedia ────────────────────────────────────────────────────

func TestCreateAdWithMedia_WritesValidJSONWhenNoLocations(t *testing.T) {
	// Hồi quy: cột target_locations kiểu JSON nên chuỗi rỗng làm MySQL trả
	// Error 3140 → tạo quảng cáo không chọn địa điểm luôn thất bại.
	repo := &fakeAdRepo{}
	pkg := &fakePackageRepo{sub: newActiveSubscription()}
	svc := newTestAdService(repo, pkg)

	ctx, _ := gin.CreateTestContext(nil)
	ad, err := svc.CreateAdWithMedia(ctx, validCreateInput(), nil, "partner-1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ad.TargetLocations != "[]" {
		t.Fatalf("expected target_locations [], got %q", ad.TargetLocations)
	}
	if repo.created == nil {
		t.Fatalf("ad chưa được ghi xuống repo")
	}
	if pkg.slotsUsed != 1 {
		t.Fatalf("expected 1 slot increment, got %d", pkg.slotsUsed)
	}
}

func TestCreateAdWithMedia_SerialisesLocations(t *testing.T) {
	repo := &fakeAdRepo{}
	svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

	input := validCreateInput()
	input.TargetLocations = []string{"Đà Nẵng"}

	ctx, _ := gin.CreateTestContext(nil)
	ad, err := svc.CreateAdWithMedia(ctx, input, nil, "partner-1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ad.TargetLocations != `["Đà Nẵng"]` {
		t.Fatalf("unexpected target_locations: %q", ad.TargetLocations)
	}
}

func TestCreateAdWithMedia_NoSubscription(t *testing.T) {
	repo := &fakeAdRepo{}
	svc := newTestAdService(repo, &fakePackageRepo{sub: nil})

	ctx, _ := gin.CreateTestContext(nil)
	if _, err := svc.CreateAdWithMedia(ctx, validCreateInput(), nil, "partner-1", false); err == nil {
		t.Fatalf("expected error khi chưa có gói")
	}
	if repo.created != nil {
		t.Fatalf("không được ghi ad khi chưa có gói")
	}
}

func TestCreateAdWithMedia_SlotsExhausted(t *testing.T) {
	sub := newActiveSubscription()
	sub.SlotsUsed = sub.Package.MaxSlots
	repo := &fakeAdRepo{}
	svc := newTestAdService(repo, &fakePackageRepo{sub: sub})

	ctx, _ := gin.CreateTestContext(nil)
	if _, err := svc.CreateAdWithMedia(ctx, validCreateInput(), nil, "partner-1", false); err == nil {
		t.Fatalf("expected error khi hết slot")
	}
	if repo.created != nil {
		t.Fatalf("không được ghi ad khi hết slot")
	}
}

func TestCreateAdWithMedia_PlatformAdminNeedsNoSubscription(t *testing.T) {
	// ADMIN/SUPER_ADMIN tạo quảng cáo hộ nên không có gói: phải tạo được,
	// không gắn PackageID và không đụng tới slot.
	repo := &fakeAdRepo{}
	pkg := &fakePackageRepo{sub: nil}
	svc := newTestAdService(repo, pkg)

	ctx, _ := gin.CreateTestContext(nil)
	ad, err := svc.CreateAdWithMedia(ctx, validCreateInput(), nil, "admin-1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ad.PackageID != nil {
		t.Fatalf("ad của admin không được gắn gói")
	}
	if pkg.slotsUsed != 0 {
		t.Fatalf("không được tăng slot cho admin")
	}
	if ad.TargetLocations != "[]" {
		t.Fatalf("expected target_locations [], got %q", ad.TargetLocations)
	}
}

func TestCreateAdWithMedia_VideoNotSupportedByPackage(t *testing.T) {
	sub := newActiveSubscription()
	sub.Package.SupportsVideo = false
	repo := &fakeAdRepo{}
	svc := newTestAdService(repo, &fakePackageRepo{sub: sub})

	input := validCreateInput()
	input.Format = "video"

	ctx, _ := gin.CreateTestContext(nil)
	if _, err := svc.CreateAdWithMedia(ctx, input, nil, "partner-1", false); err == nil {
		t.Fatalf("expected error khi gói không hỗ trợ video")
	}
}

func TestCreateAdWithMedia_RepoErrorPropagates(t *testing.T) {
	repo := &fakeAdRepo{createErr: errors.New("db down")}
	pkg := &fakePackageRepo{sub: newActiveSubscription()}
	svc := newTestAdService(repo, pkg)

	ctx, _ := gin.CreateTestContext(nil)
	if _, err := svc.CreateAdWithMedia(ctx, validCreateInput(), nil, "partner-1", false); err == nil {
		t.Fatalf("expected repo error to propagate")
	}
	if pkg.slotsUsed != 0 {
		t.Fatalf("không được tăng slot khi ghi ad thất bại")
	}
}

// ─── DeleteAd ─────────────────────────────────────────────────────────────

func TestDeleteAd_AllowsActiveAdAndFreesSlot(t *testing.T) {
	// Hồi quy: quảng cáo tạo ra là active ngay, nên chặn xoá active khiến đối
	// tác không thể xoá quảng cáo vừa tạo và slot bị kẹt.
	repo := &fakeAdRepo{
		isOwner:  true,
		existing: &models.Ad{ID: "ad-1", Status: models.AdStatusActive, PackageID: ptrString("pkg-1")},
	}
	pkg := &fakePackageRepo{sub: newActiveSubscription()}
	svc := newTestAdService(repo, pkg)

	if err := svc.DeleteAd("ad-1", "partner-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.deleted != "ad-1" {
		t.Fatalf("ad chưa bị xoá")
	}
	if repo.mediaPurge != "ad-1" {
		t.Fatalf("media chưa bị xoá")
	}
	if pkg.slotsUsed != 1 {
		t.Fatalf("expected slot được giải phóng, got %d", pkg.slotsUsed)
	}
	// Phải giải phóng theo ID subscription, không phải ID gói — nếu không câu
	// UPDATE không khớp dòng nào và slot không bao giờ được trả lại.
	if pkg.lastSlots != "sub-1" {
		t.Fatalf("expected decrement theo subscription sub-1, got %q", pkg.lastSlots)
	}
}

func TestDeleteAd_RejectsNonOwner(t *testing.T) {
	repo := &fakeAdRepo{isOwner: false, existing: &models.Ad{ID: "ad-1", Status: models.AdStatusPaused}}
	svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

	if err := svc.DeleteAd("ad-1", "partner-2"); err == nil {
		t.Fatalf("expected error khi không phải chủ sở hữu")
	}
	if repo.deleted != "" {
		t.Fatalf("không được xoá ad của người khác")
	}
}

func TestDeleteAd_SkipsSlotWhenNoActiveSubscription(t *testing.T) {
	repo := &fakeAdRepo{isOwner: true, existing: &models.Ad{ID: "ad-1", Status: models.AdStatusPaused}}
	pkg := &fakePackageRepo{sub: nil}
	svc := newTestAdService(repo, pkg)

	if err := svc.DeleteAd("ad-1", "partner-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.slotsUsed != 0 {
		t.Fatalf("không được giải phóng slot khi không có gói đang hoạt động")
	}
}

func ptrString(v string) *string { return &v }
