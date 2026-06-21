package dto

import "time"

type ViewProfileResponse struct {
	DisplayName                string     `json:"display_name"`
	PhoneNumber                string     `json:"phone_number"`
	DateOfBirth                *time.Time `json:"date_of_birth,omitempty"`
	AvatarURI                  string     `json:"avatar_uri"`
	Bio                        string     `json:"bio"`
	IsPrivateProfile           bool       `json:"is_private_profile"`
	IsPrivatePosts             bool       `json:"is_private_posts"`
	AllowStrangerFriendRequest bool       `json:"allow_stranger_friend_request"`
	UpdatedAt                  *time.Time `json:"updated_at,omitempty"`
}

type EditProfileInput struct {
	DisplayName                string     `json:"display_name" binding:"required,max=100"`
	PhoneNumber                string     `json:"phone_number" binding:"max=20"`
	DateOfBirth                *time.Time `json:"date_of_birth"`
	AvatarURI                  string     `json:"avatar_uri" binding:"max=500"`
	Bio                        string     `json:"bio" binding:"max=500"`
	IsPrivateProfile           bool       `json:"is_private_profile"`
	IsPrivatePosts             bool       `json:"is_private_posts"`
	AllowStrangerFriendRequest bool       `json:"allow_stranger_friend_request"`
}

type EditProfileResponse struct {
	Message string              `json:"message"`
	Data    ViewProfileResponse `json:"data"`
}
