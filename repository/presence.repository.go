package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type PresenceRepository struct {
	db *gorm.DB
}

func NewPresenceRepository(db *gorm.DB) *PresenceRepository {
	return &PresenceRepository{db: db}
}

// UpdateLastSeen updates the last_seen timestamp for a user.
func (r *PresenceRepository) UpdateLastSeen(ctx context.Context, userID string, lastSeen time.Time) error {
	tx := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("users").
		Where("id = ?", userID).
		Update("last_seen", lastSeen)
	if tx.Error != nil {
		return fmt.Errorf("update last seen: %w", tx.Error)
	}
	return nil
}

// BatchUpdateLastSeen updates last_seen for multiple users in one query.
func (r *PresenceRepository) BatchUpdateLastSeen(ctx context.Context, updates map[string]time.Time) error {
	if len(updates) == 0 {
		return nil
	}

	// Build CASE WHEN statement for batch update
	caseSQL := "CASE id "
	values := make([]interface{}, 0, len(updates)*2)
	for userID, lastSeen := range updates {
		caseSQL += "WHEN ? THEN ? "
		values = append(values, userID, lastSeen)
	}
	caseSQL += "END"

	tx := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("users").
		Where("id IN ?", keys(updates)).
		Update("last_seen", gorm.Expr(caseSQL, values...))
	if tx.Error != nil {
		return fmt.Errorf("batch update last seen: %w", tx.Error)
	}
	return nil
}

// GetLastSeen retrieves the last_seen timestamp for a user.
func (r *PresenceRepository) GetLastSeen(ctx context.Context, userID string) (*time.Time, error) {
	var result struct {
		LastSeen *time.Time
	}
	tx := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("users").
		Select("last_seen").
		Where("id = ?", userID).
		Scan(&result)
	if tx.Error != nil {
		return nil, fmt.Errorf("get last seen: %w", tx.Error)
	}
	return result.LastSeen, nil
}

// BatchGetLastSeen retrieves last_seen for multiple users.
func (r *PresenceRepository) BatchGetLastSeen(ctx context.Context, userIDs []string) (map[string]*time.Time, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	var results []struct {
		ID       string
		LastSeen *time.Time
	}
	tx := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("users").
		Select("id, last_seen").
		Where("id IN ?", userIDs).
		Scan(&results)
	if tx.Error != nil {
		return nil, fmt.Errorf("batch get last seen: %w", tx.Error)
	}

	lastSeenMap := make(map[string]*time.Time, len(results))
	for _, r := range results {
		lastSeenMap[r.ID] = r.LastSeen
	}
	return lastSeenMap, nil
}

// keys extracts the keys from a map.
func keys(m map[string]time.Time) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
