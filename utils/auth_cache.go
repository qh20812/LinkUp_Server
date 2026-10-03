package utils

import (
	"sync"
	"time"

	"linkup/models"
)

// Cache kết quả tra cứu users trong AuthMiddleware (status + token_version).
// Mỗi authenticated request tốn 1 PK lookup — với 1000 user poll 60s thì
// cache này cắt ~một nửa tổng DB QPS. TTL ngắn (30s) nên hành vi revoke/ban
// chậm nhất 30s; session revoke vẫn check realtime per-request trong middleware.
// Các chỗ đổi status/bump version (ban, đổi password, logout, reset-all)
// PHẢI gọi InvalidateAuthCache/InvalidateAllAuthCache ngay sau khi ghi DB.

const authCacheTTL = 30 * time.Second

// Giới hạn mềm để map không phình khi bị quét userID lạ.
const authCacheMaxEntries = 50000

type authCacheEntry struct {
	status       models.UserStatus
	tokenVersion int
	expiresAt    time.Time
}

var (
	authCacheMu sync.RWMutex
	authCache   = make(map[string]authCacheEntry)
)

// GetCachedAuth trả về (status, tokenVersion, hit).
func GetCachedAuth(userID string) (models.UserStatus, int, bool) {
	now := time.Now()
	authCacheMu.RLock()
	entry, ok := authCache[userID]
	authCacheMu.RUnlock()
	if !ok || now.After(entry.expiresAt) {
		return "", 0, false
	}
	return entry.status, entry.tokenVersion, true
}

// SetCachedAuth lưu kết quả lookup với TTL, kèm dọn entry hết hạn khi đầy.
func SetCachedAuth(userID string, status models.UserStatus, tokenVersion int) {
	now := time.Now()
	authCacheMu.Lock()
	defer authCacheMu.Unlock()
	if len(authCache) >= authCacheMaxEntries {
		for k, v := range authCache {
			if now.After(v.expiresAt) {
				delete(authCache, k)
			}
		}
	}
	authCache[userID] = authCacheEntry{
		status:       status,
		tokenVersion: tokenVersion,
		expiresAt:    now.Add(authCacheTTL),
	}
}

// InvalidateAuthCache xóa cache của 1 user sau khi đổi status/version.
func InvalidateAuthCache(userID string) {
	authCacheMu.Lock()
	delete(authCache, userID)
	authCacheMu.Unlock()
}

// InvalidateAllAuthCache xóa toàn bộ cache (vd IncrementAllTokenVersions).
func InvalidateAllAuthCache() {
	authCacheMu.Lock()
	authCache = make(map[string]authCacheEntry)
	authCacheMu.Unlock()
}
