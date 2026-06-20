package validations

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrUsernameRequired       = errors.New("username is required")
	ErrUsernameTooShort       = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong        = errors.New("username must be at most 30 characters")
	ErrUsernameInvalid        = errors.New("username can only contain letters, numbers, underscores, and dots")
	ErrEmailRequired          = errors.New("email is required")
	ErrEmailInvalid           = errors.New("invalid email format")
	ErrPasswordRequired       = errors.New("password is required")
	ErrPasswordTooShort       = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong        = errors.New("password must be at most 128 characters")
	ErrPasswordMissingUpper   = errors.New("password must contain at least one uppercase letter")
	ErrPasswordMissingLower   = errors.New("password must contain at least one lowercase letter")
	ErrPasswordMissingDigit   = errors.New("password must contain at least one digit")
	ErrPasswordMissingSpecial = errors.New("password must contain at least one special character")
	ErrDisplayNameRequired    = errors.New("display name is required")
	ErrDisplayNameTooLong     = errors.New("display name must be at most 55 characters")
	ErrDisplayNameTooShort    = errors.New("display name must be at least 3 characters")
	ErrPasswordSameAsOld      = errors.New("new password must be different from old password")
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
	if len(password) > 128 {
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
