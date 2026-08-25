package dto

import "time"

type CreateGroupInput struct {
	Name      string   `json:"name" binding:"required,min=3,max=50"`
	AvatarURI string   `json:"avatar_uri"`
	MemberIDs []string `json:"member_ids"`
}

type AddMemberInput struct {
	UserID string `json:"user_id" binding:"required"`
}

// Input cho tính năng Ban thành viên
type BanMemberInput struct {
	UserID string `json:"user_id" binding:"required"`
}

// Input cho tính năng gửi tin nhắn nhóm qua HTTP (nếu không dùng qua trực tiếp WS)
type SendGroupMessageInput struct {
	Content          string  `json:"content" binding:"required"`
	EmojiID          *string `json:"emoji_id"`
	MediaID          *string `json:"media_id"`
	ReplyToMessageID *string `json:"reply_to_message_id,omitempty"`
	SharedPostID     *string `json:"shared_post_id,omitempty"`
}

type GroupChatSettingsDTO struct {
	NotificationsEnabled *bool   `json:"notifications_enabled,omitempty"`
	AllowMemberAdd       *bool   `json:"allow_member_add,omitempty"`
	Name                 *string `json:"name,omitempty"`
	AvatarURI            *string `json:"avatar_uri,omitempty"`
}

type TransferAdminInput struct {
	TargetUserID string `json:"target_user_id" binding:"required"`
}

type GroupChatMemberSettingsResponse struct {
	NotificationsEnabled bool `json:"notifications_enabled"`
}

type GroupChatMemberDTO struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	AvatarURI   string `json:"avatar_uri"`
	Role        string `json:"role"`
}

type GroupChatSettingsResponse struct {
	ChatID         string                          `json:"chat_id"`
	Name           string                          `json:"name"`
	AvatarURI      string                          `json:"avatar_uri"`
	AllowMemberAdd bool                            `json:"allow_member_add"`
	MemberSettings GroupChatMemberSettingsResponse `json:"member_settings"`
	Members        []GroupChatMemberDTO            `json:"members,omitempty"`
}

type MuteMemberInput struct {
	UserID       string `json:"user_id" binding:"required"`
	Reason       string `json:"reason" binding:"required,oneof=spam abuse harassment violation other"`
	DurationMins int    `json:"duration_minutes" binding:"required,oneof=1 30 60 1440 0"` // 0 = permanent
}

type UnmuteMemberInput struct {
	UserID string `json:"user_id" binding:"required"`
}

// GroupChatTransferOwnershipInput là input cho tính năng chuyển quyền sở hữu nhóm chat.
type GroupChatTransferOwnershipInput struct {
	TargetUserID string `json:"target_user_id" binding:"required"`
	KeepAdmin    bool   `json:"keep_admin"`
}

type GroupChatConversationDTO struct {
	ChatID      string          `json:"chat_id"`
	Name        string          `json:"name"`
	AvatarURI   string          `json:"avatar_uri"`
	MemberCount int             `json:"member_count"`
	LastMessage *MessagePayload `json:"last_message,omitempty"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type GroupChatListResponse struct {
	Data []GroupChatConversationDTO `json:"data"`
}
