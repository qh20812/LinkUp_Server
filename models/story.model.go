package models

import (
	"strings"
	"time"
)

type StoryMediaType string

const (
	StoryMediaTypeImage StoryMediaType = "image"
	StoryMediaTypeVideo StoryMediaType = "video"
)

type Story struct {
	ID         int64          `json:"id" db:"id"`
	UserID     int64          `json:"user_id" db:"user_id"`
	MediaURI   string         `json:"media_uri" db:"media_uri"`
	MediaType  StoryMediaType `json:"media_type" db:"media_type"`
	Caption    string         `json:"caption" db:"caption"`
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
	ExpiresAt  *time.Time     `json:"expires_at,omitempty" db:"expires_at"`
}

func NewStory(userID int64, mediaURI string, mediaType StoryMediaType, caption string) Story {
	return Story{
		UserID:    userID,
		MediaURI:  mediaURI,
		MediaType: mediaType,
		Caption:   caption,
	}
}

func (m StoryMediaType) String() string {
	return string(m)
}

func ParseStoryMediaType(value string) StoryMediaType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(StoryMediaTypeImage):
		return StoryMediaTypeImage
	case string(StoryMediaTypeVideo):
		return StoryMediaTypeVideo
	default:
		return StoryMediaTypeVideo
	}
}
