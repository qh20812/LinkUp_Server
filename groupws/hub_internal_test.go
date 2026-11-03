package groupws

// Kiểm thử nội bộ hub: hình dạng event gửi cho client, chống double-close,
// drop khi buffer đầy và các nhánh cleanup. Package-level nên chạy bằng
// `go test ./groupws/...` (CI hiện chỉ chạy ./tests/... — xem tests/groupcall
// cho phần chạy trong CI).

import (
	"encoding/json"
	"testing"
	"time"

	"linkup/dto"
)

func newTestClient(hub *Hub, userID string) *Client {
	client := &Client{
		hub:    hub,
		userID: userID,
		send:   make(chan []byte, 32),
	}
	hub.clients[userID] = map[*Client]bool{client: true}
	return client
}

func readEvents(t *testing.T, client *Client, want int) []dto.WsEvent {
	t.Helper()
	events := make([]dto.WsEvent, 0, want)
	for i := 0; i < want; i++ {
		select {
		case raw := <-client.send:
			var event dto.WsEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				t.Fatalf("event %d không phải JSON: %v", i+1, err)
			}
			events = append(events, event)
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout chờ event %d/%d", i+1, want)
		}
	}
	return events
}

func payloadMap(t *testing.T, event dto.WsEvent) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		t.Fatalf("payload không phải JSON object: %v", err)
	}
	return payload
}

// ─── trySend / closeSend ───────────────────────────────────────

func TestTrySendDropsWhenBufferIsFull(t *testing.T) {
	client := &Client{userID: "u", send: make(chan []byte, 1)}

	if !client.trySend([]byte("first")) {
		t.Fatal("frame đầu phải gửi được")
	}
	if client.trySend([]byte("second")) {
		t.Fatal("buffer đầy phải drop frame, không được đóng socket")
	}
}

func TestCloseSendIsIdempotent(t *testing.T) {
	client := &Client{userID: "u", send: make(chan []byte, 1)}

	client.closeSend()
	// Gọi lần hai không được panic (trước đây close() hai lần giết cả process).
	client.closeSend()

	if client.trySend([]byte("after-close")) {
		t.Fatal("không được gửi sau khi đã đóng")
	}
}

func TestSendEventWritesToChannel(t *testing.T) {
	client := &Client{userID: "u", send: make(chan []byte, 2)}
	client.sendEvent("group:call:joined", map[string]any{"call_id": "c1"})

	events := readEvents(t, client, 1)
	if events[0].Type != "group:call:joined" {
		t.Fatalf("sai event type: %q", events[0].Type)
	}
	if payloadMap(t, events[0])["call_id"] != "c1" {
		t.Fatal("payload bị mất call_id")
	}
}

// ─── Cleanup: rớt socket quá ân hạn ────────────────────────────

func TestCleanupSendsLeftThenCreatorChanged(t *testing.T) {
	hub := NewHub()
	session, err := hub.CreateGroupCall("chat-1", "caller", "video", []string{"b"}, []string{"caller", "b"})
	if err != nil {
		t.Fatalf("CreateGroupCall: %v", err)
	}
	if _, err := hub.RequestJoinCall("b", session.CallID); err != nil {
		t.Fatalf("RequestJoinCall: %v", err)
	}

	member := newTestClient(hub, "b")
	hub.MarkDisconnected(session.CallID, "caller")
	session.DisconnectedAt["caller"] = time.Now().UTC().Add(-time.Minute)
	hub.CleanupStaleCalls()

	events := readEvents(t, member, 2)

	if events[0].Type != "group:call:left" {
		t.Fatalf("event đầu phải là left, nhận %q", events[0].Type)
	}
	if got := payloadMap(t, events[0])["user_id"]; got != "caller" {
		t.Fatalf("left phải báo caller rời, nhận %v", got)
	}
	if events[1].Type != "group:call:creator-changed" {
		t.Fatalf("event hai phải là creator-changed, nhận %q", events[1].Type)
	}
	if got := payloadMap(t, events[1])["caller_id"]; got != "b" {
		t.Fatalf("creator mới phải là b, nhận %v", got)
	}
}

func TestCleanupEndsCallWhenNobodyRemains(t *testing.T) {
	hub := NewHub()
	session, err := hub.CreateGroupCall("chat-1", "caller", "video", nil, []string{"caller"})
	if err != nil {
		t.Fatalf("CreateGroupCall: %v", err)
	}
	caller := newTestClient(hub, "caller")

	hub.MarkDisconnected(session.CallID, "caller")
	session.DisconnectedAt["caller"] = time.Now().UTC().Add(-time.Minute)
	hub.CleanupStaleCalls()

	events := readEvents(t, caller, 1)
	if events[0].Type != "group:call:ended" {
		t.Fatalf("phải kết thúc cuộc gọi, nhận %q", events[0].Type)
	}
	if reason := payloadMap(t, events[0])["reason"]; reason != "disconnected" {
		t.Fatalf("reason phải là disconnected, nhận %v", reason)
	}
	if _, ok := hub.groupCalls[session.CallID]; ok {
		t.Fatal("session phải bị xoá khỏi hub")
	}
}

func TestCleanupDoesNotDropWithinGrace(t *testing.T) {
	hub := NewHub()
	session, err := hub.CreateGroupCall("chat-1", "caller", "video", []string{"b"}, []string{"caller", "b"})
	if err != nil {
		t.Fatalf("CreateGroupCall: %v", err)
	}
	if _, err := hub.RequestJoinCall("b", session.CallID); err != nil {
		t.Fatalf("RequestJoinCall: %v", err)
	}

	member := newTestClient(hub, "b")
	hub.MarkDisconnected(session.CallID, "b")
	hub.CleanupStaleCalls()

	select {
	case raw := <-member.send:
		t.Fatalf("trong ân hạn không được gửi event, nhận %s", raw)
	default:
	}
	if _, ok := session.Joined["b"]; !ok {
		t.Fatal("trong ân hạn participant phải còn trong Joined")
	}
}

func TestCleanupExpiresCallSendsTimeout(t *testing.T) {
	hub := NewHub()
	session, err := hub.CreateGroupCall("chat-1", "caller", "video", nil, []string{"caller"})
	if err != nil {
		t.Fatalf("CreateGroupCall: %v", err)
	}
	caller := newTestClient(hub, "caller")

	expired := time.Now().UTC().Add(-time.Minute)
	session.ExpiresAt = &expired
	hub.CleanupStaleCalls()

	events := readEvents(t, caller, 1)
	if events[0].Type != "group:call:ended" {
		t.Fatalf("phải là ended, nhận %q", events[0].Type)
	}
	if reason := payloadMap(t, events[0])["reason"]; reason != "timeout" {
		t.Fatalf("reason phải là timeout, nhận %v", reason)
	}
}

func TestCleanupIgnoresUnknownSession(t *testing.T) {
	hub := NewHub()
	hub.MarkDisconnected("không-tồn-tại", "u") // không panic
	hub.CleanupStaleCalls()

	if len(hub.groupCalls) != 0 {
		t.Fatal("không được tạo session từ call id lạ")
	}
}

// ─── Chuyển quyền creator ──────────────────────────────────────

func TestFirstJoinedSkipsDisconnectedUsers(t *testing.T) {
	session := &GroupCallSession{
		Joined: map[string]struct{}{
			"a": {},
			"b": {},
		},
		DisconnectedAt: map[string]time.Time{
			"a": time.Now().UTC(),
		},
	}

	next, ok := session.firstJoinedLocked()
	if !ok || next != "b" {
		t.Fatalf("phải chọn b (đang sống), nhận %q ok=%v", next, ok)
	}

	session.DisconnectedAt["b"] = time.Now().UTC()
	if _, ok := session.firstJoinedLocked(); ok {
		t.Fatal("không còn ai sống thì không được chọn creator")
	}
}

func TestMarkDisconnectedOnlyForJoinedUsers(t *testing.T) {
	hub := NewHub()
	session, err := hub.CreateGroupCall("chat-1", "caller", "video", []string{"b"}, []string{"caller", "b"})
	if err != nil {
		t.Fatalf("CreateGroupCall: %v", err)
	}

	if hub.MarkDisconnected(session.CallID, "b") {
		t.Fatal("người chưa join không được đánh dấu disconnected")
	}
	if !hub.MarkDisconnected(session.CallID, "caller") {
		t.Fatal("caller đang joined phải đánh dấu được")
	}
	if _, ok := session.DisconnectedAt["caller"]; !ok {
		t.Fatal("thiếu dấu vết disconnect")
	}
}

// ─── Client tracking ───────────────────────────────────────────

func TestHasOtherClients(t *testing.T) {
	hub := NewHub()
	first := newTestClient(hub, "u")

	if hub.HasOtherClients("u", first) {
		t.Fatal("chỉ có 1 socket thì không được coi là còn tab khác")
	}

	second := &Client{hub: hub, userID: "u", send: make(chan []byte, 1)}
	hub.clients["u"][second] = true

	if !hub.HasOtherClients("u", first) {
		t.Fatal("có 2 socket phải trả true")
	}
	if !hub.HasOtherClients("u", second) {
		t.Fatal("socket còn lại cũng phải thấy socket kia")
	}
	if hub.HasOtherClients("khác", first) {
		t.Fatal("user không tồn tại phải trả false")
	}
}

func TestRelaySignalMarksSenderActive(t *testing.T) {
	hub := NewHub()
	session, err := hub.CreateGroupCall("chat-1", "caller", "video", nil, []string{"caller"})
	if err != nil {
		t.Fatalf("CreateGroupCall: %v", err)
	}

	if _, _, err := hub.RelaySignalTargets(session.CallID, "caller", ""); err != nil {
		t.Fatalf("RelaySignalTargets: %v", err)
	}
	if _, ok := session.ActiveParticipants["caller"]; !ok {
		t.Fatal("người gửi signal phải được đánh dấu active")
	}
}

func TestListCallIDsByUserAcrossChats(t *testing.T) {
	hub := NewHub()
	first, err := hub.CreateGroupCall("chat-1", "caller", "video", nil, []string{"caller"})
	if err != nil {
		t.Fatalf("CreateGroupCall chat-1: %v", err)
	}
	second, err := hub.CreateGroupCall("chat-2", "caller", "video", nil, []string{"caller"})
	if err != nil {
		t.Fatalf("CreateGroupCall chat-2: %v", err)
	}

	ids := hub.ListCallIDsByUser("caller")
	if len(ids) != 2 {
		t.Fatalf("mong đợi 2 call, nhận %v", ids)
	}
	if ids[0] > ids[1] {
		t.Fatal("danh sách call phải được sắp xếp")
	}
	if !containsString(ids, first.CallID) || !containsString(ids, second.CallID) {
		t.Fatalf("thiếu call trong kết quả: %v", ids)
	}
}

func containsString(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}
