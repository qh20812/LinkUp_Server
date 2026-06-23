package dto

type FollowToggleResponse struct {
	Action         string `json:"action"`
	IsFollowing    bool   `json:"is_following"`
	FollowerCount  int64  `json:"follower_count"`
	FollowingCount int64  `json:"following_count"`
	Message        string `json:"message"`
}

type FollowStatsResponse struct {
	FollowerCount  int64 `json:"follower_count"`
	FollowingCount int64 `json:"following_count"`
	IsFollowing    bool  `json:"is_following"`
}
