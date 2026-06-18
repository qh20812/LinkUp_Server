package models

import "time"

type Community struct {
	ID          string         `json:"id" db:"id"`
	CreatorID   string         `json:"creator_id" db:"creator_id"`
	Name        string         `json:"name" db:"name"`
	Role        CommunityRole  `json:"role" db:"role"`
	Description string         `json:"description" db:"description"`
	AvatarURI   string         `json:"avatar_uri" db:"avatar_uri"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time     `json:"updated_at,omitempty" db:"updated_at"`
}

func NewCommunity(creatorID, name, description, avatarURI string, role CommunityRole) Community {
	return Community{CreatorID: creatorID, Name: name, Role: role, Description: description, AvatarURI: avatarURI}
}
