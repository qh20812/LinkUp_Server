package repository

import (
	"testing"
	"time"

	"linkup/models"
)

func feedTestPost(id, author string) models.Post {
	return models.Post{ID: id, UserID: author}
}

func TestPersonalCursor_Roundtrip(t *testing.T) {
	snap := time.Unix(0, 1234567890000000000)
	c := PersonalCursor{
		MainScore: 0.823, MainID: "post-main", MainOK: true,
		SmallNano: 987654321, SmallID: "post-small", SmallOK: true,
		GenScore: 1.5, GenID: "post-gen", GenOK: true,
	}
	enc := c.Encode(snap)
	parsed, parsedSnap, ok := ParsePersonalCursor(enc)
	if !ok {
		t.Fatalf("parse failed for %q", enc)
	}
	if !parsedSnap.Equal(snap) || parsed != c {
		t.Fatalf("roundtrip mismatch: got %+v snap %v", parsed, parsedSnap)
	}
}

func TestPersonalCursor_InvalidFallsBack(t *testing.T) {
	for _, raw := range []string{"", "123_0.5_x", "v2_broken", "v1cursor_x"} {
		if _, _, ok := ParsePersonalCursor(raw); ok {
			t.Fatalf("expected parse failure for %q", raw)
		}
	}
}

func TestMergeFeedStreams_ExploreSlots(t *testing.T) {
	main := []models.Post{
		feedTestPost("m1", "a1"), feedTestPost("m2", "a2"), feedTestPost("m3", "a3"),
		feedTestPost("m4", "a4"), feedTestPost("m5", "a5"), feedTestPost("m6", "a6"),
		feedTestPost("m7", "a7"), feedTestPost("m8", "a8"),
	}
	small := []models.Post{feedTestPost("s1", "b1"), feedTestPost("s2", "b2")}
	general := []models.Post{feedTestPost("g1", "c1"), feedTestPost("g2", "c2")}

	merged, advM, advS, advG, _, _, _ := MergeFeedStreams(main, small, general, 8, 2, 2, 10, 5, 2)

	if len(merged) != 10 {
		t.Fatalf("expected 10 items, got %d", len(merged))
	}
	// Slot explore là vị trí thứ 5 và 10 (index 4, 9), luân phiên small/general.
	if merged[4].ID != "s1" {
		t.Fatalf("slot 5: expected s1, got %s", merged[4].ID)
	}
	if merged[9].ID != "g1" {
		t.Fatalf("slot 10: expected g1, got %s", merged[9].ID)
	}
	if advM != 8 || advS != 1 || advG != 1 {
		t.Fatalf("advance mismatch: m=%d s=%d g=%d", advM, advS, advG)
	}
}

func TestMergeFeedStreams_DiversityCap(t *testing.T) {
	main := []models.Post{
		feedTestPost("m1", "same"), feedTestPost("m2", "same"), feedTestPost("m3", "same"),
		feedTestPost("m4", "other"),
	}

	merged, _, _, _, _, _, _ := MergeFeedStreams(main, nil, nil, 4, 0, 0, 4, 5, 2)

	count := 0
	for _, p := range merged {
		if p.UserID == "same" {
			count++
		}
	}
	if count > 2 {
		t.Fatalf("diversity cap violated: author 'same' appears %d times in %v", count, merged)
	}
}

func TestMergeFeedStreams_FillsWhenExploreEmpty(t *testing.T) {
	main := []models.Post{
		feedTestPost("m1", "a1"), feedTestPost("m2", "a2"), feedTestPost("m3", "a3"),
		feedTestPost("m4", "a4"), feedTestPost("m5", "a5"),
	}

	merged, _, _, _, _, _, _ := MergeFeedStreams(main, nil, nil, 5, 0, 0, 5, 5, 2)

	if len(merged) != 5 {
		t.Fatalf("expected main to fill explore slots, got %d items", len(merged))
	}
}
