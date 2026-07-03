package services

import (
	"testing"

	"linkup/dto"
	"linkup/models"
)

func TestContributionScoreAndBadgeHelpers(t *testing.T) {
	svc := &ContributionService{}
	policy := &models.CommunityPolicy{
		PostWeight:                  10,
		CommentWeight:               5,
		ReactionWeight:              2,
		EventWeight:                 20,
		TopContributorThreshold:     2500,
		ModeratorPromotionThreshold: 5000,
		AutoPromoteEnabled:          true,
		BadgeEnabled:                true,
	}

	contribution := &models.MemberContribution{
		ValidPosts:          100,
		QualityComments:     50,
		PositiveReactions:   25,
		EventParticipations: 10,
	}

	score := svc.calculateContributionScore(contribution, policy)
	if score != 1000+250+50+200 {
		t.Fatalf("score = %d, want %d", score, 1500)
	}

	contribution.ContributionScore = 2500
	svc.checkAndAssignBadge(contribution, policy)
	if contribution.BadgeType == nil || *contribution.BadgeType != "Top Contributor" {
		t.Fatalf("badge = %v, want Top Contributor", contribution.BadgeType)
	}

	contribution.ContributionScore = 5000
	svc.checkAndPromoteToMod(contribution, policy)
	if !contribution.PromotedToMod {
		t.Fatal("expected promoted_to_mod to be true")
	}
}

func TestExtractHashtags(t *testing.T) {
	content := "Joining the #LinkUpPhoto challenge! also #golang and #LinkUpPhoto."
	hashtags := extractHashtags(content)
	if len(hashtags) != 2 {
		t.Fatalf("len = %d, want 2; value = %#v", len(hashtags), hashtags)
	}
	if hashtags[0] != "#LinkUpPhoto" || hashtags[1] != "#golang" {
		t.Fatalf("hashtags = %#v", hashtags)
	}
}

func TestSortLeaderboard(t *testing.T) {
	svc := &ContributionService{}
	items := []dto.LeaderboardItem{
		{UserID: "b", ContributionScore: 100},
		{UserID: "a", ContributionScore: 200},
		{UserID: "c", ContributionScore: 200},
	}

	svc.sortLeaderboard(items)

	if items[0].UserID != "a" || items[1].UserID != "c" || items[2].UserID != "b" {
		t.Fatalf("sorted order incorrect: %#v", items)
	}
}
