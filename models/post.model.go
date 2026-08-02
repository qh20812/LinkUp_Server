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
	ViewsCount  int        `json:"views_count"`
	Status      PostStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`

	LikesCount    int `json:"likes_count" gorm:"->"`
	CommentsCount int `json:"comments_count" gorm:"->"`
	SharesCount   int `json:"shares_count" gorm:"->"`

	Username    string `json:"username" gorm:"->"`
	DisplayName string `json:"display_name" gorm:"->"`
	AvatarURI   string `json:"avatar_uri" gorm:"->"`

	Media []Media `json:"media" gorm:"-"`

	IsLiked     bool `json:"is_liked" gorm:"->"`
	IsSaved     bool `json:"is_saved" gorm:"->"`
	IsFollowing bool `json:"is_following" gorm:"->"`

	SavedAt    *time.Time `json:"saved_at,omitempty" gorm:"->"`
	BookmarkID *string    `json:"bookmark_id,omitempty" gorm:"->"`
}

func NewPost(userID, title, content string, status PostStatus) Post {
	if status == "" {
		status = PostStatusPublic // Mặc định là public nếu client không truyền
	}
	return Post{
		UserID:  userID,
		Title:   title,
		Content: content,
		Status:  status,
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
