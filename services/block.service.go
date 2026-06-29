package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
)

type BlockService struct {
	blockRepo  *repository.BlockRepository
	authRepo   *repository.AuthRepository
	validation *validations.BlockValidation
}

func NewBlockService(blockRepo *repository.BlockRepository, authRepo *repository.AuthRepository, validation *validations.BlockValidation) *BlockService {
	return &BlockService{
		blockRepo:  blockRepo,
		authRepo:   authRepo,
		validation: validation,
	}
}

func (s *BlockService) ToggleBlock(ctx context.Context, userID, targetUserID string) (dto.BlockUserResponse, error) {
	if err := s.validation.ValidateToggleBlock(userID, targetUserID); err != nil {
		return dto.BlockUserResponse{}, err
	}

	isAdmin, err := s.authRepo.HasRole(ctx, targetUserID, models.RoleAdmin)
	if err != nil {
		return dto.BlockUserResponse{}, fmt.Errorf("toggle block: %w", err)
	}
	isSuperAdmin, err := s.authRepo.HasRole(ctx, targetUserID, models.RoleSuperAdmin)
	if err != nil {
		return dto.BlockUserResponse{}, fmt.Errorf("toggle block: %w", err)
	}
	if isAdmin || isSuperAdmin {
		return dto.BlockUserResponse{}, errors.New("không thể chặn quản trị viên hoặc siêu quản trị viên")
	}

	existing, err := s.blockRepo.FindByUserAndTarget(ctx, userID, targetUserID)
	if err != nil {
		return dto.BlockUserResponse{}, fmt.Errorf("toggle block: %w", err)
	}

	if existing != nil {
		if err := s.blockRepo.Delete(ctx, existing.ID); err != nil {
			return dto.BlockUserResponse{}, fmt.Errorf("toggle block: %w", err)
		}
		return dto.BlockUserResponse{
			Status:  "unblocked",
			Message: "Đã bỏ chặn người dùng thành công",
		}, nil
	}

	block := models.Block{
		ID:            utils.GenerateUUID(),
		UserID:        userID,
		BlockedUserID: targetUserID,
		CreatedAt:     time.Now().UTC(),
	}
	if err := s.blockRepo.Create(ctx, &block); err != nil {
		return dto.BlockUserResponse{}, fmt.Errorf("toggle block: %w", err)
	}

	return dto.BlockUserResponse{
		Status:  "blocked",
		Message: "Đã chặn người dùng thành công",
	}, nil
}

func (s *BlockService) GetBlockedUsers(ctx context.Context, userID string) ([]dto.BlockedUserResponse, error) {
	blocks, err := s.blockRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get blocked users: %w", err)
	}

	result := make([]dto.BlockedUserResponse, len(blocks))
	for i, b := range blocks {
		result[i] = dto.BlockedUserResponse{
			BlockedUserID: b.BlockedUserID,
			BlockedAt:     b.CreatedAt,
		}
	}
	return result, nil
}
