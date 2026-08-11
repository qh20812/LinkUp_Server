package errors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RespondError sends an error response with code + message + optional params.
func RespondError(c *gin.Context, status int, err *AppError) {
	if err.HasParams() {
		c.JSON(status, gin.H{
			"error":   err.Code,
			"message": err.Message,
			"params":  err.Params,
		})
	} else {
		c.JSON(status, gin.H{
			"error":   err.Code,
			"message": err.Message,
		})
	}
}

// RespondErrorf sends an error response with rendered message from params.
func RespondErrorf(c *gin.Context, status int, code string, params map[string]any) {
	err := Newf(code, params)
	RespondError(c, status, err)
}

// Respond sends an error response from a plain error.
// If err is an *AppError, it uses the structured format.
// Otherwise, it falls back to a generic HTTP error.
func Respond(c *gin.Context, status int, err error) {
	if appErr, ok := err.(*AppError); ok {
		RespondError(c, status, appErr)
		return
	}
	c.JSON(status, gin.H{
		"error":   fmt.Sprintf("HTTP_%d", status),
		"message": err.Error(),
	})
}

// RenderTemplate renders {{param}} placeholders in a message template.
func RenderTemplate(template string, params map[string]any) string {
	result := template
	for k, v := range params {
		result = strings.ReplaceAll(result, fmt.Sprintf("{{%s}}", k), fmt.Sprintf("%v", v))
	}
	return result
}

// IsAppError checks if an error is an AppError and returns it.
func IsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}

// Wrap wraps a non-AppError into an AppError with a given code.
func Wrap(code string, err error) *AppError {
	return &AppError{Code: code, Message: err.Error()}
}

// Wrapf wraps a non-AppError into an AppError with code and params.
func Wrapf(code string, err error, params map[string]any) *AppError {
	msg := RenderTemplate(Messages[code], params)
	return &AppError{Code: code, Message: msg, Params: params}
}

// StatusCode returns the appropriate HTTP status code for an error code prefix.
func StatusCode(code string) int {
	switch {
	case strings.HasPrefix(code, "auth."):
		return http.StatusUnauthorized
	case strings.HasPrefix(code, "post.NOT_FOUND") || strings.HasPrefix(code, "post.POST_NOT_FOUND"):
		return http.StatusNotFound
	case strings.HasPrefix(code, "post."):
		return http.StatusBadRequest
	case strings.HasPrefix(code, "common.INVALID"):
		return http.StatusBadRequest
	case strings.HasPrefix(code, "common.UNAUTHORIZED"):
		return http.StatusUnauthorized
	case strings.HasPrefix(code, "common.FORBIDDEN"):
		return http.StatusForbidden
	case strings.HasPrefix(code, "common.NOT_FOUND"):
		return http.StatusNotFound
	case strings.HasPrefix(code, "group_chat."):
		switch {
		case strings.HasSuffix(code, "NOT_FOUND") || strings.HasSuffix(code, "REPLY_NOT_FOUND") || strings.HasSuffix(code, "EMOJI_NOT_FOUND") || strings.HasSuffix(code, "MEDIA_NOT_FOUND"):
			return http.StatusNotFound
		case strings.HasSuffix(code, "BANNED") || strings.HasSuffix(code, "ADMIN_ONLY") || strings.HasSuffix(code, "NOT_MEMBER") || strings.HasSuffix(code, "SELF_BAN"):
			return http.StatusForbidden
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "chat."):
		switch {
		case strings.HasSuffix(code, "MESSAGE_NOT_FOUND"):
			return http.StatusNotFound
		case strings.HasSuffix(code, "ACCESS_DENIED") || strings.HasSuffix(code, "NOT_PARTICIPANT"):
			return http.StatusForbidden
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "friend."):
		switch {
		case strings.HasSuffix(code, "NOT_FOUND"):
			return http.StatusNotFound
		case strings.HasSuffix(code, "NOT_AUTHORIZED") || strings.HasSuffix(code, "ADMIN_RESTRICTED"):
			return http.StatusForbidden
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "follow."):
		switch {
		case strings.HasSuffix(code, "USER_NOT_FOUND"):
			return http.StatusNotFound
		case strings.HasSuffix(code, "ADMIN_RESTRICTED") || strings.HasSuffix(code, "SUPERADMIN_RESTRICTED"):
			return http.StatusForbidden
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "block."):
		switch {
		case strings.HasSuffix(code, "ADMIN_RESTRICTED"):
			return http.StatusForbidden
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "media."):
		switch {
		case strings.HasSuffix(code, "NOT_FOUND"):
			return http.StatusNotFound
		case strings.HasSuffix(code, "FORBIDDEN"):
			return http.StatusForbidden
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "contribution."):
		switch {
		case strings.HasSuffix(code, "NOT_FOUND") || strings.HasSuffix(code, "CHALLENGE_NOT_FOUND"):
			return http.StatusNotFound
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "call."):
		switch {
		case strings.HasSuffix(code, "NOT_FOUND"):
			return http.StatusNotFound
		case strings.HasSuffix(code, "NOT_PARTICIPANT"):
			return http.StatusForbidden
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "moderation."):
		return http.StatusInternalServerError
	case strings.HasPrefix(code, "profile."):
		switch {
		case strings.HasSuffix(code, "NOT_FOUND"):
			return http.StatusNotFound
		case strings.HasSuffix(code, "PRIVATE"):
			return http.StatusForbidden
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "password_reset."):
		switch {
		case strings.HasSuffix(code, "NOT_FOUND"):
			return http.StatusNotFound
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "email_verification."):
		switch {
		case strings.HasSuffix(code, "NOT_FOUND"):
			return http.StatusNotFound
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "user_settings."):
		switch {
		case strings.HasSuffix(code, "SESSION_NOT_FOUND"):
			return http.StatusNotFound
		case strings.HasSuffix(code, "WRONG_PASSWORD"):
			return http.StatusUnauthorized
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "story."):
		switch {
		case strings.HasSuffix(code, "NOT_FOUND") || strings.HasSuffix(code, "INTERACT_NOT_FOUND"):
			return http.StatusNotFound
		case strings.HasSuffix(code, "FORBIDDEN"):
			return http.StatusForbidden
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "search."):
		return http.StatusBadRequest
	case strings.HasPrefix(code, "ad."):
		switch {
		case strings.HasSuffix(code, "NOT_FOUND") || strings.HasSuffix(code, "NOT_UPDATED") || strings.HasSuffix(code, "NOT_DELETED"):
			return http.StatusNotFound
		case strings.HasSuffix(code, "NOT_OWNER"):
			return http.StatusForbidden
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "package."):
		switch {
		case strings.HasSuffix(code, "NOT_SUBSCRIBED"):
			return http.StatusNotFound
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "notification."):
		return http.StatusInternalServerError
	case strings.HasPrefix(code, "admin."):
		switch {
		case strings.HasSuffix(code, "NOT_FOUND"):
			return http.StatusNotFound
		case strings.HasSuffix(code, "NO_ACCESS") || strings.HasSuffix(code, "NOT_SUPERADMIN"):
			return http.StatusForbidden
		default:
			return http.StatusBadRequest
		}
	case strings.HasPrefix(code, "rbac."):
		switch {
		case strings.HasSuffix(code, "AUTHENTICATION_REQUIRED"):
			return http.StatusUnauthorized
		case strings.HasSuffix(code, "PERMISSION_DENIED") || strings.HasSuffix(code, "CONTRIBUTION_LEVEL_INSUFFICIENT") || strings.HasSuffix(code, "AD_ACCESS_DENIED"):
			return http.StatusForbidden
		case strings.HasSuffix(code, "AD_NOT_FOUND"):
			return http.StatusNotFound
		default:
			return http.StatusForbidden
		}
	case strings.HasPrefix(code, "ban."):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
