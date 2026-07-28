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

type FollowSuggestionItem struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURI   string `json:"avatar_uri"`
	MutualCount int    `json:"mutual_count"`
}

type FollowSuggestionsResponse struct {
	Data     []FollowSuggestionItem `json:"data"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Total    int64                  `json:"total"`
	HasMore  bool                   `json:"has_more"`
}
