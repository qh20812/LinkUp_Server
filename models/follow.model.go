package models

import "time"

type Follow struct {
	ID          int64     `json:"id" db:"id"`
	FollowerID  int64     `json:"follower_id" db:"follower_id"`
	FollowingID int64     `json:"following_id" db:"following_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

func NewFollow(followerID, followingID int64) Follow {
	return Follow{FollowerID: followerID, FollowingID: followingID}
}
