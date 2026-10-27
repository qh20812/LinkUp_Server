package repository

import (
	"strings"
	"testing"
	"time"

	"linkup/config"
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

func TestPersonalCursor_DoneRoundtrip(t *testing.T) {
	snap := time.Unix(0, 1234567890000000000)
	c := PersonalCursor{
		MainScore: 0.5, MainID: "post-main", MainOK: true,
		SmallDone: true,
		GenDone:   true,
	}
	enc := c.Encode(snap)
	parsed, _, ok := ParsePersonalCursor(enc)
	if !ok {
		t.Fatalf("parse failed for %q", enc)
	}
	if parsed != c {
		t.Fatalf("roundtrip mismatch: got %+v want %+v (raw %q)", parsed, c, enc)
	}
	if !parsed.SmallDone || !parsed.GenDone || parsed.SmallOK || parsed.GenOK {
		t.Fatalf("done flags not preserved: %+v", parsed)
	}
}

func TestMergeFeedStreams_DedupesAcrossStreams(t *testing.T) {
	// main và explore có thể trả cùng một bài; trang merge không được lặp.
	main := []models.Post{
		feedTestPost("p1", "a1"), feedTestPost("p2", "a2"), feedTestPost("p3", "a3"),
		feedTestPost("p4", "a4"), feedTestPost("p5", "a5"), feedTestPost("p6", "a6"),
	}
	small := []models.Post{feedTestPost("p1", "a1"), feedTestPost("s2", "b2")}
	general := []models.Post{feedTestPost("p2", "a2"), feedTestPost("g2", "c2")}

	merged, _, _, _, _, _, _ := MergeFeedStreams(main, small, general, 6, 2, 2, 7, 1, 2)

	seen := map[string]bool{}
	for _, p := range merged {
		if seen[p.ID] {
			t.Fatalf("duplicate post %s in merged page", p.ID)
		}
		seen[p.ID] = true
	}
	if len(merged) != 7 {
		t.Fatalf("expected a full page of 7 unique items, got %d", len(merged))
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

func TestBoostCaseExpr(t *testing.T) {
	age := ageHoursExpr

	repost := boostCaseExpr("posts.shared_from_post_id IS NULL", age, 6, 0.5)
	if !strings.Contains(repost, "posts.shared_from_post_id IS NULL AND ") {
		t.Fatalf("expected repost guard before age condition, got %s", repost)
	}
	if strings.Count(repost, age) != 2 {
		t.Fatalf("expected age expression exactly twice, got %s", repost)
	}
	if !strings.Contains(repost, "0.500000 * (1 - ") {
		t.Fatalf("expected linear decay formula, got %s", repost)
	}

	plain := boostCaseExpr("", age, 6, 0.5)
	if !strings.HasPrefix(plain, "CASE WHEN (") {
		t.Fatalf("expected no dangling AND without extraCond, got %s", plain)
	}
	if strings.Contains(plain, "AND (") {
		t.Fatalf("unexpected AND in plain boost expression: %s", plain)
	}
}

// TestAgeHoursExpr chặn tái phát lỗi: từng có fmt.Sprintf(ageH, "?") sinh
// "%!(EXTRA string=?)" lọt vào SQL và MySQL trả 1064.
func TestAgeHoursExpr(t *testing.T) {
	if strings.Contains(ageHoursExpr, "%") {
		t.Fatalf("ageHoursExpr must not contain fmt verbs: %s", ageHoursExpr)
	}
	if strings.Count(ageHoursExpr, "?") != 1 {
		t.Fatalf("expected exactly one placeholder in ageHoursExpr, got %s", ageHoursExpr)
	}
}

// TestMainScoreExpr_PlaceholderCount chặn lỗi lệch placeholder/args — đúng
// lớp lỗi từng làm feed v2 trả 500 (%!(EXTRA string=?)).
func TestMainScoreExpr_PlaceholderCount(t *testing.T) {
	terms := mainScoreTerms(config.DefaultFeedWeights, config.DefaultFeedRankV2)
	expr := joinScoreTerms(terms)
	args := mainScoreArgs(time.Now(), "user-1")
	if got, want := strings.Count(expr, "?"), len(args); got != want {
		t.Fatalf("placeholder mismatch: %d placeholders, %d args\nexpr=%s", got, want, expr)
	}
}

func TestMainScoreTerms_HaveDebugColumns(t *testing.T) {
	terms := mainScoreTerms(config.DefaultFeedWeights, config.DefaultFeedRankV2)
	want := []string{
		"recency", "engagement", "affinity", "interest", "new_boost",
		"small_bonus", "seen_penalty", "seen_penalty_24h", "hidden_author_penalty",
	}
	if len(terms) != len(want) {
		t.Fatalf("term count: got %d want %d", len(terms), len(want))
	}
	for i, name := range want {
		if terms[i].Name != name {
			t.Fatalf("term[%d]: got %q want %q", i, terms[i].Name, name)
		}
	}
}

func heroTestPosts(ids ...string) []models.Post {
	posts := make([]models.Post, len(ids))
	for i, id := range ids {
		posts[i] = feedTestPost(id, "author-"+id)
	}
	return posts
}

func TestRotateHeroTop_CyclesInOrder(t *testing.T) {
	base := heroTestPosts("a", "b", "c", "d")
	want := [][]string{
		{"a", "b", "c", "d"}, // round 1
		{"b", "a", "c", "d"}, // round 2
		{"c", "b", "a", "d"}, // round 3
		{"a", "b", "c", "d"}, // round 4 quay lại
		{"b", "a", "c", "d"}, // round 5
	}
	for i, expected := range want {
		round := i + 1
		got := RotateHeroTop(base, round, 3)
		for j := range expected {
			if got[j].ID != expected[j] {
				t.Fatalf("round %d: position %d = %s, want %s", round, j, got[j].ID, expected[j])
			}
		}
	}
}

func TestRotateHeroTop_DoesNotMutateInput(t *testing.T) {
	base := heroTestPosts("a", "b", "c")
	_ = RotateHeroTop(base, 2, 3)
	if base[0].ID != "a" {
		t.Fatalf("input was mutated: %v", ids(base))
	}
}

func TestRotateHeroTop_Guards(t *testing.T) {
	base := heroTestPosts("a", "b", "c")
	for _, tc := range []struct {
		name  string
		round int
		slots int
	}{{"round zero", 0, 3}, {"slots one", 2, 1}, {"slots zero", 2, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			got := RotateHeroTop(base, tc.round, tc.slots)
			if len(got) != len(base) || got[0].ID != base[0].ID {
				t.Fatalf("expected unchanged order, got %v", ids(got))
			}
		})
	}
	if got := RotateHeroTop(heroTestPosts("a"), 5, 3); len(got) != 1 {
		t.Fatalf("single post must stay, got %v", ids(got))
	}
}

func ids(posts []models.Post) []string {
	out := make([]string, len(posts))
	for i, p := range posts {
		out[i] = p.ID
	}
	return out
}

func TestSpreadAuthors_PushesDuplicatesOutOfWindow(t *testing.T) {
	posts := []models.Post{
		feedTestPost("p1", "a"),
		feedTestPost("p2", "a"),
		feedTestPost("p3", "b"),
		feedTestPost("p4", "c"),
		feedTestPost("p5", "d"),
	}
	got := SpreadAuthors(posts, 5)
	want := []string{"p1", "p3", "p4", "p5", "p2"}
	if strings.Join(ids(got), ",") != strings.Join(want, ",") {
		t.Fatalf("got %v want %v", ids(got), want)
	}
	// Không mất bài.
	if len(got) != len(posts) {
		t.Fatalf("lost posts: %d -> %d", len(posts), len(got))
	}
}

func TestSpreadAuthors_KeepsFullPage(t *testing.T) {
	posts := []models.Post{
		feedTestPost("p1", "a"), feedTestPost("p2", "a"), feedTestPost("p3", "a"),
		feedTestPost("p4", "a"), feedTestPost("p5", "a"), feedTestPost("p6", "a"),
	}
	got := SpreadAuthors(posts, 5)
	if len(got) != len(posts) {
		t.Fatalf("page size changed: %d -> %d", len(posts), len(got))
	}
}

func TestEnsureExploreInTop(t *testing.T) {
	main := func(id string) models.Post {
		p := feedTestPost(id, "m-"+id)
		p.FeedSource = "main"
		return p
	}
	explore := func(id string) models.Post {
		p := feedTestPost(id, "e-"+id)
		p.FeedSource = "explore"
		return p
	}

	t.Run("đã có explore trong top thì giữ nguyên", func(t *testing.T) {
		posts := []models.Post{main("m1"), explore("e1"), main("m2"), main("m3")}
		got := EnsureExploreInTop(posts, 3)
		if got[1].ID != "e1" {
			t.Fatalf("expected unchanged, got %v", ids(got))
		}
	})

	t.Run("kéo explore lên slot cuối của top", func(t *testing.T) {
		posts := []models.Post{main("m1"), main("m2"), main("m3"), explore("e1"), main("m4")}
		got := EnsureExploreInTop(posts, 3)
		if got[2].ID != "e1" {
			t.Fatalf("expected e1 at slot 3, got %v", ids(got))
		}
		if len(got) != len(posts) {
			t.Fatalf("page size changed: %d -> %d", len(posts), len(got))
		}
	})

	t.Run("không có explore thì giữ nguyên", func(t *testing.T) {
		posts := []models.Post{main("m1"), main("m2"), main("m3")}
		got := EnsureExploreInTop(posts, 3)
		if strings.Join(ids(got), ",") != "m1,m2,m3" {
			t.Fatalf("expected unchanged, got %v", ids(got))
		}
	})

	t.Run("tắt khi slots=0", func(t *testing.T) {
		posts := []models.Post{main("m1"), main("m2"), explore("e1")}
		got := EnsureExploreInTop(posts, 0)
		if got[2].ID != "e1" {
			t.Fatalf("expected unchanged, got %v", ids(got))
		}
	})
}

func TestMergeFeedStreams_MarksSource(t *testing.T) {
	main := []models.Post{feedTestPost("m1", "a"), feedTestPost("m2", "b")}
	small := []models.Post{feedTestPost("s1", "c")}
	general := []models.Post{feedTestPost("g1", "d")}

	merged, _, _, _, _, _, _ := MergeFeedStreams(main, small, general, 2, 1, 1, 4, 5, 2)
	sources := map[string]string{}
	for _, p := range merged {
		sources[p.ID] = p.FeedSource
	}
	if sources["m1"] != "main" || sources["s1"] != "explore" || sources["g1"] != "explore" {
		t.Fatalf("unexpected sources: %v", sources)
	}
}
