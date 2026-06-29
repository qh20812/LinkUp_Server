package dto

type CreateCommunityInput struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
	AvatarURI   string `json:"avatar_uri" binding:"omitempty,url"`
}
