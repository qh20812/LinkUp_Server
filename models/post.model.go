package models

import (
	"strings"
	"time"
)

type PostStatus string

const (
	PostStatusPublic  PostStatus = "public"
	PostStatusPrivate PostStatus = "private"
	PostStatusHidden  PostStatus = "hidden"
	PostStatusFriend  PostStatus = "friend"
	PostStatusDeleted PostStatus = "deleted"
)

type Post struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	UserID      string     `json:"user_id"`
	CommunityID *string    `json:"community_id,omitempty"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	ViewsCount      int        `json:"views_count"`
	Status          PostStatus `json:"status"`
	IsPinned        bool       `json:"is_pinned"`
	PinnedAt        *time.Time `json:"pinned_at,omitempty"`
	CommentsEnabled bool       `json:"comments_enabled"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`

	SharedFromPostID *string `json:"shared_from_post_id,omitempty"`
	ShareContent     *string `json:"share_content,omitempty"`

	LikesCount    int `json:"likes_count" gorm:"->"`
	CommentsCount int `json:"comments_count" gorm:"->"`
	SharesCount   int `json:"shares_count" gorm:"->"`

	Username    string `json:"username" gorm:"->"`
	DisplayName string `json:"display_name" gorm:"->"`
	AvatarURI   string `json:"avatar_uri" gorm:"->"`

	Media      []Media `json:"media" gorm:"-"`
	SharedPost *Post   `json:"shared_post,omitempty" gorm:"-"`

	IsLiked     bool `json:"is_liked" gorm:"->"`
	IsSaved     bool `json:"is_saved" gorm:"->"`
	IsShared    bool `json:"is_shared" gorm:"->"`
	IsFollowing bool `json:"is_following" gorm:"->"`

	FeedScore float64 `json:"-" gorm:"->"`

	// FeedSource chỉ dùng trong bộ nhớ để rerank biết bài đến từ stream nào
	// ("main" / "explore"). Không trả về client, không map cột DB.
	FeedSource string `json:"-" gorm:"-"`

	// FeedDebug chỉ được gắn khi bật FEED_DEBUG + ?debug=1.
	FeedDebug *FeedDebugScore `json:"feed_debug,omitempty" gorm:"-"`

	SavedAt    *time.Time `json:"saved_at,omitempty" gorm:"->"`
	BookmarkID *string    `json:"bookmark_id,omitempty" gorm:"->"`
}

// FeedDebugScore là điểm từng thành phần của rank v2, dùng để trả lời
// "vì sao bài này đứng đây". Các giá trị đã bao gồm dấu (phạt là số âm),
// nên tổng của chúng xấp xỉ Final.
type FeedDebugScore struct {
	Final               float64 `json:"final"`
	Recency             float64 `json:"recency"`
	Engagement          float64 `json:"engagement"`
	Affinity            float64 `json:"affinity"`
	Interest            float64 `json:"interest"`
	NewBoost            float64 `json:"new_boost"`
	SmallBonus          float64 `json:"small_bonus"`
	SeenPenalty         float64 `json:"seen_penalty"`
	SeenPenalty24h      float64 `json:"seen_penalty_24h"`
	HiddenAuthorPenalty float64 `json:"hidden_author_penalty"`
}

func NewPost(userID, title, content string, status PostStatus) Post {
	if status == "" {
		status = PostStatusPublic // Mặc định là public nếu client không truyền
	}
	return Post{
		UserID:          userID,
		Title:           title,
		Content:         content,
		Status:          status,
		CommentsEnabled: true,
	}
}

func (s PostStatus) String() string {
	return string(s)
}

func (Post) TableName() string {
	return "posts"
}

func ParsePostStatus(value string) PostStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "active", "public":
		return PostStatusPublic
	case string(PostStatusPrivate):
		return PostStatusPrivate
	case string(PostStatusHidden):
		return PostStatusHidden
	case string(PostStatusFriend):
		return PostStatusFriend
	case string(PostStatusDeleted):
		return PostStatusDeleted
	default:
		return PostStatusPublic
	}
}
