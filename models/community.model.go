package models

import "time"

type Community struct {
	ID          int64     `json:"id" db:"id"`
	CreatorID   int64     `json:"creator_id" db:"creator_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	AvatarURI   string    `json:"avatar_uri" db:"avatar_uri"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

func NewCommunity(creatorID int64, name, description, avatarURI string) Community {
	return Community{CreatorID: creatorID, Name: name, Description: description, AvatarURI: avatarURI}
}
