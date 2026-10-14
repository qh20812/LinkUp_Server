package community_test

import (
	"net/http"
	"testing"

	errorsapp "linkup/errors"
)

func TestCommunityStatusCodeMapping(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		// Validation → 400
		{errorsapp.ErrCodeCommunityNameRequired, http.StatusBadRequest},
		{errorsapp.ErrCodeCommunityNameTooShort, http.StatusBadRequest},
		{errorsapp.ErrCodeCommunityAvatarInvalid, http.StatusBadRequest},
		{errorsapp.ErrCodeCommunityInvalidFormat, http.StatusBadRequest},
		{errorsapp.ErrCodeInvalidRole, http.StatusBadRequest},
		{errorsapp.ErrCodeBackgroundRejected, http.StatusBadRequest},
		// Forbidden → 403
		{errorsapp.ErrCodeAdminCannotCreate, http.StatusForbidden},
		{errorsapp.ErrCodeNotCommunityAdmin, http.StatusForbidden},
		{errorsapp.ErrCodeCannotKickCreator, http.StatusForbidden},
		{errorsapp.ErrCodeOnlyCreatorCanTransfer, http.StatusForbidden},
		// Not found → 404
		{errorsapp.ErrCodeCommunityNotFound, http.StatusNotFound},
		{errorsapp.ErrCodeGroupChatNotFound, http.StatusNotFound},
		{errorsapp.ErrCodeMemberNotFound, http.StatusNotFound},
		{errorsapp.ErrCodeInviteCodeNotFound, http.StatusNotFound},
		// Conflict → 409 (incl. check-then-insert race on unique name)
		{errorsapp.ErrCodeCommunityNameExists, http.StatusConflict},
		{errorsapp.ErrCodeAlreadyMember, http.StatusConflict},
		{errorsapp.ErrCodeJoinRequestPending, http.StatusConflict},
		{errorsapp.ErrCodeInviteCodeMaxUses, http.StatusConflict},
		// Quota → 403, throttle → 429
		{errorsapp.ErrCodeCommunityCreateLimited, http.StatusForbidden},
		{errorsapp.ErrCodeCommunityCreateThrottled, http.StatusTooManyRequests},
		// Internal → 500
		{errorsapp.ErrCodeEncryptionKeyFailed, http.StatusInternalServerError},
		{errorsapp.ErrCodeBackgroundUploadFailed, http.StatusInternalServerError},
		{errorsapp.ErrCodeMemberCheckFailed, http.StatusInternalServerError},
		{errorsapp.ErrCodeInternal, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := errorsapp.StatusCode(tt.code); got != tt.want {
				t.Errorf("StatusCode(%q) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}
}
