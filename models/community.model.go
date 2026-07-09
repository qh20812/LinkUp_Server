package models

import (
	"strings"
	"time"
)

type CommunityPrivacy string

const (
	PrivacyPublic         CommunityPrivacy = "public"
	PrivacyCode           CommunityPrivacy = "code"
	PrivacyInvitationOnly CommunityPrivacy = "invitation_only"
)

type CommunityStatus string

const (
	CommunityStatusActive   CommunityStatus = "active"
	CommunityStatusHidden   CommunityStatus = "hidden"
	CommunityStatusArchived CommunityStatus = "archived"
)

type Community struct {
	ID            string           `json:"id"`
	CreatorID     string           `json:"creator_id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	AvatarURI     string           `json:"avatar_uri"`
	BackgroundURI string           `json:"background_uri"`
	AutoApprove   bool             `json:"auto_approve"`
	Privacy       CommunityPrivacy `json:"privacy" gorm:"type:varchar(20);default:public"`
	Status        CommunityStatus  `json:"status" gorm:"type:varchar(20);default:active"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     *time.Time       `json:"updated_at,omitempty"`
}

func NewCommunity(creatorID, name, description, avatarURI string) Community {
	return Community{
		CreatorID:     creatorID,
		Name:          name,
		Description:   description,
		AvatarURI:     avatarURI,
		BackgroundURI: "",
		Privacy:       PrivacyPublic,
		Status:        CommunityStatusActive,
	}
}

func (s CommunityStatus) String() string {
	return string(s)
}

func ParseCommunityStatus(value string) CommunityStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(CommunityStatusHidden):
		return CommunityStatusHidden
	case string(CommunityStatusArchived):
		return CommunityStatusArchived
	default:
		return CommunityStatusActive
	}
}
