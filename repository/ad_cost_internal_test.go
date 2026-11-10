package repository

import (
	"testing"
	"time"

	"linkup/models"
)

func TestActionCost(t *testing.T) {
	ad := &models.Ad{CPMPrice: 25000, CPCPrice: 1200}

	tests := []struct {
		action string
		want   float64
	}{
		{string(models.ActionImpression), 25},
		{string(models.ActionView), 25},
		{string(models.ActionClick), 1200},
		{string(models.ActionSwipe), 0},
		{string(models.ActionVideoStart), 0},
	}

	for _, tc := range tests {
		if got := actionCost(ad, tc.action); got != tc.want {
			t.Fatalf("actionCost(%q) = %v, want %v", tc.action, got, tc.want)
		}
	}
}

func TestStartOfDayKeepsLocation(t *testing.T) {
	loc := time.FixedZone("ICT", 7*3600)
	in := time.Date(2026, 11, 7, 15, 30, 0, 0, loc)

	got := startOfDay(in)
	want := time.Date(2026, 11, 7, 0, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Fatalf("startOfDay = %v, want %v", got, want)
	}
	if got.Location() != loc {
		t.Fatalf("startOfDay phải giữ nguyên timezone")
	}
}

func TestActionDedupeWindow(t *testing.T) {
	tests := []struct {
		action string
		want   time.Duration
	}{
		{string(models.ActionImpression), 30 * time.Minute},
		{string(models.ActionClick), 5 * time.Minute},
		{string(models.ActionView), 0},
		{string(models.ActionSwipe), 0},
		{string(models.ActionVideoStart), 0},
		{string(models.ActionVideoEnd), 0},
	}
	for _, tc := range tests {
		if got := actionDedupeWindow(tc.action); got != tc.want {
			t.Fatalf("actionDedupeWindow(%q) = %v, want %v", tc.action, got, tc.want)
		}
	}
}
