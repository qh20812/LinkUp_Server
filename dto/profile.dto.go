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
    CoverURI                   string     `json:"cover_uri"`
    Username                   string     `json:"username"`
    PostCount                  int64      `json:"post_count"`
    CreatedAt                  time.Time  `json:"created_at"`
    UpdatedAt                  *time.Time `json:"updated_at,omitempty"`
}

type EditProfileInput struct {
    DisplayName                *string    `json:"display_name"`
    PhoneNumber                *string    `json:"phone_number"`
    DateOfBirth                *time.Time `json:"date_of_birth"`
    AvatarURI                  *string    `json:"avatar_uri"`
    CoverURI                   *string    `json:"cover_uri"`
    Bio                        *string    `json:"bio"`
    IsPrivateProfile           *bool      `json:"is_private_profile"`
    IsPrivatePosts             *bool      `json:"is_private_posts"`
    AllowStrangerFriendRequest *bool      `json:"allow_stranger_friend_request"`
}

type EditProfileResponse struct {
    Message string              `json:"message"`
    Data    ViewProfileResponse `json:"data"`
}