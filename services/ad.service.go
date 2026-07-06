package services

import (
	"crypto/rand"
	"fmt"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"time"
)

type AdService interface {
	CreateAd(input dto.CreateAdInput, partnerID string) (*models.Ad, error)
	UpdateStatus(id string, statusStr string) (*models.Ad, error)
	GetAdPerformance(id string) (*dto.AdPerformanceResponse, error)
	GetDashboardList() ([]models.Ad, error)
	GetAdsForUserFeed() ([]models.Ad, error)
	TrackUserAction(adID string, userID *string, actionType, ip string) error
	GetAdByID(id string) (*models.Ad, error)
}

type adServiceImpl struct {
	repo repository.AdRepository
}

func NewAdService(repo repository.AdRepository) AdService {
	return &adServiceImpl{repo: repo}
}

func (s *adServiceImpl) CreateAd(input dto.CreateAdInput, partnerID string) (*models.Ad, error) {
	ad := models.NewAd(input.Title, input.Content, input.TargetURL, input.Budget)
	ad.ID = uuidGenerate()
	ad.PartnerID = partnerID
	ad.MediaID = input.MediaID
	ad.StartedAt = input.StartedAt
	ad.ExpiresAt = input.ExpiresAt
	ad.CreatedAt = time.Now()

	if err := s.repo.Create(&ad); err != nil {
		return nil, err
	}
	return &ad, nil
}

func (s *adServiceImpl) GetAdByID(id string) (*models.Ad, error) {
	return s.repo.FindByID(id)
}

func (s *adServiceImpl) UpdateStatus(id string, statusStr string) (*models.Ad, error) {
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

func (s *adServiceImpl) GetAdPerformance(id string) (*dto.AdPerformanceResponse, error) {
	ad, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	imp, clk, intt, err := s.repo.GetCountsByAction(id)
	if err != nil {
		return nil, err
	}

	var ctr float64
	if imp > 0 {
		ctr = (float64(clk) / float64(imp)) * 100
	}

	return &dto.AdPerformanceResponse{
		AdID:         ad.ID,
		Title:        ad.Title,
		Status:       string(ad.Status),
		Budget:       ad.Budget,
		Impressions:  imp,
		Clicks:       clk,
		Interactions: intt,
		CTR:          ctr,
	}, nil
}

func (s *adServiceImpl) GetDashboardList() ([]models.Ad, error) {
	return s.repo.GetAll()
}

func (s *adServiceImpl) GetAdsForUserFeed() ([]models.Ad, error) {
	return s.repo.FindActiveAds(time.Now())
}

func (s *adServiceImpl) TrackUserAction(adID string, userID *string, actionType, ip string) error {
	_, err := s.repo.FindByID(adID)
	if err != nil {
		return err
	}

	log := models.NewAdAnalytics(adID, userID, actionType, ip)
	log.ID = uuidGenerate()
	log.CreatedAt = time.Now()
	return s.repo.LogAnalytics(&log)
}

func uuidGenerate() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("ad_%x", b)
}
