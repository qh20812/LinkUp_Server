package services

import (
	"context"

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
	// SetNotificationService để báo cho partner khi gói hết hạn.
	SetNotificationService(notifier *NotificationService)
}

type packageServiceImpl struct {
	repo     repository.PackageRepository
	notifier *NotificationService
}

func NewPackageService(repo repository.PackageRepository) PackageService {
	return &packageServiceImpl{repo: repo}
}

func (s *packageServiceImpl) SetNotificationService(notifier *NotificationService) {
	s.notifier = notifier
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

	newSub := newPartnerSubscription(userID, pkg)

	// Gọi Transaction: tự động xử lý hủy gói cũ + tạo gói mới + nâng role người dùng
	if err := s.repo.SubscribeWithTransaction(activeSub, newSub); err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodePackageSubscribeFailed, err)
	}

	return newSub, nil
}

// newPartnerSubscription dựng bản ghi gói mới từ gói hiện tại. Dùng chung cho
// luồng đăng ký trực tiếp và luồng thanh toán VNPay để hai bên luôn tính cùng
// thời hạn/slot.
func newPartnerSubscription(userID string, pkg *models.AdPackage) *models.PartnerSubscription {
	now := time.Now()
	return &models.PartnerSubscription{
		ID:        utils.GenerateUUID(),
		UserID:    userID,
		PackageID: pkg.ID,
		SlotsUsed: 0,
		StartedAt: now,
		ExpiresAt: now.AddDate(0, 0, pkg.MaxDurationDays),
		Status:    models.SubscriptionStatusActive,
		AutoRenew: true,
		CreatedAt: now,
		UpdatedAt: now,
	}
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
		// Gói vừa hết hạn: báo partner để họ biết quảng cáo ngừng phân phối.
		// Gửi đúng 1 lần tại thời điểm expire (không lặp theo giờ).
		if s.notifier != nil {
			_, _ = s.notifier.Create(
				context.Background(),
				sub.UserID,
				nil,
				models.NotificationTypeMessage,
				"Gói quảng cáo của bạn đã hết hạn. Quảng cáo sẽ tạm ngưng phân phối cho tới khi bạn gia hạn.",
				nil, nil, nil,
			)
		}
		count++
	}
	return count, nil
}
