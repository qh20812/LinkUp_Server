package models

import "time"

type Profile struct {
	ID                         string     `json:"id"`
	UserID                     string     `json:"user_id"`
	DisplayName                string     `json:"display_name"`
	PhoneNumber                string     `json:"phone_number"`
	DateOfBirth                *time.Time `json:"date_of_birth,omitempty"`
	AvatarURI                  string     `json:"avatar_uri"`
	Bio                        string     `json:"bio"`
	IsPrivateProfile           bool       `json:"is_private_profile"`
	IsPrivatePosts             bool       `json:"is_private_posts"`
	AllowStrangerFriendRequest bool       `json:"allow_stranger_friend_request"`
	LastReadMissedAt           *time.Time `json:"last_read_missed_at,omitempty"`
	UpdatedAt                  *time.Time `json:"updated_at,omitempty"`
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
