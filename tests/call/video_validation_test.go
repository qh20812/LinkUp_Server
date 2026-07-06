package call_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"linkup/dto"
	"linkup/models"
)

// ─── Model tests ─────────────────────────────────────────────────

func TestParseCallType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  models.CallType
	}{
		{"empty string", "", models.CallTypeVoice},
		{"voice lowercase", "voice", models.CallTypeVoice},
		{"video lowercase", "video", models.CallTypeVideo},
		{"voice uppercase", "VOICE", models.CallTypeVoice},
		{"video mixed case", "Video", models.CallTypeVideo},
		{"unknown type fallback", "fax", models.CallTypeVoice},
		{"whitespace trimmed", "  video  ", models.CallTypeVideo},
		{"whitespace only", "   ", models.CallTypeVoice},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.ParseCallType(tt.input)
			if got != tt.want {
				t.Errorf("ParseCallType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewCallVideoDefaults(t *testing.T) {
	call := models.NewCall("caller-uuid", "callee-uuid", models.CallTypeVideo, false)

	if call.VideoEnabledCaller {
		t.Error("VideoEnabledCaller should default to false")
	}
	if call.VideoEnabledCallee {
		t.Error("VideoEnabledCallee should default to false")
	}
	if call.CallType != models.CallTypeVideo {
		t.Errorf("CallType = %q, want %q", call.CallType, models.CallTypeVideo)
	}
	if call.Status != models.CallStatusCalling {
		t.Errorf("Status = %q, want %q", call.Status, models.CallStatusCalling)
	}
	if call.MutedCaller {
		t.Error("MutedCaller should default to false")
	}
	if call.MutedCallee {
		t.Error("MutedCallee should default to false")
	}
}

func TestNewCallEmptyTypeDefaultsToVoice(t *testing.T) {
	call := models.NewCall("caller-uuid", "callee-uuid", "", false)

	if call.CallType != models.CallTypeVoice {
		t.Errorf("empty CallType → %q, want %q", call.CallType, models.CallTypeVoice)
	}
}

func TestNewCallSetsCreatedAt(t *testing.T) {
	before := time.Now().UTC()
	call := models.NewCall("a", "b", models.CallTypeVoice, false)
	after := time.Now().UTC()

	if call.CreatedAt.Before(before) || call.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v not in range [%v, %v]", call.CreatedAt, before, after)
	}
}

func TestParseCallStatus(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  models.CallStatus
	}{
		{"calling", "calling", models.CallStatusCalling},
		{"ringing", "ringing", models.CallStatusRinging},
		{"connected", "connected", models.CallStatusConnected},
		{"ended", "ended", models.CallStatusEnded},
		{"missed", "missed", models.CallStatusMissed},
		{"rejected", "rejected", models.CallStatusRejected},
		{"busy", "busy", models.CallStatusBusy},
		{"unknown fallback to ended", "invalid", models.CallStatusEnded},
		{"case insensitive", "CONNECTED", models.CallStatusConnected},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.ParseCallStatus(tt.input)
			if got != tt.want {
				t.Errorf("ParseCallStatus(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ─── ICE Server URL parsing tests ────────────────────────────────
//
// Test việc parse biến môi trường ICE_SERVER_URLS (comma-separated)
// thành danh sách IceServer struct. Logic này tương tự controller
// GetIceServers handler.

func parseIceURLs(raw string) []dto.IceServer {
	servers := make([]dto.IceServer, 0)
	if raw != "" {
		urls := strings.Split(raw, ",")
		for _, url := range urls {
			url = strings.TrimSpace(url)
			if url != "" {
				servers = append(servers, dto.IceServer{URLs: url})
			}
		}
	}
	return servers
}

func TestParseIceServerURLs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		count int
		urls  []string
	}{
		{"empty string", "", 0, nil},
		{"single STUN", "stun:stun.l.google.com:19302", 1,
			[]string{"stun:stun.l.google.com:19302"}},
		{"multiple STUN", "stun:a:1,stun:b:2", 2,
			[]string{"stun:a:1", "stun:b:2"}},
		{"with whitespace", " stun:a:1 , stun:b:2 ", 2,
			[]string{"stun:a:1", "stun:b:2"}},
		{"empty entries skipped", "stun:a:1,,stun:b:2", 2,
			[]string{"stun:a:1", "stun:b:2"}},
		{"only empties", ",, ,", 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseIceURLs(tt.input)
			if len(got) != tt.count {
				t.Errorf("got %d servers, want %d", len(got), tt.count)
				return
			}
			for i, url := range tt.urls {
				if got[i].URLs != url {
					t.Errorf("server[%d].URLs = %q, want %q", i, got[i].URLs, url)
				}
				if got[i].Username != "" {
					t.Errorf("server[%d].Username = %q, want empty", i, got[i].Username)
				}
				if got[i].Credential != "" {
					t.Errorf("server[%d].Credential = %q, want empty", i, got[i].Credential)
				}
			}
		})
	}
}

func TestParseIceServerURLsPreservesTurnCredentials(t *testing.T) {
	// Mô phỏng logic controller khi có TURN: giống parseIceURLs
	// nhưng thêm TURN server với credentials.
	input := "stun:stun.l.google.com:19302"
	servers := parseIceURLs(input)
	servers = append(servers, dto.IceServer{
		URLs:       "turn:turn.example.com:3478",
		Username:   "testuser",
		Credential: "testpass",
	})

	if len(servers) != 2 {
		t.Fatalf("got %d servers, want 2", len(servers))
	}

	turn := servers[1]
	if turn.URLs != "turn:turn.example.com:3478" {
		t.Errorf("TURN URLs = %q, want %q", turn.URLs, "turn:turn.example.com:3478")
	}
	if turn.Username != "testuser" {
		t.Errorf("TURN username = %q, want %q", turn.Username, "testuser")
	}
	if turn.Credential != "testpass" {
		t.Errorf("TURN credential = %q, want %q", turn.Credential, "testpass")
	}
}

// ─── DTO serialization tests ─────────────────────────────────────

func TestToggleVideoRequestJSON(t *testing.T) {
	data := `{"video_enabled":true}`
	var req dto.ToggleVideoRequest
	if err := unmarshal([]byte(data), &req); err != nil {
		t.Fatalf("unmarshal ToggleVideoRequest: %v", err)
	}
	if !req.VideoEnabled {
		t.Error("VideoEnabled should be true")
	}
}

func TestToggleVideoRequestFalse(t *testing.T) {
	data := `{"video_enabled":false}`
	var req dto.ToggleVideoRequest
	if err := unmarshal([]byte(data), &req); err != nil {
		t.Fatalf("unmarshal ToggleVideoRequest: %v", err)
	}
	if req.VideoEnabled {
		t.Error("VideoEnabled should be false")
	}
}

func TestIceServersResponseJSON(t *testing.T) {
	resp := dto.IceServersResponse{
		IceServers: []dto.IceServer{
			{URLs: "stun:stun.l.google.com:19302"},
			{URLs: "turn:turn.example.com:3478", Username: "user", Credential: "pass"},
		},
	}
	data, err := marshal(resp)
	if err != nil {
		t.Fatalf("marshal IceServersResponse: %v", err)
	}

	var decoded dto.IceServersResponse
	if err := unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal IceServersResponse: %v", err)
	}
	if len(decoded.IceServers) != 2 {
		t.Fatalf("got %d servers, want 2", len(decoded.IceServers))
	}
	if decoded.IceServers[1].Username != "user" {
		t.Errorf("Username = %q, want %q", decoded.IceServers[1].Username, "user")
	}
}

func TestCallStatusPayloadIncludesVideo(t *testing.T) {
	now := time.Now().UnixMilli()
	payload := dto.CallStatusPayload{
		CallID:             "call-123",
		Status:             "connected",
		CallerID:           "user-a",
		CalleeID:           "user-b",
		CallType:           "video",
		VideoEnabledCaller: true,
		VideoEnabledCallee: false,
		StartedAt:          &now,
	}

	data, err := marshal(payload)
	if err != nil {
		t.Fatalf("marshal CallStatusPayload: %v", err)
	}

	// Verify JSON contains video fields
	if !contains(string(data), `"video_enabled_caller":true`) {
		t.Error("JSON missing video_enabled_caller:true")
	}
	if !contains(string(data), `"video_enabled_callee":false`) {
		t.Error("JSON missing video_enabled_callee:false")
	}
}

// ─── Helpers ─────────────────────────────────────────────────────

func unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
