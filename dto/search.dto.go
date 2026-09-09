package dto

import "time"

type SearchInput struct {
	Keyword string `json:"keyword" form:"keyword"`
	Type    string `json:"type" form:"type"`
}

type UserSearchResult struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURI   string `json:"avatar_uri"`
}

type PostSearchResult struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type HashtagSearchResult struct {
	Name      string `json:"name"`
	PostCount int64  `json:"post_count"`
}

type CommunitySearchResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AvatarURI   string `json:"avatar_uri"`
	MemberCount int    `json:"member_count"`
	Privacy     string `json:"privacy"`
}

type SearchResponse struct {
	Users       []UserSearchResult       `json:"users,omitempty"`
	Posts       []PostSearchResult       `json:"posts,omitempty"`
	Hashtags    []HashtagSearchResult    `json:"hashtags,omitempty"`
	Communities []CommunitySearchResult  `json:"communities,omitempty"`
	Message     string                   `json:"message,omitempty"`
}
