package notification_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"linkup/controllers"
)

// TestUpdatePreferencesInput_BindNewFields đảm bảo 3 cờ mới
// (story_react_enabled, share_enabled, media_enabled) được bind từ JSON.
func TestUpdatePreferencesInput_BindNewFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := `{"story_react_enabled":false,"share_enabled":false,"media_enabled":true}`
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPut, "/api/notifications/preferences", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var input controllers.UpdatePreferencesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		t.Fatalf("bind failed: %v", err)
	}

	if input.StoryReactEnabled == nil || *input.StoryReactEnabled {
		t.Fatalf("expected story_react_enabled=false, got %v", input.StoryReactEnabled)
	}
	if input.ShareEnabled == nil || *input.ShareEnabled {
		t.Fatalf("expected share_enabled=false, got %v", input.ShareEnabled)
	}
	if input.MediaEnabled == nil || !*input.MediaEnabled {
		t.Fatalf("expected media_enabled=true, got %v", input.MediaEnabled)
	}
	// Các cờ cũ không gửi → nil
	if input.LikeEnabled != nil || input.VoiceCallEnabled != nil {
		t.Fatal("expected legacy flags to remain nil")
	}
}

// TestUpdatePreferences_AllNilRejected: body rỗng {} → 400 (không chạm service).
func TestUpdatePreferences_AllNilRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := controllers.NewNotificationController(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/notifications/preferences", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", "u1")

	ctrl.UpdatePreferences(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty payload, got %d: %s", w.Code, w.Body.String())
	}
}

// TestUpdatePreferences_RejectsMalformedJSON: body hỏng → 400.
func TestUpdatePreferences_RejectsMalformedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := controllers.NewNotificationController(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/notifications/preferences", strings.NewReader(`{"like_enabled":`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", "u1")

	ctrl.UpdatePreferences(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed JSON, got %d", w.Code)
	}
}

// TestUpdatePreferencesInput_DecodesNull: JSON null → pointer nil (coi như không đổi).
func TestUpdatePreferencesInput_DecodesNull(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var input controllers.UpdatePreferencesInput
	raw := []byte(`{"share_enabled":null}`)
	if err := json.Unmarshal(raw, &input); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if input.ShareEnabled != nil {
		t.Fatal("expected share_enabled nil for JSON null")
	}
}