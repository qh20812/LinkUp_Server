package validations

import (
	"errors"
	"strings"
)

var (
	ErrTargetUserRequired = errors.New("target_user_id là bắt buộc")
	ErrSelfBlock          = errors.New("không thể chặn chính mình")
)

type BlockValidation struct{}

func NewBlockValidation() *BlockValidation {
	return &BlockValidation{}
}

func (v *BlockValidation) ValidateToggleBlock(userID, targetUserID string) error {
	if strings.TrimSpace(targetUserID) == "" {
		return ErrTargetUserRequired
	}
	if userID == targetUserID {
		return ErrSelfBlock
	}
	return nil
}
