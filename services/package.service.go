package services

import (
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"time"
)

type PackageService interface {
	GetPackages() ([]models.AdPackage, error)
	SubscribePackage(userID string, packageID string) (*models.PartnerSubscription, error)
	GetUserSubscription(userID string) (*dto.SubscriptionResponse, error)
	ProcessExpiredSubscriptions() (int, error)
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
	expiresAt := now.AddDate(0, 0, pkg.MaxDurationDays)

	newSub := &models.PartnerSubscription{
		ID:        utils.GenerateUUID(),
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

// ProcessExpiredSubscriptions tìm và xử lý các subscription đã hết hạn (mục 1.3)
// Trả về số lượng subscription đã xử lý.
func (s *packageServiceImpl) ProcessExpiredSubscriptions() (int, error) {
	subs, err := s.repo.FindExpiredSubscriptions()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, sub := range subs {
		if err := s.repo.ExpireSubscription(sub.ID); err != nil {
			continue
		}
		if err := s.repo.DowngradePartnerRole(sub.UserID); err != nil {
			continue
		}
		count++
	}
	return count, nil
}
