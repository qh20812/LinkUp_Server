package validations

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrChatContentRequired = errors.New("nội dung tin nhắn, emoji hoặc media là bắt buộc")
	ErrChatContentTooLong  = errors.New("nội dung tin nhắn không được vượt quá 2000 ký tự")
	ErrSelfInvite          = errors.New("không thể mời chính mình")
	ErrSearchKeywordEmpty  = errors.New("từ khóa tìm kiếm là bắt buộc")
	ErrDeleteModeInvalid   = errors.New("chế độ xóa phải là 'all' hoặc 'me'")

	ErrMessageNotFound     = errors.New("tin nhắn không tồn tại")
	ErrMessageAccessDenied = errors.New("bạn không có quyền truy cập tin nhắn này")
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
		return errors.New("bạn không phải người nhận lời mời này")
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
		return errors.New("bạn không phải người gửi tin nhắn này")
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
		return errors.New("không thể tạo chat trực tiếp với chính mình")
	}
	return nil
}
