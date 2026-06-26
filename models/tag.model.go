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
	ID           string   `json:"id"`
	PostID       string `json:"post_id"`
	CommentID    *string `json:"comment_id,omitempty"`
	TagType      TagType `json:"tag_type"`
	TargetUserID *string `json:"target_user_id,omitempty"`
	Name         string  `json:"name"`
}

func NewTag(postID string, commentID *string, tagType TagType, targetUserID *string, name string) Tag {
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
