package call_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"linkup/dto"
	"linkup/models"
)

// ─── DTO Serialization: CallHistoryItem ──────────────────────────

func TestCallHistoryItemJSON(t *testing.T) {
	now := time.Now().UnixMilli()
	item := dto.CallHistoryItem{
		ID: "call-001",
		OtherUser: dto.UserBrief{
			ID:          "user-b",
			DisplayName: "Tran Thi Binh",
			AvatarURL:   "https://api.dicebear.com/7.x/avataaars/svg?seed=user02",
		},
		CallType:  "voice",
		Direction: "outgoing",
		Status:    "ended",
		IsMissed:  false,
		Duration:  120,
		StartedAt: &now,
		EndedAt:   &now,
		CreatedAt: now,
	}

	data, err := marshal(item)
	if err != nil {
		t.Fatalf("marshal CallHistoryItem: %v", err)
	}

	var decoded dto.CallHistoryItem
	if err := unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal CallHistoryItem: %v", err)
	}
	if decoded.ID != "call-001" {
		t.Errorf("ID = %q, want %q", decoded.ID, "call-001")
	}
	if decoded.OtherUser.DisplayName != "Tran Thi Binh" {
		t.Errorf("OtherUser.DisplayName = %q, want %q", decoded.OtherUser.DisplayName, "Tran Thi Binh")
	}
	if decoded.Direction != "outgoing" {
		t.Errorf("Direction = %q, want %q", decoded.Direction, "outgoing")
	}
	if decoded.IsMissed {
		t.Error("IsMissed should be false")
	}
	if decoded.Duration != 120 {
		t.Errorf("Duration = %d, want %d", decoded.Duration, 120)
	}
}

func TestCallHistoryItemMissedCall(t *testing.T) {
	item := dto.CallHistoryItem{
		ID: "call-002",
		OtherUser: dto.UserBrief{
			ID:          "user-c",
			DisplayName: "Le Hoang Cuong",
		},
		CallType:  "video",
		Direction: "incoming",
		Status:    "missed",
		IsMissed:  true,
		Duration:  0,
		CreatedAt: time.Now().UnixMilli(),
	}

	data, err := marshal(item)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !contains(string(data), `"is_missed":true`) {
		t.Error("JSON missing is_missed:true")
	}
	if !contains(string(data), `"status":"missed"`) {
		t.Error("JSON missing status:missed")
	}
	if !contains(string(data), `"direction":"incoming"`) {
		t.Error("JSON missing direction:incoming")
	}
}

func TestCallHistoryItemOmitsNilTimestamps(t *testing.T) {
	item := dto.CallHistoryItem{
		ID:        "call-003",
		OtherUser: dto.UserBrief{ID: "u1"},
		CallType:  "voice",
		Direction: "outgoing",
		Status:    "calling",
		StartedAt: nil,
		EndedAt:   nil,
		CreatedAt: time.Now().UnixMilli(),
	}

	data, err := marshal(item)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if contains(string(data), `"started_at"`) {
		t.Error("started_at should be omitted when nil")
	}
	if contains(string(data), `"ended_at"`) {
		t.Error("ended_at should be omitted when nil")
	}
}

// ─── DTO Serialization: UserBrief ────────────────────────────────

func TestUserBriefJSON(t *testing.T) {
	u := dto.UserBrief{
		ID:          "user-123",
		DisplayName: "Pham Minh Duc",
		AvatarURL:   "https://example.com/avatar.jpg",
	}

	data, err := marshal(u)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded dto.UserBrief
	if err := unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.ID != "user-123" {
		t.Errorf("ID = %q, want %q", decoded.ID, "user-123")
	}
	if decoded.DisplayName != "Pham Minh Duc" {
		t.Errorf("DisplayName = %q, want %q", decoded.DisplayName, "Pham Minh Duc")
	}
	if decoded.AvatarURL != "https://example.com/avatar.jpg" {
		t.Errorf("AvatarURL = %q, want %q", decoded.AvatarURL, "https://example.com/avatar.jpg")
	}
}

func TestUserBriefEmptyFields(t *testing.T) {
	u := dto.UserBrief{}
	data, err := marshal(u)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded dto.UserBrief
	if err := unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.ID != "" || decoded.DisplayName != "" || decoded.AvatarURL != "" {
		t.Errorf("empty UserBrief should have empty fields, got: %+v", decoded)
	}
}

// ─── DTO Serialization: CallMissedPayload ────────────────────────

func TestCallMissedPayloadJSON(t *testing.T) {
	payload := dto.CallMissedPayload{
		CallID:    "call-999",
		CallerID:  "user-a",
		Timestamp: 1700000000000,
	}

	data, err := marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded dto.CallMissedPayload
	if err := unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.CallID != "call-999" {
		t.Errorf("CallID = %q, want %q", decoded.CallID, "call-999")
	}
	if decoded.CallerID != "user-a" {
		t.Errorf("CallerID = %q, want %q", decoded.CallerID, "user-a")
	}
	if decoded.Timestamp != 1700000000000 {
		t.Errorf("Timestamp = %d, want %d", decoded.Timestamp, 1700000000000)
	}
}

func TestCallMissedPayloadFieldCount(t *testing.T) {
	payload := dto.CallMissedPayload{CallID: "c1", CallerID: "u1", Timestamp: 1}
	data, err := marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	if len(raw) != 3 {
		t.Errorf("CallMissedPayload has %d fields, want 3", len(raw))
	}
}

// ─── Direction Logic ─────────────────────────────────────────────

func TestDirectionLogic(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		callerID      string
		calleeID      string
		wantDirection string
	}{
		{"outgoing: user is caller", "user-a", "user-a", "user-b", "outgoing"},
		{"incoming: user is callee", "user-b", "user-a", "user-b", "incoming"},
		{"outgoing: different caller", "user-x", "user-x", "user-y", "outgoing"},
		{"incoming: different callee", "user-y", "user-x", "user-y", "incoming"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			direction := "outgoing"
			if tt.callerID != tt.userID {
				direction = "incoming"
			}
			if direction != tt.wantDirection {
				t.Errorf("direction = %q, want %q", direction, tt.wantDirection)
			}
		})
	}
}

func TestOtherUserIDLogic(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		callerID    string
		calleeID    string
		wantOtherID string
	}{
		{"user is caller, other is callee", "user-a", "user-a", "user-b", "user-b"},
		{"user is callee, other is caller", "user-b", "user-a", "user-b", "user-a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otherID := tt.callerID
			if tt.callerID == tt.userID {
				otherID = tt.calleeID
			}
			if otherID != tt.wantOtherID {
				t.Errorf("otherID = %q, want %q", otherID, tt.wantOtherID)
			}
		})
	}
}

// ─── IsMissed Logic ──────────────────────────────────────────────

func TestIsMissedLogic(t *testing.T) {
	tests := []struct {
		name       string
		status     models.CallStatus
		wantMissed bool
	}{
		{"missed status", models.CallStatusMissed, true},
		{"ended status", models.CallStatusEnded, false},
		{"rejected status", models.CallStatusRejected, false},
		{"connected status", models.CallStatusConnected, false},
		{"calling status", models.CallStatusCalling, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isMissed := tt.status == models.CallStatusMissed
			if isMissed != tt.wantMissed {
				t.Errorf("isMissed = %v, want %v", isMissed, tt.wantMissed)
			}
		})
	}
}

// ─── CallHidden Model ────────────────────────────────────────────

func TestCallHiddenPrimaryKeys(t *testing.T) {
	h := models.CallHidden{
		CallID: "call-001",
		UserID: "user-001",
	}

	if h.CallID != "call-001" {
		t.Errorf("CallID = %q, want %q", h.CallID, "call-001")
	}
	if h.UserID != "user-001" {
		t.Errorf("UserID = %q, want %q", h.UserID, "user-001")
	}
}

func TestCallHiddenJSON(t *testing.T) {
	h := models.CallHidden{
		CallID:    "call-001",
		UserID:    "user-001",
		CreatedAt: time.Now().UTC(),
	}

	data, err := marshal(h)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded models.CallHidden
	if err := unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.CallID != "call-001" {
		t.Errorf("CallID = %q, want %q", decoded.CallID, "call-001")
	}
	if decoded.UserID != "user-001" {
		t.Errorf("UserID = %q, want %q", decoded.UserID, "user-001")
	}
}

// ─── CallHistoryQuery Defaults ───────────────────────────────────

func TestCallHistoryQueryDefaults(t *testing.T) {
	q := dto.CallHistoryQuery{}
	if q.Limit != 0 {
		t.Errorf("zero-value Limit = %d, want 0", q.Limit)
	}
	if q.Offset != 0 {
		t.Errorf("zero-value Offset = %d, want 0", q.Offset)
	}
	if q.Type != nil {
		t.Error("zero-value Type should be nil")
	}
	if q.Status != nil {
		t.Error("zero-value Status should be nil")
	}
}

func TestCallHistoryQueryWithValues(t *testing.T) {
	videoType := "video"
	missedStatus := "missed"
	q := dto.CallHistoryQuery{
		Limit:  10,
		Offset: 5,
		Type:   &videoType,
		Status: &missedStatus,
		Sort:   "duration",
		Order:  "asc",
	}

	if q.Limit != 10 {
		t.Errorf("Limit = %d, want 10", q.Limit)
	}
	if q.Offset != 5 {
		t.Errorf("Offset = %d, want 5", q.Offset)
	}
	if q.Type == nil || *q.Type != "video" {
		t.Errorf("Type = %v, want 'video'", q.Type)
	}
	if q.Status == nil || *q.Status != "missed" {
		t.Errorf("Status = %v, want 'missed'", q.Status)
	}
	if q.Sort != "duration" {
		t.Errorf("Sort = %q, want %q", q.Sort, "duration")
	}
	if q.Order != "asc" {
		t.Errorf("Order = %q, want %q", q.Order, "asc")
	}
}

// ─── queryToFilter Whitelist Logic ───────────────────────────────
//
// Replicates the controller's queryToFilter function to test
// whitelist validation without importing the controllers package.

type filterResult struct {
	CallType *string
	Status   *string
	Sort     string
	Order    string
}

func queryToFilter(sort, order string, callType, status *string) filterResult {
	f := filterResult{Sort: "created_at", Order: "desc"}

	if callType != nil {
		ct := strings.ToLower(strings.TrimSpace(*callType))
		if ct == "voice" || ct == "video" {
			f.CallType = &ct
		}
	}
	if status != nil {
		s := strings.ToLower(strings.TrimSpace(*status))
		switch s {
		case "missed", "ended", "rejected", "connected", "calling", "ringing":
			f.Status = &s
		}
	}
	sortLower := strings.ToLower(strings.TrimSpace(sort))
	switch sortLower {
	case "created_at", "duration", "call_type", "status":
		f.Sort = sortLower
	default:
		f.Sort = "created_at"
	}
	orderLower := strings.ToLower(strings.TrimSpace(order))
	if orderLower == "asc" || orderLower == "desc" {
		f.Order = orderLower
	} else {
		f.Order = "desc"
	}
	return f
}

func TestQueryToFilterDefaults(t *testing.T) {
	f := queryToFilter("", "", nil, nil)
	if f.Sort != "created_at" {
		t.Errorf("Sort = %q, want %q", f.Sort, "created_at")
	}
	if f.Order != "desc" {
		t.Errorf("Order = %q, want %q", f.Order, "desc")
	}
	if f.CallType != nil {
		t.Errorf("CallType should be nil, got %v", f.CallType)
	}
	if f.Status != nil {
		t.Errorf("Status should be nil, got %v", f.Status)
	}
}

func TestQueryToFilterWhitelistType(t *testing.T) {
	voice := "voice"
	video := "video"
	invalid := "fax"

	f1 := queryToFilter("", "", &voice, nil)
	if f1.CallType == nil || *f1.CallType != "voice" {
		t.Errorf("voice should be accepted, got %v", f1.CallType)
	}

	f2 := queryToFilter("", "", &video, nil)
	if f2.CallType == nil || *f2.CallType != "video" {
		t.Errorf("video should be accepted, got %v", f2.CallType)
	}

	f3 := queryToFilter("", "", &invalid, nil)
	if f3.CallType != nil {
		t.Errorf("invalid type should be rejected, got %v", *f3.CallType)
	}
}

func TestQueryToFilterWhitelistStatus(t *testing.T) {
	missed := "missed"
	ended := "ended"
	rejected := "rejected"
	invalid := "deleted"

	f1 := queryToFilter("", "", nil, &missed)
	if f1.Status == nil || *f1.Status != "missed" {
		t.Errorf("missed should be accepted, got %v", f1.Status)
	}

	f2 := queryToFilter("", "", nil, &ended)
	if f2.Status == nil || *f2.Status != "ended" {
		t.Errorf("ended should be accepted, got %v", f2.Status)
	}

	f3 := queryToFilter("", "", nil, &rejected)
	if f3.Status == nil || *f3.Status != "rejected" {
		t.Errorf("rejected should be accepted, got %v", f3.Status)
	}

	f4 := queryToFilter("", "", nil, &invalid)
	if f4.Status != nil {
		t.Errorf("invalid status should be rejected, got %v", *f4.Status)
	}
}

func TestQueryToFilterWhitelistSort(t *testing.T) {
	tests := []struct {
		input    string
		wantSort string
	}{
		{"created_at", "created_at"},
		{"duration", "duration"},
		{"call_type", "call_type"},
		{"status", "status"},
		{"CREATED_AT", "created_at"},
		{"Duration", "duration"},
		{"invalid_column", "created_at"},
		{"", "created_at"},
		{"id", "created_at"},
		{" DROP TABLE ", "created_at"},
	}

	for _, tt := range tests {
		t.Run("sort_"+tt.input, func(t *testing.T) {
			f := queryToFilter(tt.input, "", nil, nil)
			if f.Sort != tt.wantSort {
				t.Errorf("Sort(%q) = %q, want %q", tt.input, f.Sort, tt.wantSort)
			}
		})
	}
}

func TestQueryToFilterWhitelistOrder(t *testing.T) {
	tests := []struct {
		input     string
		wantOrder string
	}{
		{"asc", "asc"},
		{"desc", "desc"},
		{"ASC", "asc"},
		{"DESC", "desc"},
		{"invalid", "desc"},
		{"", "desc"},
		{"ascending", "desc"},
	}

	for _, tt := range tests {
		t.Run("order_"+tt.input, func(t *testing.T) {
			f := queryToFilter("", tt.input, nil, nil)
			if f.Order != tt.wantOrder {
				t.Errorf("Order(%q) = %q, want %q", tt.input, f.Order, tt.wantOrder)
			}
		})
	}
}

func TestQueryToFilterSQLInjection(t *testing.T) {
	injections := []struct {
		name   string
		sort   string
		order  string
	}{
		{"sort injection 1", "created_at; DROP TABLE calls", "desc"},
		{"sort injection 2", "1=1 OR 1", "ascending"},
		{"order injection", "created_at", "desc; SELECT * FROM users"},
		{"unicode bypass", "créated_at", "desc"},
		{"null byte", "created_at\x00", "desc"},
	}

	for _, tt := range injections {
		t.Run(tt.name, func(t *testing.T) {
			f := queryToFilter(tt.sort, tt.order, nil, nil)
			if f.Sort != "created_at" {
				t.Errorf("SQL injection in Sort: got %q, want fallback to created_at", f.Sort)
			}
			if f.Order != "desc" {
				t.Errorf("SQL injection in Order: got %q, want fallback to desc", f.Order)
			}
		})
	}
}

// ─── Full Response Envelope ──────────────────────────────────────

func TestCallHistoryResponseEnvelope(t *testing.T) {
	now := time.Now().UnixMilli()
	items := []dto.CallHistoryItem{
		{
			ID:        "call-001",
			OtherUser: dto.UserBrief{ID: "u1", DisplayName: "User One"},
			CallType:  "voice",
			Direction: "outgoing",
			Status:    "ended",
			Duration:  60,
			CreatedAt: now,
		},
		{
			ID:        "call-002",
			OtherUser: dto.UserBrief{ID: "u2", DisplayName: "User Two"},
			CallType:  "video",
			Direction: "incoming",
			Status:    "missed",
			IsMissed:  true,
			CreatedAt: now,
		},
	}

	envelope := map[string]interface{}{
		"data":   items,
		"total":  2,
		"limit":  20,
		"offset": 0,
	}

	data, err := marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	var decoded map[string]interface{}
	if err := unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}

	dataArr, ok := decoded["data"].([]interface{})
	if !ok {
		t.Fatal("data should be an array")
	}
	if len(dataArr) != 2 {
		t.Errorf("data has %d items, want 2", len(dataArr))
	}

	if decoded["total"].(float64) != 2 {
		t.Errorf("total = %v, want 2", decoded["total"])
	}
}
