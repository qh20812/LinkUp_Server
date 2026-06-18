package models

import "time"

type Follow struct {
	ID          string     `json:"id" db:"id"`
	FollowerID  string `json:"follower_id" db:"follower_id"`
	FollowingID string `json:"following_id" db:"following_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

func NewFollow(followerID, followingID string) Follow {
	return Follow{FollowerID: followerID, FollowingID: followingID}
}
