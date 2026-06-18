package models

import "time"

type ChatParticipant struct {
	ID       string    `json:"id" db:"id"`
	ChatID   string    `json:"chat_id" db:"chat_id"`
	UserID   string    `json:"user_id" db:"user_id"`
	Role     ChatRole  `json:"role" db:"role"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

func NewChatParticipant(chatID, userID string, role ChatRole) ChatParticipant {
	if role == "" {
		role = ChatRoleMember
	}
	return ChatParticipant{ChatID: chatID, UserID: userID, Role: role}
}
