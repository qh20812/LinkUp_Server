package models

import "testing"

func TestAdStatusConsumesSlot(t *testing.T) {
	tests := []struct {
		status AdStatus
		want   bool
	}{
		{AdStatusPending, true},
		{AdStatusActive, true},
		{AdStatusPaused, true},
		{AdStatusRejected, true},
		{AdStatusCompleted, false},
		{AdStatusSuspended, false},
		{AdStatus("unknown"), false},
	}

	for _, tc := range tests {
		if got := tc.status.ConsumesSlot(); got != tc.want {
			t.Errorf("%s.ConsumesSlot() = %v, want %v", tc.status, got, tc.want)
		}
	}
}
