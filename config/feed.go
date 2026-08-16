package config

// FeedWeights cấu hình trọng số cho feed ranking algorithm.
// Điều chỉnh các giá trị này để thay đổi "personality" của feed.
type FeedWeights struct {
	Recency    float64 // Trọng số cho độ tươi mới (default: 0.4)
	Engagement float64 // Trọng số cho tương tác (default: 0.35)
	Affinity   float64 // Trọng số cho mối quan hệ (default: 0.25)

	DecayRate float64 // Lambda cho exponential decay (default: 0.1)

	LikeWeight    float64 // Hệ số cho likes (default: 1.0)
	CommentWeight float64 // Hệ số cho comments (default: 4.0)
	ShareWeight   float64 // Hệ số cho shares (default: 6.0)
}

// DefaultFeedWeights là cấu hình mặc định cho feed ranking.
//
// Công thức scoring:
//
//	Final Score = Recency × RecencyScore + Engagement × EngagementScore + Affinity × AffinityScore
//
// Trong đó:
//   - RecencyScore = EXP(-DecayRate × age_hours)
//   - EngagementScore = (LOG(1+likes)×LikeWeight + LOG(1+comments)×CommentWeight + LOG(1+shares)×ShareWeight) / Normalizer
//   - AffinityScore = 1.0 nếu đang follow, 0.0 nếu không
var DefaultFeedWeights = FeedWeights{
	Recency:    0.4,
	Engagement: 0.35,
	Affinity:   0.25,

	DecayRate: 0.1,

	LikeWeight:    1.0,
	CommentWeight: 4.0,
	ShareWeight:   6.0,
}
