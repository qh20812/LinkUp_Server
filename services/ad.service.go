package services

import (
	"crypto/rand"
	"fmt"
	"mime/multipart"
	"time"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"

	"github.com/gin-gonic/gin"
)

type AdService interface {
	CreateAdWithMedia(ctx *gin.Context, input dto.CreateAdInput, files []*multipart.FileHeader, partnerID string) (*models.Ad, error)
	UpdateStatus(id string, statusStr string, partnerID string) (*models.Ad, error)
	GetAdPerformance(id string, partnerID string) (*dto.AdPerformanceResponse, error)
	GetDashboardList() ([]models.Ad, error)
	GetAdsForUserFeed() ([]models.Ad, error)
	TrackUserAction(adID string, userID *string, actionType, ip string) error
	GetAdByID(id string) (*models.Ad, error)
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
	ad.ID = uuidGenerate()
	ad.PartnerID = partnerID
	ad.PackageID = &sub.PackageID
	ad.DailyBudget = input.DailyBudget
	ad.CPMPrice = input.CPMPrice
	ad.CPCPrice = input.CPCPrice
	ad.MaxImpressions = input.MaxImpressions
	ad.StartedAt = input.StartedAt
	ad.ExpiresAt = input.ExpiresAt
	ad.CreatedAt = time.Now()

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

func (s *adServiceImpl) GetAdsForUserFeed() ([]models.Ad, error) {
	return s.repo.FindActiveAds(time.Now())
}

func (s *adServiceImpl) TrackUserAction(adID string, userID *string, actionType, ip string) error {
	ad, err := s.repo.FindByID(adID)
	if err != nil {
		return err
	}

	log := models.NewAdAnalytics(adID, userID, actionType, ip)
	log.ID = uuidGenerate()
	log.CreatedAt = time.Now()

	// Ghi nhận tương tác và tính toán khấu trừ ngân sách
	return s.repo.TrackActionAndDeduct(&log, ad.CPMPrice, ad.CPCPrice)
}

func uuidGenerate() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("ad_%x", b)
}

func generateMediaUUID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("media_%x", b)
}
