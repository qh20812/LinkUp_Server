package validations

import (
	errorsapp "linkup/errors"
	"linkup/models"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrCommunityNameRequired      = errorsapp.New(errorsapp.ErrCodeCommunityNameRequired)
	ErrCommunityNameMinLength     = errorsapp.New(errorsapp.ErrCodeCommunityNameTooShort)
	ErrCommunityNameMaxLength     = errorsapp.New(errorsapp.ErrCodeCommunityNameTooLong)
	ErrCommunityDescMaxLength     = errorsapp.New(errorsapp.ErrCodeCommunityDescTooLong)
	ErrCommunityAvatarInvalid     = errorsapp.New(errorsapp.ErrCodeCommunityAvatarInvalid)
	ErrCommunityBackgroundInvalid = errorsapp.New(errorsapp.ErrCodeCommunityBackgroundInvalid)
	ErrCommunityNameExists        = errorsapp.New(errorsapp.ErrCodeCommunityNameExists)
	ErrAlreadyMember              = errorsapp.New(errorsapp.ErrCodeAlreadyMember)
	ErrJoinRequestPending         = errorsapp.New(errorsapp.ErrCodeJoinRequestPending)
	ErrJoinRequestNotFound        = errorsapp.New(errorsapp.ErrCodeJoinRequestNotFound)
	ErrJoinRequestAlreadyHandled  = errorsapp.New(errorsapp.ErrCodeJoinRequestHandled)
	ErrNotCommunityAdmin          = errorsapp.New(errorsapp.ErrCodeNotCommunityAdmin)
	ErrNotCommunityMember         = errorsapp.New(errorsapp.ErrCodeNotCommunityMember)
	ErrCommunityNotFound          = errorsapp.New(errorsapp.ErrCodeCommunityNotFound)
	ErrMemberNotFound             = errorsapp.New(errorsapp.ErrCodeMemberNotFound)
	ErrInvalidRole                = errorsapp.New(errorsapp.ErrCodeInvalidRole)
	ErrCannotChangeOwnRole       = errorsapp.New(errorsapp.ErrCodeCannotChangeOwnRole)
	ErrCannotTargetAdmin          = errorsapp.New(errorsapp.ErrCodeCannotTargetAdmin)
	ErrCreatorCannotLeave        = errorsapp.New(errorsapp.ErrCodeCreatorCannotLeave)
	ErrCannotKickCreator         = errorsapp.New(errorsapp.ErrCodeCannotKickCreator)
	ErrCannotKickAdmin           = errorsapp.New(errorsapp.ErrCodeCannotKickAdmin)
	ErrKickReasonRequired        = errorsapp.New(errorsapp.ErrCodeKickReasonRequired)
	ErrKickReasonTooShort        = errorsapp.New(errorsapp.ErrCodeKickReasonTooShort)
	ErrKickReasonTooLong         = errorsapp.New(errorsapp.ErrCodeKickReasonTooLong)
	ErrInviteCodeNotFound        = errorsapp.New(errorsapp.ErrCodeInviteCodeNotFound)
	ErrInviteCodeExpired         = errorsapp.New(errorsapp.ErrCodeInviteCodeExpired)
	ErrInviteCodeInactive        = errorsapp.New(errorsapp.ErrCodeInviteCodeInactive)
	ErrInviteCodeMaxUses         = errorsapp.New(errorsapp.ErrCodeInviteCodeMaxUses)
	ErrInvitationNotFound        = errorsapp.New(errorsapp.ErrCodeInvitationNotFound)
	ErrInvitationAlreadyHandled  = errorsapp.New(errorsapp.ErrCodeInvitationHandled)
	ErrCannotInviteSelf          = errorsapp.New(errorsapp.ErrCodeCannotInviteSelf)
)

type CommunityValidation struct{}

func NewCommunityValidation() *CommunityValidation {
	return &CommunityValidation{}
}

func (v *CommunityValidation) ValidateCreateCommunity(name, description, avatarURI string) error {
	if err := v.ValidateName(name); err != nil {
		return err
	}
	if err := v.ValidateDescription(description); err != nil {
		return err
	}
	if avatarURI != "" {
		if err := v.ValidateAvatarURI(avatarURI); err != nil {
			return err
		}
	}
	return nil
}

func (v *CommunityValidation) ValidateUpdateCommunity(name string, description *string) error {
	if name != "" {
		if err := v.ValidateName(name); err != nil {
			return err
		}
	}
	if description != nil {
		if err := v.ValidateDescription(*description); err != nil {
			return err
		}
	}
	return nil
}

func (v *CommunityValidation) ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrCommunityNameRequired
	}
	if utf8.RuneCountInString(name) < 3 {
		return ErrCommunityNameMinLength
	}
	if utf8.RuneCountInString(name) > 100 {
		return ErrCommunityNameMaxLength
	}
	return nil
}

func (v *CommunityValidation) ValidateDescription(description string) error {
	description = strings.TrimSpace(description)
	if description == "" {
		return nil
	}
	if utf8.RuneCountInString(description) > 500 {
		return ErrCommunityDescMaxLength
	}
	return nil
}

func (v *CommunityValidation) ValidateAvatarURI(avatarURI string) error {
	avatarURI = strings.TrimSpace(avatarURI)
	if avatarURI == "" {
		return nil
	}
	parsed, err := url.ParseRequestURI(avatarURI)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ErrCommunityAvatarInvalid
	}
	return nil
}

func (v *CommunityValidation) ValidateBackgroundURI(backgroundURI string) error {
	backgroundURI = strings.TrimSpace(backgroundURI)
	if backgroundURI == "" {
		return nil
	}
	parsed, err := url.ParseRequestURI(backgroundURI)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ErrCommunityBackgroundInvalid
	}
	return nil
}

func (v *CommunityValidation) ValidateUpdateRole(role string) error {
	switch role {
	case "GROUP_ADMIN", "GROUP_MOD", "GROUP_MEMBER":
		return nil
	default:
		return ErrInvalidRole
	}
}

func (v *CommunityValidation) ValidateInviteCode(code *models.CommunityInviteCode) error {
	if code == nil {
		return ErrInviteCodeNotFound
	}
	if !code.IsActive {
		return ErrInviteCodeInactive
	}
	if code.ExpiresAt != nil && time.Now().UTC().After(*code.ExpiresAt) {
		return ErrInviteCodeExpired
	}
	if code.MaxUses > 0 && code.UsedCount >= code.MaxUses {
		return ErrInviteCodeMaxUses
	}
	return nil
}
