package models

import (
	"strings"
	"time"
)

type MediaStatus string

const (
	MediaStatusPending  MediaStatus = "pending"
	MediaStatusApproved MediaStatus = "approved"
	MediaStatusRejected MediaStatus = "rejected"
	MediaStatusFlagged  MediaStatus = "flagged"
)

type Media struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	PostID    *string     `json:"post_id,omitempty"`
	FileURI   string      `json:"file_uri"`
	FileType  string      `json:"file_type"`
	FileSize  float64     `json:"file_size"`
	Status       MediaStatus `json:"status"`
	ReviewReason *string     `json:"review_reason,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

func NewMedia(userID string, postID *string, fileURI, fileType string, fileSize float64) Media {
	return Media{
		UserID:   userID,
		PostID:   postID,
		FileURI:  fileURI,
		FileType: fileType,
		FileSize: fileSize,
		Status:   MediaStatusPending,
	}
}

func (m MediaStatus) String() string {
	return string(m)
}

func ParseMediaStatus(value string) MediaStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(MediaStatusPending):
		return MediaStatusPending
	case string(MediaStatusApproved):
		return MediaStatusApproved
	case string(MediaStatusRejected):
		return MediaStatusRejected
	case string(MediaStatusFlagged):
		return MediaStatusFlagged
	default:
		return MediaStatusPending
	}
}
