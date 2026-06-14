package models

import (
	"strings"
	"time"
)

type ChatParticipantRole string

const (
	ChatParticipantRoleAdmin  ChatParticipantRole = "admin"
	ChatParticipantRoleMember ChatParticipantRole = "member"
)

type ChatParticipant struct {
	ID       int64               `json:"id" db:"id"`
	ChatID   int64               `json:"chat_id" db:"chat_id"`
	UserID   int64               `json:"user_id" db:"user_id"`
	Role     ChatParticipantRole `json:"role" db:"role"`
	JoinedAt time.Time           `json:"joined_at" db:"joined_at"`
}

func NewChatParticipant(chatID, userID int64, role ChatParticipantRole) ChatParticipant {
	if role == "" {
		role = ChatParticipantRoleMember
	}
	return ChatParticipant{ChatID: chatID, UserID: userID, Role: role}
}

func (r ChatParticipantRole) String() string {
	return string(r)
}

func ParseChatParticipantRole(value string) ChatParticipantRole {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ChatParticipantRoleAdmin):
		return ChatParticipantRoleAdmin
	case string(ChatParticipantRoleMember):
		return ChatParticipantRoleMember
	default:
		return ChatParticipantRoleMember
	}
}
