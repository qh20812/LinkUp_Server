package models

type NotificationPreference struct {
	UserID               string `json:"user_id" gorm:"primaryKey"`
	LikeEnabled          bool   `json:"like_enabled"`
	CommentEnabled       bool   `json:"comment_enabled"`
	FollowEnabled        bool   `json:"follow_enabled"`
	MessageEnabled       bool   `json:"message_enabled"`
	FriendRequestEnabled bool   `json:"friend_request_enabled"`
}
