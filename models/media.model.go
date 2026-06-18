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
)

type Media struct {
	ID        string      `json:"id" db:"id"`
	UserID    string      `json:"user_id" db:"user_id"`
	PostID    *string     `json:"post_id,omitempty" db:"post_id"`
	FileURI   string      `json:"file_uri" db:"file_uri"`
	FileType  string      `json:"file_type" db:"file_type"`
	FileSize  float64     `json:"file_size" db:"file_size"`
	Status    MediaStatus `json:"status" db:"status"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
}

func NewMedia(userID string, postID *string, fileURI, fileType string, fileSize float64) Media {
	return Media{
		UserID:   userID,
		PostID:   postID,
		FileURI:  fileURI,
		FileType: fileType,
		FileSize: fileSize,
		Status:   MediaStatusApproved,
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
	default:
		return MediaStatusApproved
	}
}
