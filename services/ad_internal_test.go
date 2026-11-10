package services

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
)

// ─── fakes ────────────────────────────────────────────────────────────────

type fakePackageRepo struct {
	repository.PackageRepository
	sub *models.PartnerSubscription
	err error

	reserveFail bool
	slotsUsed   int
	released    int
	lastSlots   string
}

func (f *fakePackageRepo) GetActiveSubscription(_ string) (*models.PartnerSubscription, error) {
	return f.sub, f.err
}

func (f *fakePackageRepo) ReserveSlot(id string) (bool, error) {
	if f.reserveFail {
		return false, nil
	}
	f.slotsUsed++
	f.lastSlots = id
	return true, nil
}

func (f *fakePackageRepo) DecrementSlotsUsed(id string) error {
	if f.slotsUsed > 0 {
		f.slotsUsed--
	}
	f.released++
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

	updated       *models.Ad
	updateErr     error
	updatedFields map[string]interface{}

	mediaBatch []models.AdMedia
	mediaErr   error
	adMedia    []models.AdMedia
	replaced   []models.AdMedia

	actionCounts map[string]int64
	countsErr    error

	partnerKeyword, partnerStatus string
	partnerLimit, partnerOffset   int
	partnerTotal                  int64
	partnerItems                  []dto.PartnerAdListItem

	adminKeyword, adminStatus string
	adminLimit, adminOffset   int
	adminTotal                int64
	adminItems                []dto.AdminAdListItem

	tracked  *models.AdAnalytics
	trackErr error

	completedRows int64
	completedAt   time.Time
	completedErr  error

	rangeMetrics     *repository.AdRangeMetrics
	rangeErr         error
	rangeFrom, rangeTo time.Time
}

func (f *fakeAdRepo) GetAdMetricsRange(_ string, from, to time.Time) (*repository.AdRangeMetrics, error) {
	f.rangeFrom, f.rangeTo = from, to
	if f.rangeErr != nil {
		return nil, f.rangeErr
	}
	if f.rangeMetrics == nil {
		return &repository.AdRangeMetrics{Counts: map[string]int64{}}, nil
	}
	return f.rangeMetrics, nil
}

func (f *fakeAdRepo) MarkFinishedAdsCompleted(now time.Time) (int64, error) {
	f.completedAt = now
	return f.completedRows, f.completedErr
}

func (f *fakeAdRepo) TrackActionAndDeduct(analytics *models.AdAnalytics) (bool, error) {
	f.tracked = analytics
	if f.trackErr != nil {
		return false, f.trackErr
	}
	return true, nil
}

// ─── Bộ lọc danh sách (P0-3) ──────────────────────────────────────────────

func (f *fakeAdRepo) ListByPartnerID(_, keyword, status string, limit, offset int) ([]dto.PartnerAdListItem, error) {
	f.partnerKeyword, f.partnerStatus = keyword, status
	f.partnerLimit, f.partnerOffset = limit, offset
	return f.partnerItems, nil
}

func (f *fakeAdRepo) CountByPartnerID(_, keyword, status string) (int64, error) {
	f.partnerKeyword, f.partnerStatus = keyword, status
	return f.partnerTotal, nil
}

func (f *fakeAdRepo) ListAds(_ context.Context, keyword, status string, limit, offset int) ([]dto.AdminAdListItem, error) {
	f.adminKeyword, f.adminStatus = keyword, status
	f.adminLimit, f.adminOffset = limit, offset
	return f.adminItems, nil
}

func (f *fakeAdRepo) CountAds(_ context.Context, keyword, status string) (int64, error) {
	f.adminKeyword, f.adminStatus = keyword, status
	return f.adminTotal, nil
}

func (f *fakeAdRepo) GetActionCounts(string) (map[string]int64, error) {
	if f.countsErr != nil {
		return nil, f.countsErr
	}
	if f.actionCounts == nil {
		return map[string]int64{}, nil
	}
	return f.actionCounts, nil
}

func (f *fakeAdRepo) GetUniqueReach(string) (int64, error) { return 0, nil }

func (f *fakeAdRepo) Update(ad *models.Ad) error {
	f.updated = ad
	return f.updateErr
}

func (f *fakeAdRepo) UpdateAdFields(ad *models.Ad, fields map[string]interface{}) error {
	f.updated = ad
	f.updatedFields = fields
	if f.updateErr != nil {
		return f.updateErr
	}
	// Mô phỏng UPDATE thật để test đọc lại thấy trạng thái mới.
	if f.existing != nil {
		if v, ok := fields["status"].(string); ok {
			f.existing.Status = models.ParseAdStatus(v)
		}
		if v, ok := fields["rejection_reason"]; ok && v == nil {
			f.existing.RejectionReason = nil
		}
	}
	return nil
}

func (f *fakeAdRepo) Create(ad *models.Ad) error {
	f.created = ad
	return f.createErr
}

func (f *fakeAdRepo) CreateMediaBatch(items []models.AdMedia) error {
	f.mediaBatch = items
	return f.mediaErr
}

func (f *fakeAdRepo) ListAdMedia(_ string) ([]models.AdMedia, error) {
	return f.adMedia, nil
}

func (f *fakeAdRepo) ReplaceAdMedia(_ string, items []models.AdMedia) error {
	f.replaced = items
	return f.mediaErr
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

// uploadFakeMediaService cho phép test luồng upload ad (thành công/lỗi) mà
// không cần Cloudinary.
type uploadFakeMediaService struct {
	MediaService
	upload   func(ctx context.Context, userID string, file *multipart.FileHeader) (*models.Media, error)
	duration func(ctx context.Context, publicID string) (float64, error)
	deleted  []string
	cleaned  []string
}

func (f *uploadFakeMediaService) AutoApproveUpload(ctx context.Context, userID string, file *multipart.FileHeader) (*models.Media, error) {
	return f.upload(ctx, userID, file)
}

func (f *uploadFakeMediaService) GetVideoDurationSeconds(ctx context.Context, publicID string) (float64, error) {
	if f.duration == nil {
		return 0, errors.New("không có metadata")
	}
	return f.duration(ctx, publicID)
}

func (f *uploadFakeMediaService) DeleteMedia(_ context.Context, _ string, mediaID string) error {
	f.deleted = append(f.deleted, mediaID)
	return nil
}

func (f *uploadFakeMediaService) DeleteByURL(_ context.Context, _ string, fileURI string) error {
	f.cleaned = append(f.cleaned, fileURI)
	return nil
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
	repo := &fakeAdRepo{}
	pkg := &fakePackageRepo{sub: newActiveSubscription(), reserveFail: true}
	svc := newTestAdService(repo, pkg)

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
	if pkg.released != 1 {
		t.Fatalf("phải trả slot đã giữ khi ghi ad thất bại, got released=%d", pkg.released)
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
	if pkg.released != 1 {
		t.Fatalf("expected slot được giải phóng, got released=%d", pkg.released)
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

func TestDeleteAd_CompletedAdDoesNotFreeSlotAgain(t *testing.T) {
	// Ad completed đã trả slot lúc chuyển trạng thái; xoá không được trừ thêm
	// (nếu không sẽ ăn slot của một ad khác đang chạy).
	repo := &fakeAdRepo{
		isOwner:  true,
		existing: &models.Ad{ID: "ad-1", Status: models.AdStatusCompleted, PackageID: ptrString("pkg-1")},
	}
	pkg := &fakePackageRepo{sub: newActiveSubscription()}
	pkg.slotsUsed = 2
	svc := newTestAdService(repo, pkg)

	if err := svc.DeleteAd("ad-1", "partner-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.slotsUsed != 2 || pkg.released != 0 {
		t.Fatalf("ad completed không được trả slot lần nữa, got slotsUsed=%d released=%d", pkg.slotsUsed, pkg.released)
	}
}

func TestDeleteAd_SuspendedAdDoesNotFreeSlotAgain(t *testing.T) {
	repo := &fakeAdRepo{
		isOwner:  true,
		existing: &models.Ad{ID: "ad-1", Status: models.AdStatusSuspended, PackageID: ptrString("pkg-1")},
	}
	pkg := &fakePackageRepo{sub: newActiveSubscription()}
	pkg.slotsUsed = 1
	svc := newTestAdService(repo, pkg)

	if err := svc.DeleteAd("ad-1", "partner-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.slotsUsed != 1 || pkg.released != 0 {
		t.Fatalf("ad suspended không được trả slot lần nữa, got slotsUsed=%d released=%d", pkg.slotsUsed, pkg.released)
	}
}

// ─── P0-1: state machine duyệt quảng cáo ──────────────────────────────────

func TestCreateAdWithMedia_PartnerAdStartsPending(t *testing.T) {
	repo := &fakeAdRepo{}
	svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

	ctx, _ := gin.CreateTestContext(nil)
	ad, err := svc.CreateAdWithMedia(ctx, validCreateInput(), nil, "partner-1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ad.Status != models.AdStatusPending {
		t.Fatalf("ad của partner phải chờ duyệt, got %q", ad.Status)
	}
}

func TestCreateAdWithMedia_PlatformAdminAdStartsActive(t *testing.T) {
	repo := &fakeAdRepo{}
	svc := newTestAdService(repo, &fakePackageRepo{sub: nil})

	ctx, _ := gin.CreateTestContext(nil)
	ad, err := svc.CreateAdWithMedia(ctx, validCreateInput(), nil, "admin-1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ad.Status != models.AdStatusActive {
		t.Fatalf("ad do admin tạo phải active ngay, got %q", ad.Status)
	}
}

func TestUpdateAd_AllowsPendingAndRejected(t *testing.T) {
	for _, status := range []models.AdStatus{models.AdStatusPending, models.AdStatusRejected} {
		repo := &fakeAdRepo{
			isOwner:  true,
			existing: &models.Ad{ID: "ad-1", Status: status},
		}
		svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

		if _, err := svc.UpdateAd("ad-1", "partner-1", dto.UpdateAdInput{Title: "Tiêu đề mới"}); err != nil {
			t.Fatalf("status %s phải được sửa: %v", status, err)
		}
	}
}

func TestUpdateAd_BlocksSuspendedAndCompleted(t *testing.T) {
	for _, status := range []models.AdStatus{models.AdStatusSuspended, models.AdStatusCompleted} {
		repo := &fakeAdRepo{
			isOwner:  true,
			existing: &models.Ad{ID: "ad-1", Status: status},
		}
		svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

		if _, err := svc.UpdateAd("ad-1", "partner-1", dto.UpdateAdInput{Title: "x"}); err == nil {
			t.Fatalf("status %s không được sửa", status)
		}
	}
}

func TestSubmitForReview_ClearsRejectionAndBackToPending(t *testing.T) {
	reason := "nội dung sai"
	repo := &fakeAdRepo{
		isOwner:  true,
		existing: &models.Ad{ID: "ad-1", Status: models.AdStatusRejected, RejectionReason: &reason},
	}
	svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

	ad, err := svc.SubmitForReview("ad-1", "partner-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ad.Status != models.AdStatusPending {
		t.Fatalf("expected pending, got %q", ad.Status)
	}
	if ad.RejectionReason != nil {
		t.Fatalf("rejection_reason phải được xoá")
	}
}

func TestSubmitForReview_RejectsNonRejectedAd(t *testing.T) {
	repo := &fakeAdRepo{
		isOwner:  true,
		existing: &models.Ad{ID: "ad-1", Status: models.AdStatusActive},
	}
	svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

	if _, err := svc.SubmitForReview("ad-1", "partner-1"); err == nil {
		t.Fatalf("chỉ ad bị từ chối mới được gửi duyệt lại")
	}
}

func TestUpdateStatus_PartnerCanOnlyTogglePause(t *testing.T) {
	// active -> paused hợp lệ
	repo := &fakeAdRepo{
		isOwner:  true,
		existing: &models.Ad{ID: "ad-1", Status: models.AdStatusActive, Budget: 100},
	}
	svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})
	ad, err := svc.UpdateStatus("ad-1", "paused", "partner-1")
	if err != nil {
		t.Fatalf("active -> paused phải được phép: %v", err)
	}
	if ad.Status != models.AdStatusPaused {
		t.Fatalf("expected paused, got %q", ad.Status)
	}

	// active -> completed bị chặn (chỉ admin/hệ thống được set)
	repo = &fakeAdRepo{
		isOwner:  true,
		existing: &models.Ad{ID: "ad-2", Status: models.AdStatusActive, Budget: 100},
	}
	svc = newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})
	if _, err := svc.UpdateStatus("ad-2", "completed", "partner-1"); err == nil {
		t.Fatalf("partner không được tự đánh dấu completed")
	}

	// pending -> active bị chặn (partner không được tự duyệt)
	repo = &fakeAdRepo{
		isOwner:  true,
		existing: &models.Ad{ID: "ad-3", Status: models.AdStatusPending, Budget: 100},
	}
	svc = newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})
	if _, err := svc.UpdateStatus("ad-3", "active", "partner-1"); err == nil {
		t.Fatalf("partner không được tự duyệt quảng cáo")
	}
}

func TestUpdateStatus_ResumeNeedsActiveSubscription(t *testing.T) {
	repo := &fakeAdRepo{
		isOwner:  true,
		existing: &models.Ad{ID: "ad-1", Status: models.AdStatusPaused, Budget: 100},
	}
	svc := newTestAdService(repo, &fakePackageRepo{sub: nil})

	if _, err := svc.UpdateStatus("ad-1", "active", "partner-1"); err == nil {
		t.Fatalf("không có gói đang hoạt động thì không được bật lại")
	}
}

func TestUpdateStatus_ResumeRejectedWhenBudgetSpent(t *testing.T) {
	repo := &fakeAdRepo{
		isOwner: true,
		existing: &models.Ad{
			ID:         "ad-1",
			Status:     models.AdStatusPaused,
			Budget:     100,
			TotalSpent: 100,
		},
	}
	svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

	if _, err := svc.UpdateStatus("ad-1", "active", "partner-1"); err == nil {
		t.Fatalf("hết ngân sách thì không được bật lại")
	}
}

func TestGetAdPerformance_AdminBypassesOwnership(t *testing.T) {
	repo := &fakeAdRepo{
		isOwner:      false,
		existing:     &models.Ad{ID: "ad-1", Status: models.AdStatusActive, Budget: 100, TotalSpent: 10},
		actionCounts: map[string]int64{"impression": 1000, "click": 20},
	}
	svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

	report, err := svc.GetAdPerformance("ad-1", "admin-1", true, nil, nil)
	if err != nil {
		t.Fatalf("admin phải xem được ad của partner: %v", err)
	}
	if report.Impressions != 1000 || report.Clicks != 20 {
		t.Fatalf("số liệu không đúng: %+v", report)
	}
}

func TestGetAdPerformance_PartnerOnlyOwnAds(t *testing.T) {
	repo := &fakeAdRepo{
		isOwner:  false,
		existing: &models.Ad{ID: "ad-1", Status: models.AdStatusActive},
	}
	svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

	if _, err := svc.GetAdPerformance("ad-1", "partner-2", false, nil, nil); err == nil {
		t.Fatalf("partner không được xem ad của người khác")
	}
}

// ─── P0-3: lọc danh sách phía server ──────────────────────────────────────

func TestGetPartnerAds_ForwardsFiltersAndPagination(t *testing.T) {
	repo := &fakeAdRepo{partnerTotal: 7}
	svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

	_, total, err := svc.GetPartnerAds("partner-1", 3, 10, "khuyến mãi", "pending")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 7 {
		t.Fatalf("expected total 7, got %d", total)
	}
	if repo.partnerKeyword != "khuyến mãi" || repo.partnerStatus != "pending" {
		t.Fatalf("filter không được truyền xuống repo: %q/%q", repo.partnerKeyword, repo.partnerStatus)
	}
	if repo.partnerLimit != 10 || repo.partnerOffset != 20 {
		t.Fatalf("phân trang sai: limit=%d offset=%d", repo.partnerLimit, repo.partnerOffset)
	}
}

// ─── P0-5: tracking nguyên tử ─────────────────────────────────────────────

func TestTrackUserAction_SkipsInactiveAd(t *testing.T) {
	for _, status := range []models.AdStatus{models.AdStatusPaused, models.AdStatusPending, models.AdStatusSuspended, models.AdStatusCompleted} {
		repo := &fakeAdRepo{existing: &models.Ad{ID: "ad-1", Status: status}}
		svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

		recorded, err := svc.TrackUserAction("ad-1", ptrString("u1"), "impression", "127.0.0.1")
		if err != nil {
			t.Fatalf("status %s: unexpected error: %v", status, err)
		}
		if recorded {
			t.Fatalf("status %s không được ghi nhận tracking", status)
		}
	}
}

func TestTrackUserAction_SkipsExpiredSubscription(t *testing.T) {
	pkgID := "pkg-1"
	repo := &fakeAdRepo{existing: &models.Ad{ID: "ad-1", Status: models.AdStatusActive, PartnerID: "partner-1", PackageID: &pkgID}}
	svc := newTestAdService(repo, &fakePackageRepo{sub: nil})

	recorded, err := svc.TrackUserAction("ad-1", ptrString("u1"), "impression", "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recorded {
		t.Fatalf("gói hết hạn thì không được ghi nhận tracking")
	}
	if repo.tracked != nil {
		t.Fatalf("repo không được gọi khi gói hết hạn")
	}
}

func TestTrackUserAction_RecordsForActiveAd(t *testing.T) {
	pkgID := "pkg-1"
	repo := &fakeAdRepo{existing: &models.Ad{ID: "ad-1", Status: models.AdStatusActive, PartnerID: "partner-1", PackageID: &pkgID}}
	svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

	recorded, err := svc.TrackUserAction("ad-1", ptrString("u1"), "click", "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !recorded {
		t.Fatalf("ad active + gói còn hạn phải được ghi nhận")
	}
	if repo.tracked.ActionType != "click" || repo.tracked.AdID != "ad-1" {
		t.Fatalf("analytics sai: %+v", repo.tracked)
	}
	if repo.tracked.ID == "" || repo.tracked.CreatedAt.IsZero() {
		t.Fatalf("analytics phải có ID và thời gian")
	}
}

// ─── P0-6: serving theo gói/max impressions + worker ──────────────────────

func TestCompleteFinishedAds_DelegatesToRepo(t *testing.T) {
	repo := &fakeAdRepo{completedRows: 3}
	svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

	before := time.Now()
	count, err := svc.CompleteFinishedAds()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 ads completed, got %d", count)
	}
	if repo.completedAt.Before(before) {
		t.Fatalf("phải truyền thời điểm hiện tại cho repo")
	}
}

// ─── P1-2: analytics theo khoảng thời gian ────────────────────────────────

func TestGetAdPerformance_RangeAdminNotLimited(t *testing.T) {
	repo := &fakeAdRepo{
		existing: &models.Ad{ID: "ad-1", PartnerID: "partner-1", Status: models.AdStatusActive},
		rangeMetrics: &repository.AdRangeMetrics{
			Counts:      map[string]int64{"impression": 100, "click": 5},
			Cost:        12.5,
			UniqueReach: 80,
			Series: []dto.AdDailyMetric{
				{Date: "2026-11-01", Impressions: 100, Clicks: 5, Cost: 12.5},
			},
		},
	}
	svc := newTestAdService(repo, &fakePackageRepo{sub: nil})

	from := time.Now().AddDate(0, 0, -90)
	to := time.Now()
	resp, err := svc.GetAdPerformance("ad-1", "admin-1", true, &from, &to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Limited {
		t.Fatalf("admin không bị giới hạn thời gian")
	}
	if len(resp.Series) != 1 || resp.Series[0].Date != "2026-11-01" {
		t.Fatalf("series không đúng: %+v", resp.Series)
	}
	if resp.Cost != 12.5 || resp.UniqueReach != 80 {
		t.Fatalf("cost/unique reach không đúng: %+v", resp)
	}
}

func TestGetAdPerformance_RangeLimitedForBasicPackage(t *testing.T) {
	repo := &fakeAdRepo{
		isOwner:  true,
		existing: &models.Ad{ID: "ad-1", PartnerID: "partner-1", Status: models.AdStatusActive},
		rangeMetrics: &repository.AdRangeMetrics{
			Counts: map[string]int64{"impression": 10},
		},
	}
	// Gói hiện tại không có advanced analytics.
	svc := newTestAdService(repo, &fakePackageRepo{sub: newActiveSubscription()})

	from := time.Now().AddDate(0, 0, -90)
	to := time.Now()
	resp, err := svc.GetAdPerformance("ad-1", "partner-1", false, &from, &to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Limited {
		t.Fatalf("gói BASIC phải bị giới hạn 30 ngày")
	}
	if !repo.rangeFrom.After(from) {
		t.Fatalf("from phải được kéo về mốc 30 ngày, got %v", repo.rangeFrom)
	}
}

func ptrString(v string) *string { return &v }

// ─── C2: thay media khi sửa ad ────────────────────────────────────────────

func TestReplaceMedia_SwapsAndCleansOld(t *testing.T) {
	repo := &fakeAdRepo{
		isOwner: true,
		existing: &models.Ad{
			ID:     "ad-1",
			Status: models.AdStatusActive,
			Format: models.AdFormatImage,
		},
		adMedia: []models.AdMedia{{ID: "old", AdID: "ad-1", URL: "https://cdn/old.jpg"}},
	}
	pkg := &fakePackageRepo{sub: newActiveSubscription()}
	mediaSvc := &uploadFakeMediaService{
		upload: func(context.Context, string, *multipart.FileHeader) (*models.Media, error) {
			return &models.Media{ID: "media-new", FileURI: "https://cdn/new.jpg", FileType: "image/jpeg"}, nil
		},
	}
	svc := NewAdService(repo, pkg, mediaSvc)

	ad, err := svc.ReplaceMedia(context.Background(), "ad-1", "partner-1", []*multipart.FileHeader{
		adUploadFile("new.jpg", "image/jpeg"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.replaced) != 1 || repo.replaced[0].URL != "https://cdn/new.jpg" {
		t.Fatalf("media mới chưa được ghi: %+v", repo.replaced)
	}
	if ad.MediaList == nil || len(ad.MediaList) != 1 {
		t.Fatalf("ad trả về phải kèm media mới")
	}
	if len(mediaSvc.cleaned) != 1 || mediaSvc.cleaned[0] != "https://cdn/old.jpg" {
		t.Fatalf("media cũ phải được dọn, got %v", mediaSvc.cleaned)
	}
	if len(mediaSvc.deleted) != 0 {
		t.Fatalf("không được xoá media mới, got %v", mediaSvc.deleted)
	}
}

func TestReplaceMedia_BlocksCompletedAd(t *testing.T) {
	repo := &fakeAdRepo{
		isOwner:  true,
		existing: &models.Ad{ID: "ad-2", Status: models.AdStatusCompleted, Format: models.AdFormatImage},
	}
	svc := NewAdService(repo, &fakePackageRepo{sub: newActiveSubscription()}, &uploadFakeMediaService{})

	_, err := svc.ReplaceMedia(context.Background(), "ad-2", "partner-1", []*multipart.FileHeader{
		adUploadFile("new.jpg", "image/jpeg"),
	})
	if err == nil {
		t.Fatal("expected error for completed ad")
	}
	appErr, ok := errorsapp.IsAppError(err)
	if !ok || appErr.Code != errorsapp.ErrCodeAdCannotEdit {
		t.Fatalf("expected ad.CANNOT_EDIT, got %v", err)
	}
}

func TestReplaceMedia_RejectsInvalidCount(t *testing.T) {
	repo := &fakeAdRepo{
		isOwner:  true,
		existing: &models.Ad{ID: "ad-3", Status: models.AdStatusPending, Format: models.AdFormatImage},
	}
	svc := NewAdService(repo, &fakePackageRepo{sub: newActiveSubscription()}, &uploadFakeMediaService{})

	_, err := svc.ReplaceMedia(context.Background(), "ad-3", "partner-1", []*multipart.FileHeader{
		adUploadFile("1.jpg", "image/jpeg"),
		adUploadFile("2.jpg", "image/jpeg"),
	})
	if err == nil {
		t.Fatal("expected error for 2 files with image format")
	}
	appErr, ok := errorsapp.IsAppError(err)
	if !ok || appErr.Code != errorsapp.ErrCodeAdMediaCountInvalid {
		t.Fatalf("expected ad.MEDIA_COUNT_INVALID, got %v", err)
	}
	if len(repo.replaced) != 0 {
		t.Fatal("không được thay media khi validate lỗi")
	}
}

// ─── C8: upload media, duration và cleanup ────────────────────────────────

func adUploadFile(name, contentType string) *multipart.FileHeader {
	header := make(textproto.MIMEHeader)
	header.Set("Content-Type", contentType)
	return &multipart.FileHeader{Filename: name, Size: 1024, Header: header}
}

// adUploadContext cần Request thật: service gọi ctx.Request.Context() khi upload.
func adUploadContext() *gin.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/ads-management", nil)
	return ctx
}

func TestCreateAdWithMedia_RejectsVideoOverLimitAndCleansUp(t *testing.T) {
	repo := &fakeAdRepo{}
	pkg := &fakePackageRepo{sub: newActiveSubscription()}
	mediaSvc := &uploadFakeMediaService{
		upload: func(context.Context, string, *multipart.FileHeader) (*models.Media, error) {
			return &models.Media{ID: "media-1", FileURI: "https://cdn/v.mp4", FileType: "video/mp4"}, nil
		},
		duration: func(context.Context, string) (float64, error) { return 61, nil },
	}
	svc := NewAdService(repo, pkg, mediaSvc)

	input := validCreateInput()
	input.Format = "video"
	ctx := adUploadContext()

	_, err := svc.CreateAdWithMedia(ctx, input, []*multipart.FileHeader{adUploadFile("v.mp4", "video/mp4")}, "partner-1", false)
	if err == nil {
		t.Fatal("expected error khi video dài quá giới hạn")
	}
	appErr, ok := errorsapp.IsAppError(err)
	if !ok || appErr.Code != errorsapp.ErrCodeAdVideoTooLong {
		t.Fatalf("expected ad.VIDEO_TOO_LONG, got %v", err)
	}
	if repo.deleted == "" {
		t.Fatal("ad phải được xoá khi upload thất bại")
	}
	if len(mediaSvc.deleted) != 1 || mediaSvc.deleted[0] != "media-1" {
		t.Fatalf("media đã upload phải được dọn, got %v", mediaSvc.deleted)
	}
	if pkg.slotsUsed != 0 {
		t.Fatalf("slot phải được trả lại, slotsUsed = %d", pkg.slotsUsed)
	}
}

func TestCreateAdWithMedia_FailsOpenWhenDurationUnavailable(t *testing.T) {
	repo := &fakeAdRepo{}
	pkg := &fakePackageRepo{sub: newActiveSubscription()}
	mediaSvc := &uploadFakeMediaService{
		upload: func(context.Context, string, *multipart.FileHeader) (*models.Media, error) {
			return &models.Media{ID: "media-1", FileURI: "https://cdn/v.mp4", FileType: "video/mp4"}, nil
		},
		duration: func(context.Context, string) (float64, error) {
			return 0, errors.New("admin api down")
		},
	}
	svc := NewAdService(repo, pkg, mediaSvc)

	input := validCreateInput()
	input.Format = "video"
	ctx := adUploadContext()

	ad, err := svc.CreateAdWithMedia(ctx, input, []*multipart.FileHeader{adUploadFile("v.mp4", "video/mp4")}, "partner-1", false)
	if err != nil {
		t.Fatalf("không đọc được duration phải fail-open, got %v", err)
	}
	if ad == nil || len(repo.mediaBatch) != 1 {
		t.Fatalf("ad và media phải được tạo, mediaBatch = %d", len(repo.mediaBatch))
	}
	if len(mediaSvc.deleted) != 0 {
		t.Fatalf("không được xoá media khi fail-open, got %v", mediaSvc.deleted)
	}
}

func TestCreateAdWithMedia_CleansUpWhenSecondUploadFails(t *testing.T) {
	repo := &fakeAdRepo{}
	pkg := &fakePackageRepo{sub: newActiveSubscription()}
	calls := 0
	mediaSvc := &uploadFakeMediaService{
		upload: func(context.Context, string, *multipart.FileHeader) (*models.Media, error) {
			calls++
			if calls == 2 {
				return nil, errors.New("upload boom")
			}
			return &models.Media{ID: "media-1", FileURI: "https://cdn/1.jpg", FileType: "image/jpeg"}, nil
		},
	}
	svc := NewAdService(repo, pkg, mediaSvc)

	input := validCreateInput()
	input.Format = "carousel"
	files := []*multipart.FileHeader{
		adUploadFile("1.jpg", "image/jpeg"),
		adUploadFile("2.jpg", "image/jpeg"),
	}
	ctx := adUploadContext()

	if _, err := svc.CreateAdWithMedia(ctx, input, files, "partner-1", false); err == nil {
		t.Fatal("expected error khi file thứ hai upload lỗi")
	}
	if repo.deleted == "" {
		t.Fatal("ad phải được xoá")
	}
	if len(mediaSvc.deleted) != 1 || mediaSvc.deleted[0] != "media-1" {
		t.Fatalf("file đầu phải được dọn, got %v", mediaSvc.deleted)
	}
	if pkg.slotsUsed != 0 {
		t.Fatalf("slot phải được trả lại, slotsUsed = %d", pkg.slotsUsed)
	}
}
