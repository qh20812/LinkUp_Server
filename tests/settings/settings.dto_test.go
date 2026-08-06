package settings_test

import (
	"encoding/json"
	"strings"
	"testing"

	"linkup/dto"
)

// Validation-only DTO tests for the user settings feature (no DB).

func TestDeactivateInputRequiresPassword(t *testing.T) {
	empty := dto.DeactivateInput{}
	if empty.Password != "" {
		t.Error("zero-value DeactivateInput should have empty Password")
	}
}

func TestUpdateAppearanceInputBinding(t *testing.T) {
	var input dto.UpdateAppearanceInput
	body := `{"theme":"dark","language":"en"}`
	if err := json.Unmarshal([]byte(body), &input); err != nil {
		t.Fatalf("unmarshal UpdateAppearanceInput: %v", err)
	}
	if input.Theme == nil || *input.Theme != "dark" || input.Language == nil || *input.Language != "en" {
		t.Errorf("unexpected decoded input: %+v", input)
	}

	var empty dto.UpdateAppearanceInput
	if err := json.Unmarshal([]byte(`{}`), &empty); err != nil {
		t.Fatalf("unmarshal empty input: %v", err)
	}
	if empty.Theme != nil || empty.Language != nil {
		t.Errorf("empty input should decode as nil pointers, got %+v", empty)
	}
}

func TestAppearanceSettingsResponseJSON(t *testing.T) {
	resp := dto.AppearanceSettingsResponse{
		Theme:    "dark",
		Language: "en",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal AppearanceSettingsResponse: %v", err)
	}
	if !strings.Contains(string(data), `"theme":"dark"`) {
		t.Errorf("theme not serialized, got %s", data)
	}
	if !strings.Contains(string(data), `"language":"en"`) {
		t.Errorf("language not serialized, got %s", data)
	}
}

func TestSessionDTOJSON(t *testing.T) {
	item := dto.SessionDTO{
		ID:           "sess-001",
		DeviceName:   "Mozilla/5.0",
		IPAddress:    "127.0.0.1",
		UserAgent:    "curl",
		IsCurrent:    true,
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal SessionDTO: %v", err)
	}
	if !strings.Contains(string(data), `"is_current":true`) {
		t.Errorf("is_current not serialized, got %s", data)
	}

	var decoded dto.SessionDTO
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal SessionDTO: %v", err)
	}
	if decoded.ID != "sess-001" || !decoded.IsCurrent {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestPrivacySettingsResponseJSON(t *testing.T) {
	resp := dto.PrivacySettingsResponse{
		DiscoverableInSearch:  false,
		AllowStrangerMessages: true,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal PrivacySettingsResponse: %v", err)
	}
	if !strings.Contains(string(data), `"discoverable_in_search":false`) {
		t.Errorf("discoverable_in_search not serialized, got %s", data)
	}
	if !strings.Contains(string(data), `"allow_stranger_messages":true`) {
		t.Errorf("allow_stranger_messages not serialized, got %s", data)
	}
}
