package models

import "time"

type Community struct {
	ID            string     `json:"id"`
	CreatorID     string     `json:"creator_id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	AvatarURI     string     `json:"avatar_uri"`
	BackgroundURI string     `json:"background_uri"`
	AutoApprove   bool       `json:"auto_approve"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

func NewCommunity(creatorID, name, description, avatarURI string) Community {
	return Community{CreatorID: creatorID, Name: name, Description: description, AvatarURI: avatarURI, BackgroundURI: ""}
}
