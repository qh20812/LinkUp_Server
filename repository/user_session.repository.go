package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"linkup/models"

	"gorm.io/gorm"
)

type UserSessionRepository struct {
	db *gorm.DB
}

func NewUserSessionRepository(db *gorm.DB) *UserSessionRepository {
	return &UserSessionRepository{db: db}
}

func (r *UserSessionRepository) Create(ctx context.Context, session *models.UserSession) error {
	tx := r.db.WithContext(ctx).Create(session)
	if tx.Error != nil {
		return fmt.Errorf("create user session: %w", tx.Error)
	}
	return nil
}

// FindActiveByID returns a non-revoked, non-expired session, or (nil, nil).
func (r *UserSessionRepository) FindActiveByID(ctx context.Context, sessionID, userID string) (*models.UserSession, error) {
	var session models.UserSession
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL AND expires_at > ?", sessionID, userID, time.Now().UTC()).
		First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user session: %w", err)
	}
	return &session, nil
}

func (r *UserSessionRepository) ListActiveByUserID(ctx context.Context, userID string) ([]models.UserSession, error) {
	var sessions []models.UserSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now().UTC()).
		Order("last_active_at DESC").
		Find(&sessions).Error
	if err != nil {
		return nil, fmt.Errorf("list user sessions: %w", err)
	}
	return sessions, nil
}

// Revoke marks a session as revoked. Returns (false, nil) when no active
// session matched (nothing to revoke).
func (r *UserSessionRepository) Revoke(ctx context.Context, sessionID, userID string) (bool, error) {
	now := time.Now().UTC()
	tx := r.db.WithContext(ctx).
		Model(&models.UserSession{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", sessionID, userID).
		Update("revoked_at", now)
	if tx.Error != nil {
		return false, fmt.Errorf("revoke user session: %w", tx.Error)
	}
	return tx.RowsAffected > 0, nil
}

func (r *UserSessionRepository) RevokeAllExcept(ctx context.Context, userID, keepSessionID string) error {
	now := time.Now().UTC()
	tx := r.db.WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ? AND id != ? AND revoked_at IS NULL", userID, keepSessionID).
		Update("revoked_at", now)
	if tx.Error != nil {
		return fmt.Errorf("revoke other user sessions: %w", tx.Error)
	}
	return nil
}

func (r *UserSessionRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	tx := r.db.WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now)
	if tx.Error != nil {
		return fmt.Errorf("revoke all user sessions: %w", tx.Error)
	}
	return nil
}

func (r *UserSessionRepository) CleanupExpired(ctx context.Context) error {
	tx := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&models.UserSession{})
	if tx.Error != nil {
		return fmt.Errorf("cleanup expired user sessions: %w", tx.Error)
	}
	return nil
}
