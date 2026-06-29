package services

import (
	"context"
	"errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"time"
)

type CommunityRuleService struct {
	ruleRepo   *repository.CommunityRuleRepository
	validation *validations.CommunityRuleValidation
}

func NewCommunityRuleService(ruleRepo *repository.CommunityRuleRepository, validation *validations.CommunityRuleValidation) *CommunityRuleService {
	return &CommunityRuleService{ruleRepo: ruleRepo, validation: validation}
}

func (s *CommunityRuleService) CreateRule(ctx context.Context, communityID string, category models.RuleCategory, title, content string, position *int) (*models.CommunityRule, error) {
	pos := 0
	if position != nil {
		pos = *position
	}
	if err := s.validation.ValidateCreateRule(category, title, content, pos); err != nil {
		return nil, err
	}

	if position == nil {
		maxPos, err := s.ruleRepo.GetMaxPosition(ctx, communityID, category)
		if err != nil {
			return nil, errors.New("lỗi khi xác định vị trí nội quy")
		}
		pos = maxPos + 1
	}

	now := time.Now().UTC()
	rule := models.NewCommunityRule(communityID, category, title, content, pos)
	rule.ID = utils.GenerateUUID()
	rule.CreatedAt = now

	if err := s.ruleRepo.Create(ctx, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *CommunityRuleService) UpdateRule(ctx context.Context, ruleID, title, content string, category *models.RuleCategory, position *int) (*models.CommunityRule, error) {
	existing, err := s.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return nil, errors.New("nội quy không tồn tại")
	}

	if title != "" {
		if err := s.validation.ValidateTitle(title); err != nil {
			return nil, err
		}
		existing.Title = title
	}
	if content != "" {
		if err := s.validation.ValidateContent(content); err != nil {
			return nil, err
		}
		existing.Content = content
	}
	if category != nil {
		if err := s.validation.ValidateCategory(*category); err != nil {
			return nil, err
		}
		existing.Category = *category
	}
	if position != nil {
		if *position < 0 {
			return nil, validations.ErrRulePositionNegative
		}
		existing.Position = *position
	}

	now := time.Now().UTC()
	existing.UpdatedAt = &now

	if err := s.ruleRepo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *CommunityRuleService) DeleteRule(ctx context.Context, ruleID string) error {
	_, err := s.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return errors.New("nội quy không tồn tại")
	}
	return s.ruleRepo.Delete(ctx, ruleID)
}

func (s *CommunityRuleService) GetRulesByCommunity(ctx context.Context, communityID string) ([]models.CommunityRule, error) {
	return s.ruleRepo.FindByCommunityID(ctx, communityID)
}
