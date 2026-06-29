package validations

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrChatContentRequired = errors.New("message content, emoji, or media is required")
	ErrChatContentTooLong  = errors.New("message content must be at most 2000 characters")
	ErrSelfInvite          = errors.New("cannot invite yourself")
	ErrSearchKeywordEmpty  = errors.New("search keyword is required")
	ErrDeleteModeInvalid   = errors.New("delete mode must be 'all' or 'me'")
)

type ChatValidation struct{}

func NewChatValidation() *ChatValidation {
	return &ChatValidation{}
}

func (v *ChatValidation) ValidateSendMessage(content string, emojiID, mediaID *string) error {
	if strings.TrimSpace(content) == "" && emojiID == nil && mediaID == nil {
		return ErrChatContentRequired
	}
	if utf8.RuneCountInString(content) > 2000 {
		return ErrChatContentTooLong
	}
	return nil
}

func (v *ChatValidation) ValidateRequestChatInvite(userID, targetUserID string) error {
	if userID == targetUserID {
		return ErrSelfInvite
	}
	return nil
}

func (v *ChatValidation) ValidateResponseChatInvite(userID, targetID string) error {
	if userID != targetID {
		return errors.New("you are not the recipient of this invite")
	}
	return nil
}

func (v *ChatValidation) ValidateSearchMessages(keyword string) error {
	if strings.TrimSpace(keyword) == "" {
		return ErrSearchKeywordEmpty
	}
	return nil
}

func (v *ChatValidation) ValidateDeleteMessage(senderID, userID string, deletedForSender, deletedForReceiver bool) error {
	if senderID != userID {
		return errors.New("you are not the sender of this message")
	}
	if deletedForSender || deletedForReceiver {
		return errors.New("message already deleted")
	}
	return nil
}

func (v *ChatValidation) ValidateDeleteMode(mode string) error {
	if mode != "" && !strings.EqualFold(mode, "all") && !strings.EqualFold(mode, "me") {
		return ErrDeleteModeInvalid
	}
	return nil
}

func (v *ChatValidation) ValidateDirectChat(userID, targetUserID string) error {
	if userID == targetUserID {
		return errors.New("cannot create direct chat with yourself")
	}
	return nil
}
