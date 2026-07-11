package models

import (
	"strings"
	"time"
)

type NotificationType string

const (
	NotificationTypeLike                  NotificationType = "like"
	NotificationTypeComment               NotificationType = "comment"
	NotificationTypeFollow                NotificationType = "follow"
	NotificationTypeMessage               NotificationType = "message"
	NotificationTypeFriendRequest         NotificationType = "friend_request"
	NotificationTypeFriendAccepted            NotificationType = "friend_accepted"
	NotificationTypeCommunityJoinRequest      NotificationType = "community_join_request"
	NotificationTypeCommunityJoinApproved     NotificationType = "community_join_approved"
	NotificationTypeCommunityJoinRejected     NotificationType = "community_join_rejected"
	NotificationTypeCommunityRoleChanged      NotificationType = "community_role_changed"
	NotificationTypeCommunityMemberLeft      NotificationType = "community_member_left"
	NotificationTypeCommunityMemberKicked    NotificationType = "community_member_kicked"
	NotificationTypeCommunityGroupChatAdded     NotificationType = "community_group_chat_added"
	NotificationTypeCommunityInviteCodeUsed     NotificationType = "community_invite_code_used"
	NotificationTypeCommunityInvitationReceived NotificationType = "community_invitation_received"
	NotificationTypeCommunityInvitationAccepted NotificationType = "community_invitation_accepted"
	NotificationTypeVoiceCall                  NotificationType = "voice_call"
)

type Notification struct {
	ID                string           `json:"id"`
	ReceiverID        string           `json:"receiver_id"`
	SenderID          *string          `json:"sender_id,omitempty"`
	Type              NotificationType `json:"type"`
	RedirectPostID    *string          `json:"redirect_post_id,omitempty"`
	RedirectUserID    *string          `json:"redirect_user_id,omitempty"`
	RedirectCommentID *string          `json:"redirect_comment_id,omitempty"`
	Content           string           `json:"content"`
	IsRead            bool             `json:"is_read"`
	CreatedAt         time.Time        `json:"created_at"`
}

func NewNotification(receiverID string, senderID *string, notifType NotificationType, content string) Notification {
	return Notification{
		ReceiverID: receiverID,
		SenderID:   senderID,
		Type:       notifType,
		Content:    content,
	}
}

func (n NotificationType) String() string {
	return string(n)
}

func ParseNotificationType(value string) NotificationType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(NotificationTypeLike):
		return NotificationTypeLike
	case string(NotificationTypeComment):
		return NotificationTypeComment
	case string(NotificationTypeFollow):
		return NotificationTypeFollow
	case string(NotificationTypeMessage):
		return NotificationTypeMessage
	case string(NotificationTypeFriendRequest):
		return NotificationTypeFriendRequest
	case string(NotificationTypeFriendAccepted):
		return NotificationTypeFriendAccepted
	case string(NotificationTypeCommunityJoinRequest):
		return NotificationTypeCommunityJoinRequest
	case string(NotificationTypeCommunityJoinApproved):
		return NotificationTypeCommunityJoinApproved
	case string(NotificationTypeCommunityJoinRejected):
		return NotificationTypeCommunityJoinRejected
	case string(NotificationTypeCommunityRoleChanged):
		return NotificationTypeCommunityRoleChanged
	case string(NotificationTypeCommunityMemberLeft):
		return NotificationTypeCommunityMemberLeft
	case string(NotificationTypeCommunityMemberKicked):
		return NotificationTypeCommunityMemberKicked
	case string(NotificationTypeCommunityGroupChatAdded):
		return NotificationTypeCommunityGroupChatAdded
	case string(NotificationTypeCommunityInviteCodeUsed):
		return NotificationTypeCommunityInviteCodeUsed
	case string(NotificationTypeCommunityInvitationReceived):
		return NotificationTypeCommunityInvitationReceived
	case string(NotificationTypeCommunityInvitationAccepted):
		return NotificationTypeCommunityInvitationAccepted
	case string(NotificationTypeVoiceCall):
		return NotificationTypeVoiceCall
	default:
		return NotificationTypeLike
	}
}
