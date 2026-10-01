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

// FeedRankV2Config cấu hình ranking cá nhân hóa 2-stage (Phase 2).
// Main stream điểm chuẩn hóa theo cohort + interest; explore slots bảo vệ
// tác giả nhỏ (< SmallAuthorFollowers followers) và bài mới.
type FeedRankV2Config struct {
	InterestWeight    float64 // Trọng số khớp interest-profile (default: 0.3)
	InterestDecayRate float64 // Lambda decay interest theo ngày (default: 0.05)

	AffinityFollow     float64 // Mức affinity khi follow tác giả (default: 1.0)
	AffinityFriend     float64 // Mức affinity khi là bạn bè (default: 0.6)
	AffinityInteracted float64 // Mức affinity khi đã tương tác tác giả (default: 0.3)

	SeenPenalty    float64 // Trừ điểm mỗi lần đã thấy mà không tương tác (default: 0.15)
	SeenWindowDays int     // Cửa sổ tính lượt đã thấy, theo ngày (default: 7)
	SeenCap        int     // Trần số lần thấy bị phạt (default: 3)

	NewPostBoost      float64 // Bonus khám phá bài mới (default: 0.5)
	NewPostBoostHours float64 // Cửa sổ giờ áp dụng boost (default: 6)

	SmallAuthorFollowers int     // Ngưỡng tác giả nhỏ theo followers (default: 100)
	SmallAuthorBonus     float64 // Bonus điểm cho tác giả nhỏ ở main stream (default: 0.2)

	ExploreMaxAgeHours float64 // Tuổi tối đa của bài explore, theo giờ (default: 72)
	ExploreEveryN      int     // Mỗi slot thứ N là explore → N=5 tương đương 20% (default: 5)
	MaxPerAuthor       int     // Trần số bài cùng tác giả mỗi trang (default: 2)

	MainRecallDays int // Giới hạn tuổi bài ở main stream, theo ngày (default: 90)

	RankV2RolloutPct int // % user tự động dùng rank v2 khi không truyền ?rank= (default: 0)
}

var DefaultFeedRankV2 = FeedRankV2Config{
	InterestWeight:    0.3,
	InterestDecayRate: 0.05,

	AffinityFollow:     1.0,
	AffinityFriend:     0.6,
	AffinityInteracted: 0.3,

	SeenPenalty:    0.15,
	SeenWindowDays: 7,
	SeenCap:        3,

	NewPostBoost:      0.5,
	NewPostBoostHours: 6,

	SmallAuthorFollowers: 100,
	SmallAuthorBonus:     0.2,

	ExploreMaxAgeHours: 72,
	ExploreEveryN:      5,
	MaxPerAuthor:       2,

	MainRecallDays: 90,

	RankV2RolloutPct: 0,
}
