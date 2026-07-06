package validations

import (
	"errors"
	"linkup/models"
	"net/url"
	"strings"
	"time"
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
	ErrNotCommunityMember         = errors.New("bạn không phải thành viên của cộng đồng này")
	ErrCommunityNotFound          = errors.New("cộng đồng không tồn tại")
	ErrMemberNotFound             = errors.New("thành viên không tồn tại trong cộng đồng")
	ErrInvalidRole                = errors.New("vai trò không hợp lệ")
	ErrCannotChangeOwnRole       = errors.New("không thể thay đổi vai trò của chính mình")
	ErrCannotTargetAdmin          = errors.New("không thể thay đổi vai trò của quản trị viên")
	ErrCreatorCannotLeave        = errors.New("người tạo cộng đồng không thể rời đi, vui lòng chuyển quyền trước")
	ErrCannotKickCreator         = errors.New("không thể đuổi người tạo cộng đồng")
	ErrCannotKickAdmin           = errors.New("chỉ người tạo cộng đồng mới có quyền đuổi quản trị viên")
	ErrKickReasonRequired        = errors.New("lý do là bắt buộc")
	ErrKickReasonTooShort        = errors.New("lý do phải có ít nhất 3 ký tự")
	ErrKickReasonTooLong         = errors.New("lý do không được vượt quá 500 ký tự")
	ErrInviteCodeNotFound        = errors.New("mã mời không tồn tại")
	ErrInviteCodeExpired         = errors.New("mã mời đã hết hạn")
	ErrInviteCodeInactive        = errors.New("mã mời đã bị vô hiệu hóa")
	ErrInviteCodeMaxUses         = errors.New("mã mời đã đạt số lần sử dụng tối đa")
	ErrInvitationNotFound        = errors.New("lời mời không tồn tại")
	ErrInvitationAlreadyHandled  = errors.New("lời mời đã được xử lý")
	ErrCannotInviteSelf          = errors.New("không thể mời chính mình")
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
