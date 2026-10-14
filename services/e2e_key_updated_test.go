package services

import (
	"reflect"
	"sort"
	"testing"

	"linkup/models"
)

// TestKeyUpdateRecipients — event chat:e2e_key_updated chỉ gửi cho participant
// còn lại (trừ caller), gom theo chat, khử trùng lặp.
func TestKeyUpdateRecipients(t *testing.T) {
	keys := []models.ChatE2EKey{
		{ChatID: "chat-1", UserID: "user-a"},
		{ChatID: "chat-1", UserID: "user-b"},
		{ChatID: "chat-1", UserID: "user-b"}, // trùng lặp
		{ChatID: "chat-2", UserID: "user-a"},
		{ChatID: "chat-2", UserID: "user-c"},
	}

	got := keyUpdateRecipients(keys, "user-a")
	want := map[string][]string{
		"chat-1": {"user-b"},
		"chat-2": {"user-c"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("recipients = %+v (want %+v)", got, want)
	}
}

// TestKeyUpdateRecipientsEmpty — caller là người duy nhất (hoặc batch rỗng)
// thì không notify ai.
func TestKeyUpdateRecipientsEmpty(t *testing.T) {
	got := keyUpdateRecipients([]models.ChatE2EKey{
		{ChatID: "chat-1", UserID: "user-a"},
	}, "user-a")
	if len(got) != 0 {
		t.Fatalf("recipients = %+v (want empty)", got)
	}

	if got := keyUpdateRecipients(nil, "user-a"); len(got) != 0 {
		t.Fatalf("recipients = %+v (want empty)", got)
	}
}

// TestNotifyKeyUpdatedNilHub — chưa wire hub (test/service-only) thì bỏ qua,
// không panic.
func TestNotifyKeyUpdatedNilHub(t *testing.T) {
	svc := NewE2EService(nil, nil)
	svc.notifyKeyUpdated([]models.ChatE2EKey{
		{ChatID: "chat-1", UserID: "user-b"},
	}, "user-a")
}

// TestNotifyKeyUpdatedSortStable — thứ tự user trong cùng chat ổn định theo
// thứ tự input (deterministic, không phụ thuộc duyệt map).
func TestNotifyKeyUpdatedSortStable(t *testing.T) {
	keys := []models.ChatE2EKey{
		{ChatID: "chat-1", UserID: "user-c"},
		{ChatID: "chat-1", UserID: "user-b"},
	}
	got := keyUpdateRecipients(keys, "user-a")["chat-1"]
	want := []string{"user-c", "user-b"}
	sorted := append([]string(nil), got...)
	sort.Strings(sorted)
	if !reflect.DeepEqual(sorted, []string{"user-b", "user-c"}) {
		t.Fatalf("recipients = %v (want %v in any order)", got, want)
	}
}
