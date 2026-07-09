package dto

type GroupJoinPayload struct {
	ChatID string `json:"chat_id"`
}

type GroupLeavePayload struct {
	ChatID      string `json:"chat_id"`
	LeaveMode   string `json:"leave_mode"`
	HistoryMode string `json:"history_mode"`
}

type GroupSendMessagePayload struct {
	ChatID           string  `json:"chat_id"`
	Content          string  `json:"content"`
	EmojiID          *string `json:"emoji_id,omitempty"`
	MediaID          *string `json:"media_id,omitempty"`
	ReplyToMessageID *string `json:"reply_to_message,omitempty"`
}

type GroupTypingPayload struct {
	ChatID string `json:"chat_id"`
}

type GroupSearchPayload struct {
	ChatID  string `json:"chat_id"`
	Keyword string `json:"keyword"`
}

type GroupMemberActionPayload struct {
	ChatID string `json:"chat_id"`
	UserID string `json:"user_id"`
}

type GroupBanPayload struct {
	ChatID string `json:"chat_id"`
	UserID string `json:"user_id"`
}

type GroupMutePayload struct {
	ChatID       string `json:"chat_id"`
	UserID       string `json:"user_id"`
	Reason       string `json:"reason"`
	DurationMins int    `json:"duration_minutes"`
}

type GroupUnmutePayload struct {
	ChatID string `json:"chat_id"`
	UserID string `json:"user_id"`
}

type GroupTransferAdminPayload struct {
	ChatID       string `json:"chat_id"`
	TargetUserID string `json:"target_user_id"`
}

type GroupSettingsUpdatePayload struct {
	ChatID               string  `json:"chat_id"`
	NotificationsEnabled *bool   `json:"notifications_enabled,omitempty"`
	AllowMemberAdd       *bool   `json:"allow_member_add,omitempty"`
	Name                 *string `json:"name,omitempty"`
	AvatarURI            *string `json:"avatar_uri,omitempty"`
}

type GroupHistoryPayload struct {
	ChatID   string           `json:"chat_id"`
	Messages []MessagePayload `json:"messages"`
}
