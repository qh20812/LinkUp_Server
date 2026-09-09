package validations

import (
	errorsapp "linkup/errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrUsernameRequired       = errorsapp.New("auth.USERNAME_REQUIRED")
	ErrUsernameTooShort       = errorsapp.Newf("auth.USERNAME_TOO_SHORT", map[string]any{"min": 3})
	ErrUsernameTooLong        = errorsapp.Newf("auth.USERNAME_TOO_LONG", map[string]any{"max": 30})
	ErrUsernameInvalid        = errorsapp.New("auth.USERNAME_INVALID")
	ErrEmailRequired          = errorsapp.New("auth.EMAIL_REQUIRED")
	ErrEmailInvalid           = errorsapp.New("auth.EMAIL_INVALID")
	ErrPasswordRequired       = errorsapp.New("auth.PASSWORD_REQUIRED")
	ErrPasswordTooShort       = errorsapp.Newf("auth.PASSWORD_TOO_SHORT", map[string]any{"min": 8})
	ErrPasswordTooLong        = errorsapp.Newf("auth.PASSWORD_TOO_LONG", map[string]any{"max": 50})
	ErrPasswordMissingUpper   = errorsapp.New("auth.PASSWORD_MISSING_UPPER")
	ErrPasswordMissingLower   = errorsapp.New("auth.PASSWORD_MISSING_LOWER")
	ErrPasswordMissingDigit   = errorsapp.New("auth.PASSWORD_MISSING_DIGIT")
	ErrPasswordMissingSpecial = errorsapp.New("auth.PASSWORD_MISSING_SPECIAL")
	ErrDisplayNameRequired    = errorsapp.New("auth.DISPLAY_NAME_REQUIRED")
	ErrDisplayNameTooShort    = errorsapp.Newf("auth.DISPLAY_NAME_TOO_SHORT", map[string]any{"min": 3})
	ErrDisplayNameTooLong     = errorsapp.Newf("auth.DISPLAY_NAME_TOO_LONG", map[string]any{"max": 55})
	ErrPasswordSameAsOld      = errorsapp.New("auth.PASSWORD_SAME_AS_OLD")
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
