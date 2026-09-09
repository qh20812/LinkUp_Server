package validations

import (
	errorsapp "linkup/errors"
	"strings"
	"unicode/utf8"
)

type ChatValidation struct{}

func NewChatValidation() *ChatValidation {
	return &ChatValidation{}
}

func (v *ChatValidation) ValidateSendMessage(content string, emojiID, mediaID *string) error {
	if strings.TrimSpace(content) == "" && emojiID == nil && mediaID == nil {
		return errorsapp.New(errorsapp.ErrCodeChatContentRequired)
	}
	if utf8.RuneCountInString(content) > 2000 {
		return errorsapp.New(errorsapp.ErrCodeChatContentTooLong)
	}
	return nil
}

func (v *ChatValidation) ValidateRequestChatInvite(userID, targetUserID string) error {
	if userID == targetUserID {
		return errorsapp.New(errorsapp.ErrCodeChatSelfChat)
	}
	return nil
}

func (v *ChatValidation) ValidateResponseChatInvite(userID, targetID string) error {
	if userID != targetID {
		return errorsapp.New(errorsapp.ErrCodeChatNotRecipient)
	}
	return nil
}

func (v *ChatValidation) ValidateSearchMessages(keyword string) error {
	if strings.TrimSpace(keyword) == "" {
		return errorsapp.New(errorsapp.ErrCodeChatSearchEmpty)
	}
	return nil
}

func (v *ChatValidation) ValidateDeleteMessage(senderID, userID, mode string) error {
	if strings.EqualFold(mode, "all") && senderID != userID {
		return errorsapp.New(errorsapp.ErrCodeChatNotSender)
	}
	return nil
}

func (v *ChatValidation) ValidateDeleteMode(mode string) error {
	if mode != "" && !strings.EqualFold(mode, "all") && !strings.EqualFold(mode, "me") {
		return errorsapp.New(errorsapp.ErrCodeChatDeleteModeInvalid)
	}
	return nil
}

func (v *ChatValidation) ValidateDirectChat(userID, targetUserID string) error {
	if userID == targetUserID {
		return errorsapp.New(errorsapp.ErrCodeChatSelfChat)
	}
	return nil
}
