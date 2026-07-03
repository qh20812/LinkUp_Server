package models

import "time"

type ChallengeParticipant struct {
	ID                string    `json:"id"`
	ChallengeID       string    `json:"challenge_id"`
	UserID            string    `json:"user_id"`
	PostsCount        int       `json:"posts_count"`
	TotalPointsEarned int       `json:"total_points_earned"`
	JoinedAt          time.Time `json:"joined_at"`
}
