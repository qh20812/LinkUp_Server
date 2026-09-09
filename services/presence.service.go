package services

import (
	"context"
	"log"
	"sync"
	"time"

	"linkup/models"
	"linkup/repository"
)

// PresenceService manages real-time online/offline presence for users.
// It uses an in-memory cache for fast lookups and periodically syncs to the database.
type PresenceService struct {
	presenceRepo *repository.PresenceRepository
	settingsRepo *repository.UserSettingsRepository
	friendRepo   *repository.FriendRepository
	chatRepo     *repository.ChatRepository
	cache        map[string]*models.PresenceCacheEntry
	mu           sync.RWMutex
	stopCh       chan struct{}
}

// NewPresenceService creates a new PresenceService and starts the background sync goroutine.
func NewPresenceService(
	presenceRepo *repository.PresenceRepository,
	settingsRepo *repository.UserSettingsRepository,
	friendRepo *repository.FriendRepository,
	chatRepo *repository.ChatRepository,
) *PresenceService {
	s := &PresenceService{
		presenceRepo: presenceRepo,
		settingsRepo: settingsRepo,
		friendRepo:   friendRepo,
		chatRepo:     chatRepo,
		cache:        make(map[string]*models.PresenceCacheEntry),
		stopCh:       make(chan struct{}),
	}
	go s.syncToDatabase()
	return s
}

// Stop stops the background sync goroutine.
func (s *PresenceService) Stop() {
	close(s.stopCh)
}

// MarkOnline sets a user's status to online and updates the last seen timestamp.
func (s *PresenceService) MarkOnline(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	s.cache[userID] = &models.PresenceCacheEntry{
		Status:    models.PresenceStatusOnline,
		LastSeen:  now,
		UpdatedAt: now,
	}
}

// MarkOffline sets a user's status to offline.
func (s *PresenceService) MarkOffline(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if entry, ok := s.cache[userID]; ok {
		entry.Status = models.PresenceStatusOffline
		entry.UpdatedAt = now
	} else {
		s.cache[userID] = &models.PresenceCacheEntry{
			Status:    models.PresenceStatusOffline,
			LastSeen:  now,
			UpdatedAt: now,
		}
	}
}

// UpdateLastSeen updates the last seen timestamp without changing online status.
func (s *PresenceService) UpdateLastSeen(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if entry, ok := s.cache[userID]; ok {
		entry.LastSeen = now
		entry.UpdatedAt = now
	} else {
		s.cache[userID] = &models.PresenceCacheEntry{
			Status:    models.PresenceStatusOffline,
			LastSeen:  now,
			UpdatedAt: now,
		}
	}
}

// IsUserOnline checks if a user is currently online (has active WebSocket connections).
func (s *PresenceService) IsUserOnline(userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.cache[userID]
	if !ok {
		return false
	}
	return entry.Status == models.PresenceStatusOnline
}

// CanUsersSeeEachOther checks if viewerID can see targetUserID's presence.
// Implements the reciprocal rule: both users must have ActivityStatusEnabled.
func (s *PresenceService) CanUsersSeeEachOther(viewerID, targetUserID string) (bool, error) {
	// 1. Get settings for both users
	viewerSettings, err := s.settingsRepo.GetByUserID(context.Background(), viewerID)
	if err != nil {
		return false, err
	}
	targetSettings, err := s.settingsRepo.GetByUserID(context.Background(), targetUserID)
	if err != nil {
		return false, err
	}

	// Apply defaults if no settings exist
	viewerEnabled := true
	targetEnabled := true
	targetVisibility := models.LastSeenVisibilityAllFriends

	if viewerSettings != nil {
		viewerEnabled = viewerSettings.ActivityStatusEnabled
	}
	if targetSettings != nil {
		targetEnabled = targetSettings.ActivityStatusEnabled
		targetVisibility = models.ParseLastSeenVisibility(targetSettings.LastSeenVisibility)
	}

	// 2. Both users must have ActivityStatusEnabled (reciprocal rule)
	if !viewerEnabled || !targetEnabled {
		return false, nil
	}

	// 3. Check visibility based on target's settings
	switch targetVisibility {
	case models.LastSeenVisibilityNobody:
		return false, nil
	case models.LastSeenVisibilityDmOnly:
		hasDM, err := s.chatRepo.FindDirectChat(context.Background(), viewerID, targetUserID)
		if err != nil {
			return false, err
		}
		return hasDM != nil, nil
	case models.LastSeenVisibilityAllFriends:
		pair, err := s.friendRepo.FindPair(context.Background(), viewerID, targetUserID)
		if err != nil {
			return false, err
		}
		// Must be accepted friends
		return pair != nil && pair.Status == models.FriendStatusAccepted, nil
	default:
		return false, nil
	}
}

// GetPresence returns the presence data for a target user as seen by a viewer.
func (s *PresenceService) GetPresence(viewerID, targetUserID string) (*models.UserPresence, error) {
	canSee, err := s.CanUsersSeeEachOther(viewerID, targetUserID)
	if err != nil {
		return nil, err
	}
	if !canSee {
		// Return offline with no last seen info
		return &models.UserPresence{
			UserID:   targetUserID,
			Status:   models.PresenceStatusOffline,
			LastSeen: nil,
		}, nil
	}

	s.mu.RLock()
	entry, ok := s.cache[targetUserID]
	s.mu.RUnlock()

	var lastSeen *time.Time
	if ok {
		lastSeen = &entry.LastSeen
	} else {
		// Fallback to database
		dbLastSeen, err := s.presenceRepo.GetLastSeen(context.Background(), targetUserID)
		if err == nil && dbLastSeen != nil {
			lastSeen = dbLastSeen
		}
	}

	status := models.PresenceStatusOffline
	if ok && entry.Status == models.PresenceStatusOnline {
		status = models.PresenceStatusOnline
	}

	return &models.UserPresence{
		UserID:   targetUserID,
		Status:   status,
		LastSeen: lastSeen,
	}, nil
}

// BatchGetPresence returns presence data for multiple target users.
func (s *PresenceService) BatchGetPresence(viewerID string, targetUserIDs []string) (map[string]*models.UserPresence, error) {
	result := make(map[string]*models.UserPresence, len(targetUserIDs))

	for _, targetID := range targetUserIDs {
		presence, err := s.GetPresence(viewerID, targetID)
		if err != nil {
			// Skip errors, return partial results
			continue
		}
		result[targetID] = presence
	}

	return result, nil
}

// GetOnlineUsers returns a list of user IDs that are currently online.
func (s *PresenceService) GetOnlineUsers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	onlineUsers := make([]string, 0)
	for userID, entry := range s.cache {
		if entry.Status == models.PresenceStatusOnline {
			onlineUsers = append(onlineUsers, userID)
		}
	}
	return onlineUsers
}

// GetOnlineUserCount returns the number of currently online users.
func (s *PresenceService) GetOnlineUserCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, entry := range s.cache {
		if entry.Status == models.PresenceStatusOnline {
			count++
		}
	}
	return count
}

// UpdatePresenceSettings updates a user's presence settings.
func (s *PresenceService) UpdatePresenceSettings(ctx context.Context, userID string, activityStatusEnabled bool, lastSeenVisibility string) error {
	return s.settingsRepo.UpdatePresenceSettings(ctx, userID, activityStatusEnabled, lastSeenVisibility)
}

// GetPresenceSettings returns a user's presence settings.
func (s *PresenceService) GetPresenceSettings(ctx context.Context, userID string) (bool, string, error) {
	settings, err := s.settingsRepo.GetByUserID(ctx, userID)
	if err != nil {
		return true, string(models.LastSeenVisibilityAllFriends), err
	}
	if settings == nil {
		return true, string(models.LastSeenVisibilityAllFriends), nil
	}
	return settings.ActivityStatusEnabled, settings.LastSeenVisibility, nil
}

// syncToDatabase periodically syncs the presence cache to the database.
func (s *PresenceService) syncToDatabase() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.doSync()
		case <-s.stopCh:
			return
		}
	}
}

// doSync performs the actual database sync.
func (s *PresenceService) doSync() {
	s.mu.RLock()
	updates := make(map[string]time.Time)
	for userID, entry := range s.cache {
		// Only sync offline users that haven't been synced recently
		if entry.Status == models.PresenceStatusOffline {
			updates[userID] = entry.LastSeen
		}
	}
	s.mu.RUnlock()

	if len(updates) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.presenceRepo.BatchUpdateLastSeen(ctx, updates); err != nil {
		log.Printf("presence: sync to database failed: %v", err)
	}
}
