package repository

import (
	"context"
	"linkup/models"

	"gorm.io/gorm"
)

type CommunityRuleRepository struct {
	db *gorm.DB
}

func NewCommunityRuleRepository(db *gorm.DB) *CommunityRuleRepository {
	return &CommunityRuleRepository{db: db}
}

func (r *CommunityRuleRepository) Create(ctx context.Context, rule *models.CommunityRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *CommunityRuleRepository) FindByID(ctx context.Context, id string) (*models.CommunityRule, error) {
	var rule models.CommunityRule
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *CommunityRuleRepository) FindByCommunityID(ctx context.Context, communityID string) ([]models.CommunityRule, error) {
	var rules []models.CommunityRule
	err := r.db.WithContext(ctx).
		Where("community_id = ?", communityID).
		Order("category, position").
		Find(&rules).Error
	return rules, err
}

func (r *CommunityRuleRepository) Update(ctx context.Context, rule *models.CommunityRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *CommunityRuleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.CommunityRule{}, "id = ?", id).Error
}

func (r *CommunityRuleRepository) GetMaxPosition(ctx context.Context, communityID string, category models.RuleCategory) (int, error) {
	var maxPos struct {
		Max int
	}
	err := r.db.WithContext(ctx).
		Model(&models.CommunityRule{}).
		Select("COALESCE(MAX(position), 0) as max").
		Where("community_id = ? AND category = ?", communityID, category).
		Scan(&maxPos).Error
	return maxPos.Max, err
}
