package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"linkup/dto"
)

// Validation-only DTO binding tests cho E2E endpoints (không cần DB).
// Pattern theo tests/chat/forward_dto_test.go + tests/notification.

func init() {
	gin.SetMode(gin.TestMode)
}

func bindBody[T any](t *testing.T, body string, out *T) error {
	t.Helper()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c.ShouldBindJSON(out)
}

func TestRekeyChatKeyRequest_Valid(t *testing.T) {
	var req dto.RekeyChatKeyRequest
	if err := bindBody(t, `{"wrapped_key":"AAAA","nonce":"BBBB"}`, &req); err != nil {
		t.Fatalf("valid payload rejected: %v", err)
	}
	if req.WrappedKey != "AAAA" || req.Nonce != "BBBB" {
		t.Fatalf("unexpected decoded: %+v", req)
	}
}

func TestRekeyChatKeyRequest_MissingWrappedKeyRejected(t *testing.T) {
	var req dto.RekeyChatKeyRequest
	if err := bindBody(t, `{"nonce":"BBBB"}`, &req); err == nil {
		t.Fatal("empty wrapped_key must fail binding (binding:required)")
	}
}

func TestRekeyChatKeyRequest_EmptyBodyRejected(t *testing.T) {
	var req dto.RekeyChatKeyRequest
	if err := bindBody(t, `{}`, &req); err == nil {
		t.Fatal("empty body must fail binding (binding:required)")
	}
}

func TestChatE2EKeyBatchRequest_Valid(t *testing.T) {
	var req dto.ChatE2EKeyBatchRequest
	body := `{"keys":[{"chat_id":"c1","user_id":"u2","wrapped_key":"AAAA","nonce":"N"}]}`
	if err := bindBody(t, body, &req); err != nil {
		t.Fatalf("valid payload rejected: %v", err)
	}
	if len(req.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(req.Keys))
	}
	k := req.Keys[0]
	if k.ChatID != "c1" || k.UserID != "u2" || k.WrappedKey != "AAAA" || k.Nonce != "N" {
		t.Fatalf("unexpected decoded: %+v", k)
	}
}

func TestChatE2EKeyBatchRequest_MissingKeysRejected(t *testing.T) {
	var req dto.ChatE2EKeyBatchRequest
	if err := bindBody(t, `{}`, &req); err == nil {
		t.Fatal("empty body must fail binding (binding:required on Keys)")
	}
}

func TestChatE2EKeyBatchRequest_EntryMissingFieldsRejected(t *testing.T) {
	var req dto.ChatE2EKeyBatchRequest
	// Mỗi entry yêu cầu chat_id, user_id, wrapped_key — thiếu wrapped_key → lỗi.
	body := `{"keys":[{"chat_id":"c1","user_id":"u2"}]}`
	if err := bindBody(t, body, &req); err == nil {
		t.Fatal("entry missing wrapped_key must fail binding")
	}
}

func TestRegisterE2EKeyRequest_RequiresPublicKey(t *testing.T) {
	var ok dto.RegisterE2EKeyRequest
	if err := bindBody(t, `{"public_key":"SPKI..."}`, &ok); err != nil {
		t.Fatalf("valid payload rejected: %v", err)
	}
	var empty dto.RegisterE2EKeyRequest
	if err := bindBody(t, `{}`, &empty); err == nil {
		t.Fatal("missing public_key must fail binding (binding:required)")
	}
}

func TestUnlockRecoveryRequest_RequiresCheck(t *testing.T) {
	// JSON {} không reset các field của struct — phải dùng biến mới để đảm bảo
	// trạng thái zero-value như request thật.
	var req dto.UnlockRecoveryRequest
	if err := bindBody(t, `{}`, &req); err == nil {
		t.Fatal("missing check must fail binding (binding:required)")
	}
}

func TestPutRecoveryRequest_RequiresAllFields(t *testing.T) {
	var req dto.PutRecoveryRequest
	body := `{"salt":"S","blob":"B","pin_check":"PC","recovery_check":"RC"}`
	if err := bindBody(t, body, &req); err != nil {
		t.Fatalf("valid payload rejected: %v", err)
	}
	missing := dto.PutRecoveryRequest{}
	if err := bindBody(t, `{"salt":"S","blob":"B","pin_check":"PC"}`, &missing); err == nil {
		t.Fatal("missing recovery_check must fail binding (binding:required)")
	}
}