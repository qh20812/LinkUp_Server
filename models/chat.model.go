package models

import (
	"strings"
	"time"
)

type ChatType string

const (
	ChatTypeDirect ChatType = "direct"
	ChatTypeGroup  ChatType = "group"
)

type ChatStatus string

const (
	ChatStatusActive   ChatStatus = "active"
	ChatStatusHidden   ChatStatus = "hidden"
	ChatStatusArchived ChatStatus = "archived"
)

type Chat struct {
	ID            string     `json:"id"`
	Type          ChatType   `json:"type"`
	CreatorID     *string    `json:"creator_id,omitempty"`
	Name          string     `json:"name"`
	AvatarURI     string     `json:"avatar_uri"`
	Status        ChatStatus `json:"status" gorm:"type:varchar(20);default:active"`
	EncryptionKey string     `json:"-" gorm:"column:encryption_key"`
	CommunityID   *string    `json:"community_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

func NewChat(chatType ChatType, name, avatarURI string) Chat {
	if chatType == "" {
		chatType = ChatTypeDirect
	}
	return Chat{
		Type:      chatType,
		Name:      name,
		AvatarURI: avatarURI,
		Status:    ChatStatusActive,
	}
}

func (t ChatType) String() string {
	return string(t)
}

func ParseChatType(value string) ChatType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ChatTypeDirect):
		return ChatTypeDirect
	case string(ChatTypeGroup):
		return ChatTypeGroup
	default:
		return ChatTypeDirect
	}
}

func (s ChatStatus) String() string {
	return string(s)
}

func ParseChatStatus(value string) ChatStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ChatStatusHidden):
		return ChatStatusHidden
	case string(ChatStatusArchived):
		return ChatStatusArchived
	default:
		return ChatStatusActive
	}
}
