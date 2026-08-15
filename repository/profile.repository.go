package repository

import (
	"context"
	"fmt"
	"time"

	"linkup/models"

	"gorm.io/gorm"
)

// MaxFindByIDs is the maximum number of user IDs accepted in a single
// FindByIDs call. Caller must split larger batches into chunks.
const MaxFindByIDs = 100

type ProfileRepository struct {
    db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
    return &ProfileRepository{db: db}
}

func (r *ProfileRepository) Create(ctx context.Context, profile *models.Profile) (*models.Profile, error) {
    tx := r.db.WithContext(ctx).Create(profile)
    if tx.Error != nil {
        return nil, fmt.Errorf("insert profile: %w", tx.Error)
    }
    return profile, nil
}

// FindByUserID returns the profile for the given user ID.
// Returns (nil, nil) when no profile exists — consistent with FindByPhoneNumber.
// Callers must check both err and the returned profile for nil.
func (r *ProfileRepository) FindByUserID(ctx context.Context, userID string) (*models.Profile, error) {
    var profile models.Profile
    tx := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile)
    if tx.Error != nil {
        if tx.Error == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, fmt.Errorf("find profile: %w", tx.Error)
    }
    return &profile, nil
}

func (r *ProfileRepository) FindByPhoneNumber(ctx context.Context, phoneNumber string, excludeUserID string) (*models.Profile, error) {
    var profile models.Profile
    tx := r.db.WithContext(ctx).
        Where("phone_number = ? AND user_id != ?", phoneNumber, excludeUserID).
        First(&profile)
    if tx.Error != nil {
        if tx.Error == gorm.ErrRecordNotFound {
            return nil, nil // Không tìm thấy = OK
        }
        return nil, fmt.Errorf("find by phone: %w", tx.Error)
    }
    return &profile, nil
}

func (r *ProfileRepository) Update(ctx context.Context, userID string, profile *models.Profile) (*models.Profile, error) {
	tx := r.db.WithContext(ctx).Where("user_id = ?", userID).Updates(profile)
	if tx.Error != nil {
		return nil, fmt.Errorf("update profile: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return nil, fmt.Errorf("profile not found")
	}
	return profile, nil
}

// ProfileView is an enriched profile that includes data from the users and posts tables.
type ProfileView struct {
 models.Profile
 Username  string    `json:"username"`
 PostCount int64     `json:"post_count"`
 CreatedAt time.Time `json:"created_at"`
}

// FindEnrichedByUserID returns an enriched profile with username, post_count, and created_at.
// Returns (nil, nil) when no profile exists.
func (r *ProfileRepository) FindEnrichedByUserID(ctx context.Context, userID string) (*ProfileView, error) {
 var pv ProfileView
 tx := r.db.WithContext(ctx).
   Raw(`SELECT p.*, u.username,
       (SELECT COUNT(*) FROM posts WHERE user_id = p.user_id AND status != 'deleted') AS post_count,
       u.created_at
       FROM profiles p
       JOIN users u ON u.id = p.user_id
       WHERE p.user_id = ?`, userID).
   Scan(&pv)
 if tx.Error != nil {
   if tx.Error == gorm.ErrRecordNotFound {
     return nil, nil
   }
   return nil, fmt.Errorf("find enriched profile: %w", tx.Error)
 }
 if pv.UserID == "" {
   return nil, nil
 }
 return &pv, nil
}

// FindByIDs returns profiles for the given user IDs.
// Deduplicates input and queries in chunks of MaxFindByIDs to handle
// arbitrarily large input without silent truncation.
// Returns an empty slice (not nil) when no profiles match.
func (r *ProfileRepository) FindByIDs(ctx context.Context, userIDs []string) ([]models.Profile, error) {
	if len(userIDs) == 0 {
		return []models.Profile{}, nil
	}

	seen := make(map[string]struct{}, len(userIDs))
	unique := make([]string, 0, len(userIDs))
	for _, id := range userIDs {
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	// Query in chunks of MaxFindByIDs to avoid oversized IN clauses.
	var allProfiles []models.Profile
	for i := 0; i < len(unique); i += MaxFindByIDs {
		end := i + MaxFindByIDs
		if end > len(unique) {
			end = len(unique)
		}
		chunk := unique[i:end]

		var profiles []models.Profile
		tx := r.db.WithContext(ctx).
			Where("user_id IN ?", chunk).
			Find(&profiles)
		if tx.Error != nil {
			return nil, fmt.Errorf("find profiles by ids: %w", tx.Error)
		}
		allProfiles = append(allProfiles, profiles...)
	}

	return allProfiles, nil
}