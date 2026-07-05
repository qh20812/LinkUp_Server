package models

import (
	"strings"
	"time"
)

type ChallengeStatus string

const (
	ChallengeStatusActive    ChallengeStatus = "active"
	ChallengeStatusCompleted ChallengeStatus = "completed"
	ChallengeStatusCancelled ChallengeStatus = "cancelled"
)

type CommunityChallenge struct {
	ID              string          `json:"id"`
	CommunityID     string          `json:"community_id"`
	CreatorID       string          `json:"creator_id"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Hashtag         string          `json:"hashtag"`
	PointsPerPost   int             `json:"points_per_post"`
	StartDate       time.Time       `json:"start_date"`
	EndDate         time.Time       `json:"end_date"`
	MaxParticipants *int            `json:"max_participants,omitempty"`
	Status          ChallengeStatus `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
}

func (s ChallengeStatus) String() string {
	return string(s)
}

func ParseChallengeStatus(value string) ChallengeStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ChallengeStatusCompleted):
		return ChallengeStatusCompleted
	case string(ChallengeStatusCancelled):
		return ChallengeStatusCancelled
	default:
		return ChallengeStatusActive
	}
}
