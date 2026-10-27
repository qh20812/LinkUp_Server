package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	errorsapp "linkup/errors"
)

// newTestContext tạo Gin context tối thiểu cho test handler (không DB/không route).
func newTestContext(method string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", nil)
	return c, w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not JSON (%q): %v", w.Body.String(), err)
	}
	return body
}

// TestNotificationHandlers_MissingUserID_NoPanic chặn tái phát lỗi: handler
// gọi `userID.(string)` trần nên panic khi middleware không gắn userID, và
// gin.Recovery() biến panic đó thành 500 body RỖNG — client chỉ thấy
// "HTTP 500", không có manh mối debug.
func TestNotificationHandlers_MissingUserID_NoPanic(t *testing.T) {
	ctrl := NewNotificationController(nil)

	handlers := map[string]gin.HandlerFunc{
		"GetNotifications":  ctrl.GetNotifications,
		"MarkAsRead":        ctrl.MarkAsRead,
		"MarkAllAsRead":     ctrl.MarkAllAsRead,
		"GetUnreadCount":    ctrl.GetUnreadCount,
		"GetSummary":        ctrl.GetSummary,
		"GetPreferences":    ctrl.GetPreferences,
		"UpdatePreferences": ctrl.UpdatePreferences,
		"RegisterPushToken": ctrl.RegisterPushToken,
	}

	for name, handler := range handlers {
		t.Run(name, func(t *testing.T) {
			c, w := newTestContext(http.MethodGet)

			// Không có userID trong context: phải trả lỗi, không được panic.
			handler(c)

			if w.Code != http.StatusInternalServerError {
				t.Fatalf("expected 500, got %d", w.Code)
			}
			body := decodeBody(t, w)
			if body["error"] != errorsapp.ErrCodeInternal {
				t.Fatalf("expected error code %q, got %v", errorsapp.ErrCodeInternal, body["error"])
			}
		})
	}
}

func TestNotificationHandlers_InvalidUserIDType_NoPanic(t *testing.T) {
	ctrl := NewNotificationController(nil)

	for name, value := range map[string]any{
		"wrong type": 42,
		"empty":      "",
		"nil":        nil,
	} {
		t.Run(name, func(t *testing.T) {
			c, w := newTestContext(http.MethodGet)
			c.Set("userID", value)

			ctrl.GetSummary(c)

			if w.Code != http.StatusInternalServerError {
				t.Fatalf("expected 500, got %d", w.Code)
			}
			if body := decodeBody(t, w); body["error"] != errorsapp.ErrCodeInternal {
				t.Fatalf("expected error code %q, got %v", errorsapp.ErrCodeInternal, body["error"])
			}
		})
	}
}

func TestAuthUserID(t *testing.T) {
	c, _ := newTestContext(http.MethodGet)

	if _, ok := authUserID(c); ok {
		t.Fatal("expected ok=false when userID is absent")
	}

	c.Set("userID", "user-123")
	id, ok := authUserID(c)
	if !ok || id != "user-123" {
		t.Fatalf("expected (user-123, true), got (%q, %v)", id, ok)
	}
}
