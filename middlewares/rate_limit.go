package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	errorsapp "linkup/errors"
)

// RateLimiter is a fixed-window per-key limiter. Zero dependencies by design
// (no x/time in go.mod). Not for distributed use: each app instance tracks
// its own counters, which is sufficient as anti-burst protection.
type RateLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

// maxTrackedKeys bounds memory: when exceeded, expired entries are swept.
const maxTrackedKeys = 10000

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{hits: make(map[string][]time.Time), limit: limit, window: window}
}

// Allow records a hit for key and reports whether it is within the limit.
func (l *RateLimiter) Allow(key string) bool {
	now := time.Now().UTC()
	cutoff := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.hits) > maxTrackedKeys {
		for k, times := range l.hits {
			kept := times[:0]
			for _, t := range times {
				if t.After(cutoff) {
					kept = append(kept, t)
				}
			}
			if len(kept) == 0 {
				delete(l.hits, k)
			} else {
				l.hits[k] = kept
			}
		}
	}

	kept := l.hits[key][:0]
	for _, t := range l.hits[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.limit {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	return true
}

// RateLimit aborts with 429 when a client IP exceeds limit requests per
// window on the wrapped route.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)
	return func(c *gin.Context) {
		if !limiter.Allow(c.ClientIP()) {
			errorsapp.RespondError(c, http.StatusTooManyRequests, errorsapp.New(errorsapp.ErrCodeCommunityCreateThrottled))
			c.Abort()
			return
		}
		c.Next()
	}
}
