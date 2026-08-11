package validations

import (
	errorsapp "linkup/errors"
	"strings"
)

type BlockValidation struct{}

func NewBlockValidation() *BlockValidation {
	return &BlockValidation{}
}

func (v *BlockValidation) ValidateToggleBlock(userID, targetUserID string) error {
	if strings.TrimSpace(targetUserID) == "" {
		return errorsapp.New(errorsapp.ErrCodeBlockTargetRequired)
	}
	if userID == targetUserID {
		return errorsapp.New(errorsapp.ErrCodeBlockSelf)
	}
	return nil
}
