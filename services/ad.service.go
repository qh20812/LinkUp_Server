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
	CreateAdWithMedia(ctx *gin.Context, input dto.CreateAdInput, files []*multipart.FileHeader, partnerID string) (*models.Ad, error)
	UpdateStatus(id string, statusStr string, partnerID string) (*models.Ad, error)
	GetAdPerformance(id string, partnerID string) (*dto.AdPerformanceResponse, error)
	GetDashboardList() ([]models.Ad, error)
	GetPartnerAds(partnerID string, page, pageSize int) ([]dto.PartnerAdListItem, int64, error)
	GetAdsForUserFeed(userID string, userGender string, userAge int, userLocation string) ([]models.Ad, error)
	TrackUserAction(adID string, userID *string, actionType, ip string) error
	GetAdByID(id string) (*models.Ad, error)
	UpdateAd(id, partnerID string, input dto.UpdateAdInput) (*models.Ad, error)
	DeleteAd(id, partnerID string) error
	SetAdStatus(id string, status models.AdStatus) error
	GetOverview(partnerID string) (*dto.AdOverviewResponse, error)
}

type adServiceImpl struct {
	repo         repository.AdRepository
	packageRepo  repository.PackageRepository
	mediaService MediaService
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

func (s *adServiceImpl) CreateAdWithMedia(ctx *gin.Context, input dto.CreateAdInput, files []*multipart.FileHeader, partnerID string) (*models.Ad, error) {
	// 1. Kiểm tra Subscription & Slot Limit (mục 2.1)[cite: 2]
	sub, err := s.packageRepo.GetActiveSubscription(partnerID)
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
	adFormat := models.AdFormat(input.Format)
	if adFormat == models.AdFormatVideo && !sub.Package.SupportsVideo {
		return nil, errorsapp.New(errorsapp.ErrCodeAdFormatNotVideo)
	}
	if adFormat == models.AdFormatCarousel && !sub.Package.SupportsCarousel {
		return nil, errorsapp.New(errorsapp.ErrCodeAdFormatNotCarousel)
	}

	// 3. Khởi tạo Quảng cáo
	ad := models.NewAd(input.Title, input.Content, input.TargetURL, input.Budget, adFormat)
	ad.ID = utils.GenerateUUID()
	ad.PartnerID = partnerID
	ad.PackageID = &sub.PackageID
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
	if input.TargetLocations != nil {
		locJSON, _ := json.Marshal(input.TargetLocations)
		ad.TargetLocations = string(locJSON)
	}

	if err := s.repo.Create(&ad); err != nil {
		return nil, err
	}

	// 4. Upload file đính kèm
	var mediaList []models.AdMedia
	for idx, fileHeader := range files {
		uploadedMedia, err := s.mediaService.UploadMedia(ctx.Request.Context(), partnerID, fileHeader)
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

	// 5. Cập nhật slot
	_ = s.packageRepo.IncrementSlotsUsed(sub.ID)

	return &ad, nil
}

func (s *adServiceImpl) GetAdByID(id string) (*models.Ad, error) {
	return s.repo.FindByID(id)
}

func (s *adServiceImpl) UpdateStatus(id string, statusStr string, partnerID string) (*models.Ad, error) {
	isOwner, err := s.repo.CheckAdOwnership(id, partnerID)
	if err != nil || !isOwner {
		return nil, errorsapp.New(errorsapp.ErrCodeAdNotOwner)
	}

	ad, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	ad.Status = models.ParseAdStatus(statusStr)
	if err := s.repo.Update(ad); err != nil {
		return nil, err
	}
	return ad, nil
}

func (s *adServiceImpl) GetAdPerformance(id string, partnerID string) (*dto.AdPerformanceResponse, error) {
	if partnerID != "" {
		isOwner, err := s.repo.CheckAdOwnership(id, partnerID)
		if err != nil || !isOwner {
			return nil, errorsapp.New(errorsapp.ErrCodeAdNotOwner)
		}
	}

	ad, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	counts, err := s.repo.GetActionCounts(id)
	if err != nil {
		return nil, err
	}

	uniqueReach, err := s.repo.GetUniqueReach(id)
	if err != nil {
		uniqueReach = 0
	}

	impressions := counts[string(models.ActionImpression)]
	clicks := counts[string(models.ActionClick)]
	videoStarts := counts[string(models.ActionVideoStart)]
	videoEnds := counts[string(models.ActionVideoEnd)]
	views := counts[string(models.ActionView)]
	swipeCount := counts[string(models.ActionSwipe)]

	// Interactions = views + clicks + swipes + video_start (tất cả tương tác có ý nghĩa)
	interactions := views + clicks + swipeCount + videoStarts

	// Chi phí đã sử dụng
	totalSpent := ad.TotalSpent
	if totalSpent > ad.Budget && ad.Budget > 0 {
		totalSpent = ad.Budget
	}

	remainingBudget := ad.Budget - totalSpent
	if remainingBudget < 0 {
		remainingBudget = 0
	}

	// Tính toán CTR, CPC, CPM chuẩn xác (mục 2.4)[cite: 2]
	var ctr, cpc, cpm float64
	if impressions > 0 {
		ctr = (float64(clicks) / float64(impressions)) * 100.0
		cpm = (totalSpent / float64(impressions)) * 1000.0
	}
	if clicks > 0 {
		cpc = totalSpent / float64(clicks)
	}

	return &dto.AdPerformanceResponse{
		AdID:             ad.ID,
		Title:            ad.Title,
		Status:           string(ad.Status),
		Format:           string(ad.Format),
		Budget:           ad.Budget,
		DailyBudget:      ad.DailyBudget,
		TotalSpent:       totalSpent,
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
	}, nil
}

func (s *adServiceImpl) GetDashboardList() ([]models.Ad, error) {
	return s.repo.GetAll()
}

// GetPartnerAds trả về ads của riêng partner với metrics, có phân trang (mục 6.4)
func (s *adServiceImpl) GetPartnerAds(partnerID string, page, pageSize int) ([]dto.PartnerAdListItem, int64, error) {
	total, err := s.repo.CountByPartnerID(partnerID)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	ads, err := s.repo.ListByPartnerID(partnerID, pageSize, offset)
	if err != nil {
		return nil, 0, err
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

	if ad.Status != models.AdStatusActive && ad.Status != models.AdStatusPaused {
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
		locJSON, _ := json.Marshal(input.TargetLocations)
		fields["target_locations"] = string(locJSON)
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

	ad, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// Chỉ cho phép xóa khi status pending/rejected/paused
	if ad.Status != models.AdStatusPending && ad.Status != models.AdStatusRejected && ad.Status != models.AdStatusPaused {
		return errorsapp.New(errorsapp.ErrCodeAdCannotDelete)
	}

	// Xóa media trước
	_ = s.repo.DeleteAdMedia(id)

	// Xóa ad
	if err := s.repo.Delete(context.Background(), id); err != nil {
		return err
	}

	// Giải phóng slot trong subscription
	if ad.PackageID != nil {
		_ = s.packageRepo.DecrementSlotsUsed(*ad.PackageID)
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

func (s *adServiceImpl) TrackUserAction(adID string, userID *string, actionType, ip string) error {
	ad, err := s.repo.FindByID(adID)
	if err != nil {
		return err
	}

	// Anti-fraud: skip nếu user đã track cùng action trong 24h (mục 2.2)
	if userID != nil && *userID != "" {
		hasRecent, _ := s.repo.HasRecentAction(adID, *userID, actionType, 24*time.Hour)
		if hasRecent {
			return nil
		}
	}

	// Daily budget check (mục 2.4)
	if ad.DailyBudget > 0 {
		dailySpend, _ := s.repo.GetDailySpend(adID, time.Now())
		if dailySpend >= ad.DailyBudget {
			return nil
		}
	}

	log := models.NewAdAnalytics(adID, userID, actionType, ip)
	log.ID = utils.GenerateUUID()
	log.CreatedAt = time.Now()

	// Ghi nhận tương tác và tính toán khấu trừ ngân sách
	return s.repo.TrackActionAndDeduct(&log, ad.CPMPrice, ad.CPCPrice)
}

func generateMediaUUID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("media_%x", b)
}

// GetOverview trả về tổng quan quảng cáo của partner (mục 6.1)
func (s *adServiceImpl) GetOverview(partnerID string) (*dto.AdOverviewResponse, error) {
	return s.repo.GetOverviewByPartner(partnerID)
}
