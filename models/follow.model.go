package models

import "time"

type Follow struct {
	ID          string     `json:"id"`
	FollowerID  string `json:"follower_id"`
	FollowingID string `json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewFollow(followerID, followingID string) Follow {
	return Follow{FollowerID: followerID, FollowingID: followingID}
}
