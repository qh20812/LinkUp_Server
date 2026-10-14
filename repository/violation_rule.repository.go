package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"linkup/models"

	"gorm.io/gorm"
)

type ViolationRuleRepository struct {
	db *gorm.DB
}

func NewViolationRuleRepository(db *gorm.DB) *ViolationRuleRepository {
	return &ViolationRuleRepository{db: db}
}

func (r *ViolationRuleRepository) Create(ctx context.Context, rule *models.ViolationRule) error {
	if err := r.db.WithContext(ctx).Create(rule).Error; err != nil {
		return fmt.Errorf("insert violation rule: %w", err)
	}
	return nil
}

func (r *ViolationRuleRepository) FindByID(ctx context.Context, id string) (*models.ViolationRule, error) {
	var rule models.ViolationRule
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("violation rule not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("find violation rule: %w", err)
	}
	return &rule, nil
}

func (r *ViolationRuleRepository) FindByIDs(ctx context.Context, ids []string) (map[string]models.ViolationRule, error) {
	result := make(map[string]models.ViolationRule, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var rules []models.ViolationRule
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("find violation rules: %w", err)
	}
	for _, rule := range rules {
		result[rule.ID] = rule
	}
	return result, nil
}

// List trả danh sách rule, mặc định chỉ rule đang bật, sắp xếp theo sort_order.
// activeOnly=false để admin xem cả rule đã tắt. targetType lọc theo phạm vi áp dụng
// (rule "all" luôn được bao gồm).
func (r *ViolationRuleRepository) List(ctx context.Context, activeOnly bool, targetType, keyword string) ([]models.ViolationRule, error) {
	var rules []models.ViolationRule
	q := r.db.WithContext(ctx).Model(&models.ViolationRule{})

	if activeOnly {
		q = q.Where("is_active = ?", true)
	}

	targetType = strings.TrimSpace(strings.ToLower(targetType))
	if targetType != "" {
		q = q.Where("applicable_to IN ?", []string{"all", targetType})
	}

	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("title LIKE ? OR description LIKE ?", like, like)
	}

	if err := q.Order("sort_order ASC, created_at ASC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("list violation rules: %w", err)
	}
	return rules, nil
}

func (r *ViolationRuleRepository) Update(ctx context.Context, rule *models.ViolationRule) error {
	now := time.Now().UTC()
	rule.UpdatedAt = &now
	if err := r.db.WithContext(ctx).Save(rule).Error; err != nil {
		return fmt.Errorf("update violation rule: %w", err)
	}
	return nil
}

func (r *ViolationRuleRepository) SetActive(ctx context.Context, id string, active bool) error {
	updates := map[string]interface{}{
		"is_active":  active,
		"updated_at": time.Now().UTC(),
	}
	tx := r.db.WithContext(ctx).Model(&models.ViolationRule{}).Where("id = ?", id).Updates(updates)
	if tx.Error != nil {
		return fmt.Errorf("update violation rule status: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("violation rule not found: %s", id)
	}
	return nil
}

func (r *ViolationRuleRepository) ExistsByTitle(ctx context.Context, title, excludeID string) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&models.ViolationRule{}).
		Where("LOWER(title) = LOWER(?)", strings.TrimSpace(title))
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, fmt.Errorf("check violation rule title: %w", err)
	}
	return count > 0, nil
}

// CountReportsUsing đếm số report đang dùng rule — dùng để chặn xóa cứng.
func (r *ViolationRuleRepository) CountReportsUsing(ctx context.Context, ruleID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("reports").
		Where("violation_rule_id = ?", ruleID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count reports using rule: %w", err)
	}
	return count, nil
}
