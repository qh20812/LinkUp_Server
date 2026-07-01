package dto

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
	Content string  `json:"content" binding:"required"`
	EmojiID *string `json:"emoji_id"`
	MediaID *string `json:"media_id"`
}