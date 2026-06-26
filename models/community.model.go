package models

import "time"

type Community struct {
	ID          string         `json:"id"`
	CreatorID   string         `json:"creator_id"`
	Name        string         `json:"name"`
	Role        CommunityRole  `json:"role"`
	Description string         `json:"description"`
	AvatarURI   string         `json:"avatar_uri"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   *time.Time     `json:"updated_at,omitempty"`
}

func NewCommunity(creatorID, name, description, avatarURI string, role CommunityRole) Community {
	return Community{CreatorID: creatorID, Name: name, Role: role, Description: description, AvatarURI: avatarURI}
}
