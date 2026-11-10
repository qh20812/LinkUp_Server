package services

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"time"
)

type PackageService interface {
	GetPackages() ([]models.AdPackage, error)
	AdminListPackages() ([]models.AdPackage, error)
	AdminCreatePackage(input dto.PackageInput) (*models.AdPackage, error)
	AdminUpdatePackage(id string, input dto.PackageInput) (*models.AdPackage, error)
	SubscribePackage(userID string, packageID string) (*models.PartnerSubscription, error)
	GetUserSubscription(userID string) (*dto.SubscriptionResponse, error)
	ProcessExpiredSubscriptions() (int, error)
	// RemindExpiringSubscriptions nhắc partner gia hạn trước khi gói hết hạn.
	RemindExpiringSubscriptions() (int, error)
	// SetNotificationService để báo cho partner khi gói hết hạn.
	SetNotificationService(notifier *NotificationService)
}

type packageServiceImpl struct {
	repo     repository.PackageRepository
	notifier *NotificationService

	reminderMu   sync.Mutex
	lastReminder map[string]time.Time
}

func NewPackageService(repo repository.PackageRepository) PackageService {
	return &packageServiceImpl{
		repo:         repo,
		lastReminder: make(map[string]time.Time),
	}
}

// validatePackageInput áp quy tắc tối thiểu cho gói bán ra.
func validatePackageInput(input dto.PackageInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("tên gói không được để trống")
	}
	if input.PriceMonthly < 0 {
		return fmt.Errorf("giá gói không được âm")
	}
	if input.MaxSlots < 1 {
		return fmt.Errorf("số slot phải >= 1")
	}
	if input.MaxDurationDays < 1 {
		return fmt.Errorf("thời hạn gói phải >= 1 ngày")
	}
	return nil
}

func (s *packageServiceImpl) GetPackages() ([]models.AdPackage, error) {
	return s.repo.GetActivePackages()
}

func (s *packageServiceImpl) AdminListPackages() ([]models.AdPackage, error) {
	return s.repo.GetAllPackages()
}

func (s *packageServiceImpl) AdminCreatePackage(input dto.PackageInput) (*models.AdPackage, error) {
	if err := validatePackageInput(input); err != nil {
		return nil, err
	}
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}
	pkg := &models.AdPackage{
		ID:                   utils.GenerateUUID(),
		Name:                 strings.TrimSpace(input.Name),
		Description:          input.Description,
		PriceMonthly:         input.PriceMonthly,
		MaxSlots:             input.MaxSlots,
		MaxDurationDays:      input.MaxDurationDays,
		SupportsVideo:        input.SupportsVideo,
		SupportsCarousel:     input.SupportsCarousel,
		HasAdvancedAnalytics: input.HasAdvancedAnalytics,
		SortOrder:            input.SortOrder,
		IsActive:             isActive,
		CreatedAt:            time.Now(),
	}
	if err := s.repo.CreatePackage(pkg); err != nil {
		return nil, err
	}
	return pkg, nil
}

func (s *packageServiceImpl) AdminUpdatePackage(id string, input dto.PackageInput) (*models.AdPackage, error) {
	if err := validatePackageInput(input); err != nil {
		return nil, err
	}
	pkg, err := s.repo.GetPackageByID(id)
	if err != nil {
		return nil, err
	}
	pkg.Name = strings.TrimSpace(input.Name)
	pkg.Description = input.Description
	pkg.PriceMonthly = input.PriceMonthly
	pkg.MaxSlots = input.MaxSlots
	pkg.MaxDurationDays = input.MaxDurationDays
	pkg.SupportsVideo = input.SupportsVideo
	pkg.SupportsCarousel = input.SupportsCarousel
	pkg.HasAdvancedAnalytics = input.HasAdvancedAnalytics
	pkg.SortOrder = input.SortOrder
	if input.IsActive != nil {
		pkg.IsActive = *input.IsActive
	}
	if err := s.repo.UpdatePackage(pkg); err != nil {
		return nil, err
	}
	return pkg, nil
}

// RemindExpiringSubscriptions nhắc mỗi gói sắp hết hạn (<=3 ngày) tối đa 1
// lần/24h. Không tự trừ tiền vì chưa có token hoá thanh toán.
func (s *packageServiceImpl) RemindExpiringSubscriptions() (int, error) {
	if s.notifier == nil {
		return 0, nil
	}
	until := time.Now().AddDate(0, 0, 3)
	subs, err := s.repo.FindExpiringSubscriptions(until)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	sent := 0
	s.reminderMu.Lock()
	defer s.reminderMu.Unlock()
	for _, sub := range subs {
		if last, ok := s.lastReminder[sub.ID]; ok && now.Sub(last) < 24*time.Hour {
			continue
		}
		daysLeft := int(sub.ExpiresAt.Sub(now).Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}
		msg := fmt.Sprintf("Gói quảng cáo của bạn sẽ hết hạn sau %d ngày. Gia hạn để quảng cáo tiếp tục hiển thị.", daysLeft)
		if _, err := s.notifier.Create(context.Background(), sub.UserID, nil, models.NotificationTypePackage, msg, nil, nil, nil); err == nil {
			s.lastReminder[sub.ID] = now
			sent++
		}
	}
	// Dọn mốc cũ để map không phình.
	for id, at := range s.lastReminder {
		if now.Sub(at) > 7*24*time.Hour {
			delete(s.lastReminder, id)
		}
	}
	return sent, nil
}

func (s *packageServiceImpl) SetNotificationService(notifier *NotificationService) {
	s.notifier = notifier
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
		// Chưa có token hoá thanh toán nên không tự trừ tiền; hệ thống chỉ nhắc
		// gia hạn trước khi hết hạn (worker).
		AutoRenew: false,
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
		PackageID:    sub.PackageID,
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
				models.NotificationTypePackage,
				"Gói quảng cáo của bạn đã hết hạn. Quảng cáo sẽ tạm ngưng phân phối cho tới khi bạn gia hạn.",
				nil, nil, nil,
			)
		}
		count++
	}
	return count, nil
}
