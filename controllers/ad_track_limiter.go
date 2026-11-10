package controllers

import (
	"sync"
	"time"
)

// trackLimiter giới hạn số event tracking quảng cáo trên mỗi key (user/IP)
// trong một cửa sổ trượt. Server chạy 1 instance nên in-memory là đủ; nếu
// scale nhiều instance cần chuyển sang Redis.
type trackLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

func newTrackLimiter(limit int, window time.Duration) *trackLimiter {
	return &trackLimiter{
		hits:   make(map[string][]time.Time),
		limit:  limit,
		window: window,
	}
}

// allow ghi nhận một lượt gọi nếu trong window chưa vượt limit.
// Key rỗng luôn được cho qua (không có user/IP để giới hạn).
func (l *trackLimiter) allow(key string, now time.Time) bool {
	if key == "" {
		return true
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.window)
	existing := l.hits[key]
	recent := make([]time.Time, 0, len(existing)+1)
	for _, t := range existing {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= l.limit {
		// Giữ lại danh sách đã prune để map không phình theo thời gian.
		l.hits[key] = recent
		return false
	}

	l.hits[key] = append(recent, now)
	return true
}

// adTrackLimiter: 60 event tracking/phút cho mỗi user (hoặc IP nếu thiếu user).
var adTrackLimiter = newTrackLimiter(60, time.Minute)
