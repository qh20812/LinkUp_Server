package models

type NotificationPreference struct {
	UserID               string `json:"user_id" gorm:"primaryKey"`
	LikeEnabled          bool   `json:"like_enabled"`
	CommentEnabled       bool   `json:"comment_enabled"`
	FollowEnabled        bool   `json:"follow_enabled"`
	MessageEnabled       bool   `json:"message_enabled"`
	FriendRequestEnabled bool   `json:"friend_request_enabled"`
	CommunityEnabled     bool   `json:"community_enabled"`
	VoiceCallEnabled     bool   `json:"voice_call_enabled"`
	StoryReactEnabled    bool   `json:"story_react_enabled"`
	ShareEnabled         bool   `json:"share_enabled"`
	MediaEnabled         bool   `json:"media_enabled"`
}

// NewDefaultNotificationPreference returns an all-enabled preference row for
// a new user (or a user predating preference rows).
func NewDefaultNotificationPreference(userID string) *NotificationPreference {
	return &NotificationPreference{
		UserID:               userID,
		LikeEnabled:          true,
		CommentEnabled:       true,
		FollowEnabled:        true,
		MessageEnabled:       true,
		FriendRequestEnabled: true,
		CommunityEnabled:     true,
		VoiceCallEnabled:     true,
		StoryReactEnabled:    true,
		ShareEnabled:         true,
		MediaEnabled:         true,
	}
}
