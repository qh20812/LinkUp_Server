package controllers

import (
	"testing"
	"time"
)

func TestTrackLimiter_AllowsUpToLimitThenBlocks(t *testing.T) {
	l := newTrackLimiter(3, time.Minute)
	now := time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)

	for i := 0; i < 3; i++ {
		if !l.allow("user-1", now) {
			t.Fatalf("lượt %d phải được phép", i+1)
		}
	}
	if l.allow("user-1", now) {
		t.Fatalf("lượt thứ 4 phải bị chặn")
	}
}

func TestTrackLimiter_WindowSlides(t *testing.T) {
	l := newTrackLimiter(2, time.Minute)
	start := time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)

	if !l.allow("user-1", start) || !l.allow("user-1", start.Add(10*time.Second)) {
		t.Fatalf("2 lượt đầu phải được phép")
	}
	if l.allow("user-1", start.Add(20*time.Second)) {
		t.Fatalf("lượt thứ 3 trong window phải bị chặn")
	}
	// Sau khi cửa sổ trượt qua lượt đầu, còn 1 lượt được phép.
	if !l.allow("user-1", start.Add(61*time.Second)) {
		t.Fatalf("sau window phải được phép lại")
	}
}

func TestTrackLimiter_SeparateKeysAndEmptyKey(t *testing.T) {
	l := newTrackLimiter(1, time.Minute)
	now := time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)

	if !l.allow("user-1", now) {
		t.Fatalf("user-1 lượt đầu phải được phép")
	}
	if l.allow("user-1", now) {
		t.Fatalf("user-1 lượt hai phải bị chặn")
	}
	if !l.allow("user-2", now) {
		t.Fatalf("key khác phải có hạn mức riêng")
	}
	for i := 0; i < 5; i++ {
		if !l.allow("", now) {
			t.Fatalf("key rỗng phải luôn được phép")
		}
	}
}
