package models

import (
	"strings"
)

type TagType string

const (
	TagTypeHashtag TagType = "hashtag"
	TagTypeMention TagType = "mention"
)

type Tag struct {
	ID           int64   `json:"id" db:"id"`
	PostID       int64   `json:"post_id" db:"post_id"`
	CommentID    *int64  `json:"comment_id,omitempty" db:"comment_id"`
	TagType      TagType `json:"tag_type" db:"tag_type"`
	TargetUserID *int64  `json:"target_user_id,omitempty" db:"target_user_id"`
	Name         string  `json:"name" db:"name"`
}

func NewTag(postID int64, commentID *int64, tagType TagType, targetUserID *int64, name string) Tag {
	return Tag{PostID: postID, CommentID: commentID, TagType: tagType, TargetUserID: targetUserID, Name: name}
}

func (t TagType) String() string {
	return string(t)
}

func ParseTagType(value string) TagType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(TagTypeHashtag):
		return TagTypeHashtag
	case string(TagTypeMention):
		return TagTypeMention
	default:
		return TagTypeHashtag
	}
}
