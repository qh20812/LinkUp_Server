package validations

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrUsernameRequired       = errors.New("tên người dùng không được để trống")
	ErrUsernameTooShort       = errors.New("tên người dùng phải có ít nhất 3 ký tự")
	ErrUsernameTooLong        = errors.New("tên người dùng không được vượt quá 30 ký tự")
	ErrUsernameInvalid        = errors.New("tên người dùng chỉ được chứa chữ cái, số, dấu gạch dưới và dấu chấm")
	ErrEmailRequired          = errors.New("email không được để trống")
	ErrEmailInvalid           = errors.New("định dạng email không hợp lệ")
	ErrPasswordRequired       = errors.New("mật khẩu không được để trống")
	ErrPasswordTooShort       = errors.New("mật khẩu phải có ít nhất 8 ký tự")
	ErrPasswordTooLong        = errors.New("mật khẩu không được vượt quá 50 ký tự")
	ErrPasswordMissingUpper   = errors.New("mật khẩu phải chứa ít nhất một chữ cái in hoa")
	ErrPasswordMissingLower   = errors.New("mật khẩu phải chứa ít nhất một chữ cái in thường")
	ErrPasswordMissingDigit   = errors.New("mật khẩu phải chứa ít nhất một chữ số")
	ErrPasswordMissingSpecial = errors.New("mật khẩu phải chứa ít nhất một ký tự đặc biệt")
	ErrDisplayNameRequired    = errors.New("tên hiển thị không được để trống")
	ErrDisplayNameTooLong     = errors.New("tên hiển thị không được vượt quá 55 ký tự")
	ErrDisplayNameTooShort    = errors.New("tên hiển thị phải có ít nhất 3 ký tự")
	ErrPasswordSameAsOld      = errors.New("mật khẩu mới phải khác mật khẩu cũ")
)

type AuthValidation struct{}

func NewAuthValidation() *AuthValidation {
	return &AuthValidation{}
}

func (v *AuthValidation) ValidateRegisterInput(displayName, email, password string) error {
	if err := v.ValidateDisplayName(displayName); err != nil {
		return err
	}
	if err := v.ValidateEmail(email); err != nil {
		return err
	}
	return v.ValidatePassword(password)
}

func (v *AuthValidation) ValidateLoginInput(email, password string) error {
	if err := v.ValidateEmail(email); err != nil {
		return err
	}
	return v.ValidatePassword(password)
}

func (v *AuthValidation) ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return ErrUsernameRequired
	}
	if len(username) < 3 {
		return ErrUsernameTooShort
	}
	if len(username) > 30 {
		return ErrUsernameTooLong
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_.]+$`, username)
	if !matched {
		return ErrUsernameInvalid
	}
	return nil
}

func (v *AuthValidation) ValidateDisplayName(displayName string) error {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return ErrDisplayNameRequired
	}
	if len([]rune(displayName)) > 55 {
		return ErrDisplayNameTooLong
	}
	if len([]rune(displayName)) < 3 {
		return ErrDisplayNameTooShort
	}
	return nil
}

func (v *AuthValidation) ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrEmailRequired
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`, email)
	if !matched {
		return ErrEmailInvalid
	}
	return nil
}

func (v *AuthValidation) ValidatePassword(password string) error {
	if password == "" {
		return ErrPasswordRequired
	}
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	if len(password) > 50 {
		return ErrPasswordTooLong
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return ErrPasswordMissingUpper
	}
	if !hasLower {
		return ErrPasswordMissingLower
	}
	if !hasDigit {
		return ErrPasswordMissingDigit
	}
	if !hasSpecial {
		return ErrPasswordMissingSpecial
	}
	return nil
}

func (v *AuthValidation) NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (v *AuthValidation) NormalizeUsername(username string) string {
	return strings.TrimSpace(username)
}
