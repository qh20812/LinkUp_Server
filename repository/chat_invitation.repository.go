package repository

import (
	"context"
	"errors"
	"fmt"
	"linkup/models"
	"time"

	"gorm.io/gorm"
)

type ChatInvitationRepository struct {
	db *gorm.DB
}

func NewChatInvitationRepository(db *gorm.DB) *ChatInvitationRepository {
	return &ChatInvitationRepository{db: db}
}

func (r *ChatInvitationRepository) Create(ctx context.Context, invite *models.ChatInvite) error {
	return r.db.WithContext(ctx).Create(invite).Error
}

func (r *ChatInvitationRepository) FindPendingByID(ctx context.Context, inviteID string) (*models.ChatInvite, error) {
	var invite models.ChatInvite
	err := r.db.WithContext(ctx).Where("id = ? AND status = ?", inviteID, models.ChatInviteStatusPending).First(&invite).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("không tìm thấy lời mời")
	}
	if err != nil {
		return nil, fmt.Errorf("find invite: %w", err)
	}
	return &invite, nil
}

func (r *ChatInvitationRepository) FindPendingBetween(ctx context.Context, requesterID, targetID string) (*models.ChatInvite, error) {
    var invite models.ChatInvite
    err := r.db.WithContext(ctx).
        Where("((requester_id = ? AND target_id = ?) OR (requester_id = ? AND target_id = ?)) AND status = ?",
            requesterID, targetID, targetID, requesterID, models.ChatInviteStatusPending).
        Order("created_at DESC").
        First(&invite).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("find pending invite: %w", err)
    }
    return &invite, nil
}

func (r *ChatInvitationRepository) FindPendingByTarget(ctx context.Context, targetID string) ([]models.ChatInvite, error) {
	var invites []models.ChatInvite
	err := r.db.WithContext(ctx).
		Where("target_id = ? AND status = ?", targetID, models.ChatInviteStatusPending).
		Order("created_at DESC").
		Find(&invites).Error
	if err != nil {
		return nil, fmt.Errorf("find pending invites by target: %w", err)
	}
	return invites, nil
}

func (r *ChatInvitationRepository) UpdateStatus(ctx context.Context, inviteID string, status models.ChatInviteStatus, chatID *string) error {
	update := map[string]any{"status": status, "updated_at": time.Now().UTC()}
	if chatID != nil {
		update["chat_id"] = chatID
	}
	return r.db.WithContext(ctx).Model(&models.ChatInvite{}).Where("id = ?", inviteID).Updates(update).Error
}

func (r *ChatInvitationRepository) FindActiveBetween(ctx context.Context, userA, userB string) (*models.ChatInvite, error) {
	var invite models.ChatInvite
	err := r.db.WithContext(ctx).
		Where("((requester_id = ? AND target_id = ?) OR (requester_id = ? AND target_id = ?)) AND status IN ?", userA, userB, userB, userA,
			[]models.ChatInviteStatus{models.ChatInviteStatusPending, models.ChatInviteStatusAccepted}).
		Order("created_at DESC").
		First(&invite).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find active invite: %w", err)
	}
	return &invite, nil
}
