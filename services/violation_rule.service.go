package services

import (
	"context"
	"strings"
	"time"

	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
)

type ViolationRuleService struct {
	ruleRepo   *repository.ViolationRuleRepository
	authRepo   *repository.AuthRepository
	validation *validations.ViolationRuleValidation
}

func NewViolationRuleService(
	ruleRepo *repository.ViolationRuleRepository,
	authRepo *repository.AuthRepository,
	validation *validations.ViolationRuleValidation,
) *ViolationRuleService {
	return &ViolationRuleService{
		ruleRepo:   ruleRepo,
		authRepo:   authRepo,
		validation: validation,
	}
}

func (s *ViolationRuleService) ensureSuperAdmin(ctx context.Context, userID string) error {
	isSuperAdmin, err := s.authRepo.HasRole(ctx, userID, models.RoleSuperAdmin)
	if err != nil {
		return err
	}
	if !isSuperAdmin {
		return errorsapp.New(errorsapp.ErrCodeAdminNotSuperadmin)
	}
	return nil
}

// ListRules: user thường xem rule đang bật (lọc theo target để đổ dropdown report);
// admin truyền includeInactive=true để xem tất cả.
func (s *ViolationRuleService) ListRules(ctx context.Context, targetType, keyword string, includeInactive bool) ([]models.ViolationRule, error) {
	return s.ruleRepo.List(ctx, !includeInactive, targetType, keyword)
}

func (s *ViolationRuleService) GetRule(ctx context.Context, ruleID string) (*models.ViolationRule, error) {
	rule, err := s.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeViolationRuleNotFound)
	}
	return rule, nil
}

func (s *ViolationRuleService) CreateRule(ctx context.Context, adminID, title, description, applicableTo, severity string, sortOrder int) (*models.ViolationRule, error) {
	if err := s.ensureSuperAdmin(ctx, adminID); err != nil {
		return nil, err
	}
	if err := s.validation.ValidateTitle(title); err != nil {
		return nil, err
	}
	if err := s.validation.ValidateDescription(description); err != nil {
		return nil, err
	}

	exists, err := s.ruleRepo.ExistsByTitle(ctx, title, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errorsapp.New(errorsapp.ErrCodeViolationRuleTitleDup)
	}

	now := time.Now().UTC()
	rule := &models.ViolationRule{
		ID:           utils.GenerateUUID(),
		Title:        strings.TrimSpace(title),
		Description:  strings.TrimSpace(description),
		ApplicableTo: models.ParseViolationApplicable(applicableTo),
		Severity:     models.ParseViolationSeverity(severity),
		SortOrder:    sortOrder,
		IsActive:     true,
		CreatedAt:    now,
	}
	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *ViolationRuleService) UpdateRule(ctx context.Context, adminID, ruleID, title, description, applicableTo, severity string, sortOrder *int) (*models.ViolationRule, error) {
	if err := s.ensureSuperAdmin(ctx, adminID); err != nil {
		return nil, err
	}

	rule, err := s.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeViolationRuleNotFound)
	}

	if title != "" {
		if err := s.validation.ValidateTitle(title); err != nil {
			return nil, err
		}
		exists, err := s.ruleRepo.ExistsByTitle(ctx, title, ruleID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errorsapp.New(errorsapp.ErrCodeViolationRuleTitleDup)
		}
		rule.Title = strings.TrimSpace(title)
	}
	if description != "" {
		if err := s.validation.ValidateDescription(description); err != nil {
			return nil, err
		}
		rule.Description = strings.TrimSpace(description)
	}
	if applicableTo != "" {
		rule.ApplicableTo = models.ParseViolationApplicable(applicableTo)
	}
	if severity != "" {
		rule.Severity = models.ParseViolationSeverity(severity)
	}
	if sortOrder != nil {
		if *sortOrder < 0 {
			*sortOrder = 0
		}
		rule.SortOrder = *sortOrder
	}

	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

// SetActive: tắt/mở rule (ẩn mềm). Tắt rule không ảnh hưởng report cũ.
func (s *ViolationRuleService) SetActive(ctx context.Context, adminID, ruleID string, active bool) error {
	if err := s.ensureSuperAdmin(ctx, adminID); err != nil {
		return err
	}
	if _, err := s.ruleRepo.FindByID(ctx, ruleID); err != nil {
		return errorsapp.New(errorsapp.ErrCodeViolationRuleNotFound)
	}
	return s.ruleRepo.SetActive(ctx, ruleID, active)
}

// resolveActiveRule kiểm tra rule dùng cho report: tồn tại + đang bật + khớp target.
func (s *ViolationRuleService) resolveActiveRule(ctx context.Context, ruleID, targetType string) (*models.ViolationRule, error) {
	rule, err := s.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeViolationRuleNotFound)
	}
	if !rule.IsActive {
		return nil, errorsapp.New(errorsapp.ErrCodeViolationRuleInactive)
	}
	if !rule.MatchesTarget(targetType) {
		return nil, errorsapp.New(errorsapp.ErrCodeViolationRuleTargetMismatch)
	}
	return rule, nil
}

// ValidateForReport dùng chung cho ReportService (giữ optional: nil = bỏ qua).
func (s *ViolationRuleService) ValidateForReport(ctx context.Context, ruleID *string, targetType string) error {
	if ruleID == nil || strings.TrimSpace(*ruleID) == "" {
		return nil
	}
	_, err := s.resolveActiveRule(ctx, strings.TrimSpace(*ruleID), targetType)
	return err
}

// DeleteRule: chỉ cho xóa cứng khi chưa có report nào dùng; ngược lại báo lỗi
// yêu cầu tắt mềm. (Service không có repo reports nên nhận count qua tham số
// — caller (controller route admin) có thể bỏ qua bước này nếu đã có guard riêng.)
func (s *ViolationRuleService) EnsureDeletable(reportsUsing int64) error {
	if reportsUsing > 0 {
		return errorsapp.New(errorsapp.ErrCodeViolationRuleInUse)
	}
	return nil
}
