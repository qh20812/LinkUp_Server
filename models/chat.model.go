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

type Chat struct {
	ID        string   `json:"id"`
	Type      ChatType `json:"type"`
	Name      string   `json:"name"`
	AvatarURI string   `json:"avatar_uri"`
	CreatedAt time.Time `json:"created_at"`
}

func NewChat(chatType ChatType, name, avatarURI string) Chat {
	if chatType == "" {
		chatType = ChatTypeDirect
	}
	return Chat{Type: chatType, Name: name, AvatarURI: avatarURI}
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
