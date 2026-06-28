package dto

type CreateGroupInput struct {
	Name      string   `json:"name" binding:"required,min=3,max=50"`
	AvatarURI string   `json:"avatar_uri"`
	MemberIDs []string `json:"member_ids"`
}

type AddMemberInput struct {
	UserID string `json:"user_id" binding:"required"`
}
