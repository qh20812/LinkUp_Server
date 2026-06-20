package models

import "time"

type Profile struct {
	ID                         string     `json:"id" db:"id"`
	UserID                     string     `json:"user_id" db:"user_id"`
	DisplayName                string     `json:"display_name" db:"display_name"`
	PhoneNumber                string     `json:"phone_number" db:"phone_number"`
	DateOfBirth                *time.Time `json:"date_of_birth,omitempty" db:"date_of_birth"`
	AvatarURI                  string     `json:"avatar_uri" db:"avatar_uri"`
	Bio                        string     `json:"bio" db:"bio"`
	IsPrivateProfile           bool       `json:"is_private_profile" db:"is_private_profile"`
	IsPrivatePosts             bool       `json:"is_private_posts" db:"is_private_posts"`
	AllowStrangerFriendRequest bool       `json:"allow_stranger_friend_request" db:"allow_stranger_friend_request"`
	UpdatedAt                  *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

func NewProfile(userID, displayName, phoneNumber string, dateOfBirth *time.Time, avatarURI, bio string) Profile {
	return Profile{
		UserID:                     userID,
		DisplayName:                displayName,
		PhoneNumber:                phoneNumber,
		DateOfBirth:                dateOfBirth,
		AvatarURI:                  avatarURI,
		Bio:                        bio,
		IsPrivateProfile:           false,
		IsPrivatePosts:             false,
		AllowStrangerFriendRequest: true,
	}
}
