package validations

import (
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"
)

var (
	ErrCommunityNameRequired      = errors.New("tên cộng đồng không được để trống")
	ErrCommunityNameMinLength     = errors.New("tên cộng đồng phải có ít nhất 3 ký tự")
	ErrCommunityNameMaxLength     = errors.New("tên cộng đồng không được vượt quá 100 ký tự")
	ErrCommunityDescMaxLength     = errors.New("mô tả cộng đồng không được vượt quá 500 ký tự")
	ErrCommunityAvatarInvalid     = errors.New("avatar URI không hợp lệ")
	ErrCommunityBackgroundInvalid = errors.New("background URI không hợp lệ")
	ErrCommunityNameExists        = errors.New("tên cộng đồng đã tồn tại")
	ErrAlreadyMember              = errors.New("bạn đã là thành viên của cộng đồng này")
	ErrJoinRequestPending         = errors.New("bạn đã có yêu cầu tham gia đang chờ xử lý")
	ErrJoinRequestNotFound        = errors.New("yêu cầu tham gia không tồn tại")
	ErrJoinRequestAlreadyHandled  = errors.New("yêu cầu tham gia đã được xử lý")
	ErrNotCommunityAdmin          = errors.New("bạn không phải quản trị viên của cộng đồng này")
	ErrCommunityNotFound          = errors.New("cộng đồng không tồn tại")
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
