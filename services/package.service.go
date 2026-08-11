package services

import (
	"crypto/rand"
	"fmt"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"time"
)

type PackageService interface {
	GetPackages() ([]models.AdPackage, error)
	SubscribePackage(userID string, packageID string) (*models.PartnerSubscription, error)
	GetUserSubscription(userID string) (*dto.SubscriptionResponse, error)
}

type packageServiceImpl struct {
	repo repository.PackageRepository
}

func NewPackageService(repo repository.PackageRepository) PackageService {
	return &packageServiceImpl{repo: repo}
}

func (s *packageServiceImpl) GetPackages() ([]models.AdPackage, error) {
	return s.repo.GetAllPackages()
}

func (s *packageServiceImpl) SubscribePackage(userID string, packageID string) (*models.PartnerSubscription, error) {
	pkg, err := s.repo.GetPackageByID(packageID)
	if err != nil {
		return nil, err
	}

	activeSub, err := s.repo.GetActiveSubscription(userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiresAt := now.AddDate(0, 1, 0) // Hạn 1 tháng

	newSub := &models.PartnerSubscription{
		ID:        generateUUID("sub_"),
		UserID:    userID,
		PackageID: pkg.ID,
		SlotsUsed: 0,
		StartedAt: now,
		ExpiresAt: expiresAt,
		Status:    models.SubscriptionStatusActive,
		AutoRenew: true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Gọi Transaction: tự động xử lý hủy gói cũ + tạo gói mới + nâng role người dùng
	if err := s.repo.SubscribeWithTransaction(activeSub, newSub); err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodePackageSubscribeFailed, err)
	}

	return newSub, nil
}

func (s *packageServiceImpl) GetUserSubscription(userID string) (*dto.SubscriptionResponse, error) {
	sub, err := s.repo.GetActiveSubscription(userID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, errorsapp.New(errorsapp.ErrCodePackageNotSubscribed)
	}

	slotsLeft := sub.Package.MaxSlots - sub.SlotsUsed
	if slotsLeft < 0 {
		slotsLeft = 0
	}

	return &dto.SubscriptionResponse{
		ID:           sub.ID,
		PackageName:  sub.Package.Name,
		MaxSlots:     sub.Package.MaxSlots,
		SlotsUsed:    sub.SlotsUsed,
		SlotsLeft:    slotsLeft,
		PriceMonthly: sub.Package.PriceMonthly,
		StartedAt:    sub.StartedAt,
		ExpiresAt:    sub.ExpiresAt,
		Status:       string(sub.Status),
	}, nil
}

func generateUUID(prefix string) string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%s%x", prefix, b)
}
