package services

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"

	"github.com/gin-gonic/gin"
)

type AdService interface {
	CreateAdWithMedia(ctx *gin.Context, input dto.CreateAdInput, files []*multipart.FileHeader, partnerID string, platformAdmin bool) (*models.Ad, error)
	UpdateStatus(id string, statusStr string, partnerID string) (*models.Ad, error)
	GetAdPerformance(id string, viewerID string, viewerIsAdmin bool, from, to *time.Time) (*dto.AdPerformanceResponse, error)
	GetDashboardList() ([]models.Ad, error)
	GetPartnerAds(partnerID string, page, pageSize int, keyword, status string) ([]dto.PartnerAdListItem, int64, error)
	GetAdminAds(page, pageSize int, keyword, status string) ([]dto.AdminAdListItem, int64, error)
	GetAdsForUserFeed(userID string, userGender string, userAge int, userLocation string) ([]models.Ad, error)
	TrackUserAction(adID string, userID *string, actionType, ip string) (bool, error)
	GetAdByID(id string) (*models.Ad, error)
	UpdateAd(id, partnerID string, input dto.UpdateAdInput) (*models.Ad, error)
	DeleteAd(id, partnerID string) error
	SubmitForReview(id, partnerID string) (*models.Ad, error)
	SetAdStatus(id string, status models.AdStatus) error
	GetOverview(partnerID string) (*dto.AdOverviewResponse, error)
	// CompleteFinishedAds chuyển ad active đã hết hạn / hết budget / đạt
	// max_impressions sang completed (dùng cho worker định kỳ).
	CompleteFinishedAds() (int, error)
}

type adServiceImpl struct {
	repo         repository.AdRepository
	packageRepo  repository.PackageRepository
	mediaService MediaService
}

// targetLocationsJSON luôn trả về một mảng JSON hợp lệ cho cột `target_locations`.
// Cột này kiểu JSON nên chuỗi rỗng bị MySQL từ chối (Error 3140 "The document is
// empty") — trước đây quảng cáo không chọn địa điểm luôn tạo thất bại vì lý do này.
func targetLocationsJSON(locations []string) (string, error) {
	if locations == nil {
		locations = []string{}
	}
	raw, err := json.Marshal(locations)
	if err != nil {
		return "", fmt.Errorf("mã hoá target_locations: %w", err)
	}
	return string(raw), nil
}

func NewAdService(
	repo repository.AdRepository,
	packageRepo repository.PackageRepository,
	mediaService MediaService,
) AdService {
	return &adServiceImpl{
		repo:         repo,
		packageRepo:  packageRepo,
		mediaService: mediaService,
	}
}

// CreateAdWithMedia tạo quảng cáo kèm media.
// platformAdmin = true khi ADMIN/SUPER_ADMIN tạo hộ: bỏ qua kiểm tra gói và
// slot (admin không mua gói), ad không gắn PackageID. PARTNER vẫn bị ràng buộc
// bởi gói đang hoạt động và số slot còn lại.
func (s *adServiceImpl) CreateAdWithMedia(ctx *gin.Context, input dto.CreateAdInput, files []*multipart.FileHeader, partnerID string, platformAdmin bool) (*models.Ad, error) {
	adFormat := models.AdFormat(input.Format)

	// 1. Kiểm tra Subscription & Slot Limit (mục 2.1)[cite: 2]
	var sub *models.PartnerSubscription
	if !platformAdmin {
		var err error
		sub, err = s.packageRepo.GetActiveSubscription(partnerID)
		if err != nil {
			return nil, err
		}
		if sub == nil {
			return nil, errorsapp.New(errorsapp.ErrCodeAdNoSubscription)
		}

		if sub.SlotsUsed >= sub.Package.MaxSlots {
			return nil, errorsapp.Newf(errorsapp.ErrCodeAdSlotsExhausted, map[string]any{"used": sub.SlotsUsed, "max": sub.Package.MaxSlots})
		}

		// 2. Kiểm tra Format có được gói hỗ trợ không
		if adFormat == models.AdFormatVideo && !sub.Package.SupportsVideo {
			return nil, errorsapp.New(errorsapp.ErrCodeAdFormatNotVideo)
		}
		if adFormat == models.AdFormatCarousel && !sub.Package.SupportsCarousel {
			return nil, errorsapp.New(errorsapp.ErrCodeAdFormatNotCarousel)
		}
	}

	// 3. Khởi tạo Quảng cáo
	ad := models.NewAd(input.Title, input.Content, input.TargetURL, input.Budget, adFormat)
	// Partner phải qua bước duyệt; admin tạo hộ thì phát ngay.
	if !platformAdmin {
		ad.Status = models.AdStatusPending
	}
	ad.ID = utils.GenerateUUID()
	ad.PartnerID = partnerID
	if sub != nil {
		ad.PackageID = &sub.PackageID
	}
	ad.DailyBudget = input.DailyBudget
	ad.CPMPrice = input.CPMPrice
	ad.CPCPrice = input.CPCPrice
	ad.MaxImpressions = input.MaxImpressions
	ad.StartedAt = input.StartedAt
	ad.ExpiresAt = input.ExpiresAt
	ad.CreatedAt = time.Now()

	// Targeting fields
	ad.TargetGender = input.TargetGender
	if ad.TargetGender == "" {
		ad.TargetGender = "all"
	}
	ad.TargetAgeMin = input.TargetAgeMin
	ad.TargetAgeMax = input.TargetAgeMax

	// Cột target_locations là kiểu JSON nên KHÔNG được để chuỗi rỗng: MySQL trả
	// Error 3140 ("The document is empty"), khiến mọi quảng cáo không chọn địa
	// điểm đều tạo thất bại. Luôn ghi một mảng JSON hợp lệ.
	locJSON, err := targetLocationsJSON(input.TargetLocations)
	if err != nil {
		return nil, err
	}
	ad.TargetLocations = locJSON

	if err := s.repo.Create(&ad); err != nil {
		return nil, err
	}

	// 4. Upload file đính kèm
	var mediaList []models.AdMedia
	for idx, fileHeader := range files {
		// Media quảng cáo auto-approve; bản thân ad vẫn qua bước admin duyệt
		// (SetAdStatus pending→active) nên không cần moderation ở mức file.
		uploadedMedia, err := s.mediaService.AutoApproveUpload(ctx.Request.Context(), partnerID, fileHeader)
		if err != nil {
			return nil, fmt.Errorf("lỗi upload file %s: %w", fileHeader.Filename, err)
		}

		mediaList = append(mediaList, models.AdMedia{
			ID:        generateMediaUUID(),
			AdID:      ad.ID,
			URL:       uploadedMedia.FileURI,
			MediaType: uploadedMedia.FileType,
			SortOrder: idx,
			CreatedAt: time.Now(),
		})
	}

	if len(mediaList) > 0 {
		if err := s.repo.CreateMediaBatch(mediaList); err != nil {
			return nil, err
		}
		ad.MediaList = mediaList
	}

	// 5. Cập nhật slot (chỉ partner có gói mới có slot)
	if sub != nil {
		_ = s.packageRepo.IncrementSlotsUsed(sub.ID)
	}

	return &ad, nil
}

func (s *adServiceImpl) GetAdByID(id string) (*models.Ad, error) {
	return s.repo.FindByID(id)
}

// UpdateStatus cho partner tạm dừng / bật lại quảng cáo của mình.
// Chỉ chấp nhận active ↔ paused; duyệt, từ chối, takedown và hoàn thành do
// admin/hệ thống quyết định nên partner không được tự đổi sang các trạng thái
// đó (trước đây partner đổi được mọi trạng thái, kể cả tự duyệt).
func (s *adServiceImpl) UpdateStatus(id string, statusStr string, partnerID string) (*models.Ad, error) {
	isOwner, err := s.repo.CheckAdOwnership(id, partnerID)
	if err != nil || !isOwner {
		return nil, errorsapp.New(errorsapp.ErrCodeAdNotOwner)
	}

	ad, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	target := models.ParseAdStatus(statusStr)
	allowed := (ad.Status == models.AdStatusActive && target == models.AdStatusPaused) ||
		(ad.Status == models.AdStatusPaused && target == models.AdStatusActive)
	if !allowed {
		return nil, errorsapp.New(errorsapp.ErrCodeAdCannotEdit)
	}

	// Bật lại chỉ hợp lệ khi còn ngân sách, chưa hết hạn và gói còn hiệu lực.
	if target == models.AdStatusActive {
		if ad.Budget > 0 && ad.TotalSpent >= ad.Budget {
			return nil, errorsapp.New(errorsapp.ErrCodeAdBudgetExceeded)
		}
		if ad.ExpiresAt != nil && ad.ExpiresAt.Before(time.Now()) {
			return nil, errorsapp.New(errorsapp.ErrCodeAdCannotEdit)
		}
		if sub, subErr := s.packageRepo.GetActiveSubscription(partnerID); subErr != nil || sub == nil {
			return nil, errorsapp.New(errorsapp.ErrCodeAdNoSubscription)
		}
	}

	ad.Status = target
	if err := s.repo.Update(ad); err != nil {
		return nil, err
	}
	return ad, nil
}

// SubmitForReview đưa quảng cáo bị từ chối trở lại hàng đợi duyệt và xoá lý do
// từ chối cũ, để partner sửa nội dung rồi gửi duyệt lại.
func (s *adServiceImpl) SubmitForReview(id, partnerID string) (*models.Ad, error) {
	isOwner, err := s.repo.CheckAdOwnership(id, partnerID)
	if err != nil || !isOwner {
		return nil, errorsapp.New(errorsapp.ErrCodeAdNotOwner)
	}

	ad, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if ad.Status != models.AdStatusRejected {
		return nil, errorsapp.New(errorsapp.ErrCodeAdCannotEdit)
	}

	fields := map[string]interface{}{
		"status":           string(models.AdStatusPending),
		"rejection_reason": nil,
	}
	if err := s.repo.UpdateAdFields(ad, fields); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// GetAdPerformance trả số liệu của một quảng cáo. Partner chỉ xem được ad của
// mình; admin (viewerIsAdmin) xem được mọi ad — trước đây admin bị chặn 403 vì
// service kiểm tra ownership bằng userID của admin.
//
// from/to nil = số liệu toàn đời (tương thích client cũ). Khi có khoảng thời
// gian, trả thêm chuỗi theo ngày; gói không có advanced analytics bị giới hạn
// 30 ngày gần nhất (cờ limited).
func (s *adServiceImpl) GetAdPerformance(id string, viewerID string, viewerIsAdmin bool, from, to *time.Time) (*dto.AdPerformanceResponse, error) {
	if !viewerIsAdmin && viewerID != "" {
		isOwner, err := s.repo.CheckAdOwnership(id, viewerID)
		if err != nil || !isOwner {
			return nil, errorsapp.New(errorsapp.ErrCodeAdNotOwner)
		}
	}

	ad, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if from == nil || to == nil {
		counts, err := s.repo.GetActionCounts(id)
		if err != nil {
			return nil, err
		}
		uniqueReach, err := s.repo.GetUniqueReach(id)
		if err != nil {
			uniqueReach = 0
		}
		return buildAdPerformanceResponse(ad, counts, uniqueReach, ad.TotalSpent), nil
	}

	start, end := *from, *to
	if !end.After(start) {
		return nil, errorsapp.New(errorsapp.ErrCodeInvalidInput)
	}

	limited := false
	if !viewerIsAdmin {
		if sub, subErr := s.packageRepo.GetActiveSubscription(ad.PartnerID); subErr != nil ||
			sub == nil || sub.Package == nil || !sub.Package.HasAdvancedAnalytics {
			maxStart := time.Now().AddDate(0, 0, -30)
			if start.Before(maxStart) {
				start = maxStart
				limited = true
			}
		}
	}

	metrics, err := s.repo.GetAdMetricsRange(id, start, end)
	if err != nil {
		return nil, err
	}

	resp := buildAdPerformanceResponse(ad, metrics.Counts, metrics.UniqueReach, metrics.Cost)
	resp.Cost = metrics.Cost
	resp.From = start.Format(time.RFC3339)
	resp.To = end.Format(time.RFC3339)
	resp.Limited = limited
	resp.Series = metrics.Series
	return resp, nil
}

// buildAdPerformanceResponse dựng response từ counts + chi phí của phạm vi đang
// xem (toàn đời hoặc một khoảng). Budget/remaining luôn tính trên toàn đời.
func buildAdPerformanceResponse(ad *models.Ad, counts map[string]int64, uniqueReach int64, spent float64) *dto.AdPerformanceResponse {
	impressions := counts[string(models.ActionImpression)]
	clicks := counts[string(models.ActionClick)]
	videoStarts := counts[string(models.ActionVideoStart)]
	videoEnds := counts[string(models.ActionVideoEnd)]
	views := counts[string(models.ActionView)]
	swipeCount := counts[string(models.ActionSwipe)]

	interactions := views + clicks + swipeCount + videoStarts

	lifetimeSpent := ad.TotalSpent
	if ad.Budget > 0 && lifetimeSpent > ad.Budget {
		lifetimeSpent = ad.Budget
	}
	remainingBudget := ad.Budget - lifetimeSpent
	if remainingBudget < 0 {
		remainingBudget = 0
	}

	var ctr, cpc, cpm float64
	if impressions > 0 {
		ctr = (float64(clicks) / float64(impressions)) * 100.0
		cpm = (spent / float64(impressions)) * 1000.0
	}
	if clicks > 0 {
		cpc = spent / float64(clicks)
	}

	return &dto.AdPerformanceResponse{
		AdID:             ad.ID,
		Title:            ad.Title,
		Status:           string(ad.Status),
		Format:           string(ad.Format),
		Budget:           ad.Budget,
		DailyBudget:      ad.DailyBudget,
		TotalSpent:       lifetimeSpent,
		RemainingBudget:  remainingBudget,
		Impressions:      impressions,
		UniqueReach:      uniqueReach,
		Clicks:           clicks,
		Interactions:     interactions,
		CTR:              ctr,
		CPC:              cpc,
		CPM:              cpm,
		VideoStarts:      videoStarts,
		VideoCompletions: videoEnds,
		StartedAt:        ad.StartedAt,
		ExpiresAt:        ad.ExpiresAt,
		Cost:             spent,
	}
}

func (s *adServiceImpl) GetDashboardList() ([]models.Ad, error) {
	return s.repo.GetAll()
}

// GetPartnerAds trả về ads của riêng partner với metrics + lọc keyword/status,
// có phân trang (mục 6.4).
func (s *adServiceImpl) GetPartnerAds(partnerID string, page, pageSize int, keyword, status string) ([]dto.PartnerAdListItem, int64, error) {
	total, err := s.repo.CountByPartnerID(partnerID, keyword, status)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	ads, err := s.repo.ListByPartnerID(partnerID, keyword, status, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	if ads == nil {
		ads = []dto.PartnerAdListItem{}
	}
	return ads, total, nil
}

// GetAdminAds trả danh sách quảng cáo cho admin (mọi partner) kèm thông tin chủ
// quảng cáo và số liệu, có lọc + phân trang. Dùng cho nhánh admin của
// /api/ads-management thay vì kéo toàn bộ bảng.
func (s *adServiceImpl) GetAdminAds(page, pageSize int, keyword, status string) ([]dto.AdminAdListItem, int64, error) {
	total, err := s.repo.CountAds(context.Background(), keyword, status)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	ads, err := s.repo.ListAds(context.Background(), keyword, status, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	if ads == nil {
		ads = []dto.AdminAdListItem{}
	}
	return ads, total, nil
}

// UpdateAd cập nhật thông tin ad (chỉ khi status active/paused)
func (s *adServiceImpl) UpdateAd(id, partnerID string, input dto.UpdateAdInput) (*models.Ad, error) {
	isOwner, err := s.repo.CheckAdOwnership(id, partnerID)
	if err != nil || !isOwner {
		return nil, errorsapp.New(errorsapp.ErrCodeAdNotOwner)
	}

	ad, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Chỉ sửa nội dung khi quảng cáo chưa kết thúc vòng đời. Cho phép cả
	// pending/rejected để partner sửa trước khi gửi duyệt lại.
	switch ad.Status {
	case models.AdStatusPending, models.AdStatusRejected, models.AdStatusActive, models.AdStatusPaused:
	default:
		return nil, errorsapp.New(errorsapp.ErrCodeAdCannotEdit)
	}

	fields := map[string]interface{}{}
	if input.Title != "" {
		fields["title"] = input.Title
	}
	if input.Content != "" {
		fields["content"] = input.Content
	}
	if input.TargetURL != "" {
		fields["target_url"] = input.TargetURL
	}
	if input.Budget > 0 {
		fields["budget"] = input.Budget
	}
	if input.DailyBudget >= 0 {
		fields["daily_budget"] = input.DailyBudget
	}
	if input.CPMPrice >= 0 {
		fields["cpm_price"] = input.CPMPrice
	}
	if input.CPCPrice >= 0 {
		fields["cpc_price"] = input.CPCPrice
	}
	if input.MaxImpressions >= 0 {
		fields["max_impressions"] = input.MaxImpressions
	}
	if input.StartedAt != nil {
		fields["started_at"] = input.StartedAt
	}
	if input.ExpiresAt != nil {
		fields["expires_at"] = input.ExpiresAt
	}
	if input.TargetGender != "" {
		fields["target_gender"] = input.TargetGender
	}
	if input.TargetAgeMin > 0 {
		fields["target_age_min"] = input.TargetAgeMin
	}
	if input.TargetAgeMax > 0 && input.TargetAgeMax < 100 {
		fields["target_age_max"] = input.TargetAgeMax
	}
	if input.TargetLocations != nil {
		locJSON, err := targetLocationsJSON(input.TargetLocations)
		if err != nil {
			return nil, err
		}
		fields["target_locations"] = locJSON
	}

	if len(fields) > 0 {
		if err := s.repo.UpdateAdFields(ad, fields); err != nil {
			return nil, err
		}
	}

	return s.repo.FindByID(id)
}

// SetAdStatus cập nhật trạng thái ad trực tiếp (dùng cho moderation khi tạo)
func (s *adServiceImpl) SetAdStatus(id string, status models.AdStatus) error {
	ad, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	ad.Status = status
	return s.repo.Update(ad)
}

// DeleteAd xóa ad + giải phóng slot (hard delete)
func (s *adServiceImpl) DeleteAd(id, partnerID string) error {
	isOwner, err := s.repo.CheckAdOwnership(id, partnerID)
	if err != nil || !isOwner {
		return errorsapp.New(errorsapp.ErrCodeAdNotOwner)
	}

	// Xác nhận ad tồn tại trước khi xoá.
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}

	// Không chặn theo trạng thái: quảng cáo tạo ra là active ngay (bỏ bước duyệt
	// pending) nên nếu chỉ cho xoá pending/rejected/paused thì đối tác tạo xong
	// là không thể xoá, slot trong gói bị kẹt vĩnh viễn.

	// Xóa media trước
	_ = s.repo.DeleteAdMedia(id)

	// Xóa ad
	if err := s.repo.Delete(context.Background(), id); err != nil {
		return err
	}

	// Giải phóng slot trong subscription. Trước đây truyền ad.PackageID (id gói)
	// nên câu UPDATE không khớp bản ghi subscription nào — slot không bao giờ
	// được trả lại và đối tác sẽ cạn slot sau vài lần xoá.
	if sub, err := s.packageRepo.GetActiveSubscription(partnerID); err == nil && sub != nil {
		_ = s.packageRepo.DecrementSlotsUsed(sub.ID)
	}

	return nil
}

// GetAdsForUserFeed trả về ads ranked + filtered theo frequency cap (mục 4.2 + 4.3)
func (s *adServiceImpl) GetAdsForUserFeed(userID string, userGender string, userAge int, userLocation string) ([]models.Ad, error) {
	if userGender == "" {
		userGender = "all"
	}

	// 1. Query ranked ads (lấy nhiều hơn limit để filter)
	allAds, err := s.repo.FindRankedAdsForUser(time.Now(), userGender, userAge, userLocation, 50)
	if err != nil {
		return nil, err
	}

	// 2. Filter theo frequency cap
	filtered, err := s.repo.FilterByFrequencyCap(allAds, userID, 20)
	if err != nil {
		return allAds, nil
	}

	return filtered, nil
}

// TrackUserAction ghi nhận một tương tác quảng cáo. Trả recorded=false khi ad
// không ở trạng thái chạy, gói của partner hết hạn, hoặc bị chặn bởi cap
// (chống trùng/budget/max impressions) — những trường hợp này không phải lỗi.
func (s *adServiceImpl) TrackUserAction(adID string, userID *string, actionType, ip string) (bool, error) {
	ad, err := s.repo.FindByID(adID)
	if err != nil {
		return false, err
	}

	// Chỉ ad đang chạy mới được ghi nhận/trừ tiền.
	if ad.Status != models.AdStatusActive {
		return false, nil
	}

	// Ad của partner chỉ hợp lệ khi gói còn hiệu lực. Ad do admin tạo (không
	// gắn PackageID) vẫn được ghi nhận bình thường.
	if ad.PackageID != nil {
		if sub, subErr := s.packageRepo.GetActiveSubscription(ad.PartnerID); subErr != nil || sub == nil {
			return false, nil
		}
	}

	log := models.NewAdAnalytics(adID, userID, actionType, ip)
	log.ID = utils.GenerateUUID()
	log.CreatedAt = time.Now()

	// Toàn bộ check chống trùng / daily budget / budget / max impressions nằm
	// trong transaction của repo.
	return s.repo.TrackActionAndDeduct(&log)
}

func generateMediaUUID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("media_%x", b)
}

// CompleteFinishedAds dọn các quảng cáo đã kết thúc vòng đời: quá expires_at,
// hết budget hoặc đạt max_impressions nhưng vẫn còn status active.
func (s *adServiceImpl) CompleteFinishedAds() (int, error) {
	rows, err := s.repo.MarkFinishedAdsCompleted(time.Now())
	return int(rows), err
}

// GetOverview trả về tổng quan quảng cáo của partner (mục 6.1)
func (s *adServiceImpl) GetOverview(partnerID string) (*dto.AdOverviewResponse, error) {
	return s.repo.GetOverviewByPartner(partnerID)
}
