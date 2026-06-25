package models

import "time"

type ChatParticipant struct {
	ID       string    `json:"id"`
	ChatID   string    `json:"chat_id"`
	UserID   string    `json:"user_id"`
	Role     ChatRole  `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

func NewChatParticipant(chatID, userID string, role ChatRole) ChatParticipant {
	if role == "" {
		role = ChatRoleMember
	}
	return ChatParticipant{ChatID: chatID, UserID: userID, Role: role}
}
