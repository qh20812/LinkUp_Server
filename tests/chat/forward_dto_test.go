package chat_test

import (
	"encoding/json"
	"testing"

	"linkup/dto"
)

func TestSendMessagePayloadForwardedFrom(t *testing.T) {
	var p dto.SendMessagePayload
	if err := json.Unmarshal([]byte(`{"chat_id":"c1","content":"hello","forwarded_from":"m1"}`), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.ForwardedFrom == nil || *p.ForwardedFrom != "m1" {
		t.Fatalf("expected forwarded_from m1, got %v", p.ForwardedFrom)
	}
}

func TestSendMessagePayloadForwardedFromAbsent(t *testing.T) {
	var p dto.SendMessagePayload
	if err := json.Unmarshal([]byte(`{"chat_id":"c1","content":"hello"}`), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.ForwardedFrom != nil {
		t.Fatalf("expected nil forwarded_from, got %v", *p.ForwardedFrom)
	}
}

func TestGroupSendMessagePayloadForwardedFrom(t *testing.T) {
	var p dto.GroupSendMessagePayload
	if err := json.Unmarshal([]byte(`{"chat_id":"c1","content":"hello","forwarded_from":"m1"}`), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.ForwardedFrom == nil || *p.ForwardedFrom != "m1" {
		t.Fatalf("expected forwarded_from m1, got %v", p.ForwardedFrom)
	}
}