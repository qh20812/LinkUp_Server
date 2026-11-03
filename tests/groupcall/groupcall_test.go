package groupcall_test

// Kiểm thử hành vi public của session cuộc gọi nhóm: chọn người tham gia,
// join/pending, scope signal, quyền mic/video, kết thúc và grace khi rớt socket.
// Không cần MySQL/MongoDB.

import (
	"testing"
	"time"

	"linkup/dto"
	"linkup/groupws"
)

func newCall(t *testing.T, hub *groupws.Hub, callerID string, participantIDs, memberIDs []string) *groupws.GroupCallSession {
	t.Helper()
	session, err := hub.CreateGroupCall("chat-1", callerID, "video", participantIDs, memberIDs)
	if err != nil {
		t.Fatalf("CreateGroupCall: %v", err)
	}
	if session == nil {
		t.Fatal("CreateGroupCall trả về session nil")
	}
	return session
}

func mustJoin(t *testing.T, hub *groupws.Hub, userID, callID string) {
	t.Helper()
	result, err := hub.RequestJoinCall(userID, callID)
	if err != nil {
		t.Fatalf("RequestJoinCall(%s): %v", userID, err)
	}
	if !result.JoinedNow {
		t.Fatalf("RequestJoinCall(%s) không join ngay", userID)
	}
}

// ─── Tạo cuộc gọi ────────────────────────────────────────────────

func TestCreateGroupCallWithoutSelectionInvitesOnlyCaller(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", nil, []string{"caller", "b", "c"})

	if len(session.Participants) != 1 {
		t.Fatalf("mong đợi chỉ caller, nhận %d người: %v", len(session.Participants), session.ParticipantIDs())
	}
	if _, ok := session.Participants["caller"]; !ok {
		t.Fatal("caller phải nằm trong Participants")
	}
	if session.CallerID != "caller" || session.ChatID != "chat-1" {
		t.Fatalf("session sai thông tin: caller=%q chat=%q", session.CallerID, session.ChatID)
	}
	if session.Muted["caller"] || !session.VideoEnabled["caller"] {
		t.Fatal("trạng thái mic/video mặc định sai")
	}
}

func TestCreateGroupCallRejectsNonMember(t *testing.T) {
	hub := groupws.NewHub()
	if _, err := hub.CreateGroupCall("chat-1", "caller", "video", []string{"stranger"}, []string{"caller"}); err == nil {
		t.Fatal("mong đợi lỗi khi mời người ngoài nhóm")
	}
}

func TestCreateGroupCallRejectsSecondCallInSameChat(t *testing.T) {
	hub := groupws.NewHub()
	newCall(t, hub, "caller", nil, []string{"caller"})

	if _, err := hub.CreateGroupCall("chat-1", "caller", "video", nil, []string{"caller"}); err == nil {
		t.Fatal("mong đợi lỗi khi nhóm đã có cuộc gọi")
	}
}

func TestCreateGroupCallRejectsNonVideo(t *testing.T) {
	hub := groupws.NewHub()
	if _, err := hub.CreateGroupCall("chat-1", "caller", "voice", nil, []string{"caller"}); err == nil {
		t.Fatal("mong đợi lỗi với call_type voice")
	}
}

// ─── Join / pending ──────────────────────────────────────────────

func TestRequestJoinCallAutoJoinsInvitee(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b"}, []string{"caller", "b"})

	result, err := hub.RequestJoinCall("b", session.CallID)
	if err != nil {
		t.Fatalf("RequestJoinCall: %v", err)
	}
	if !result.JoinedNow {
		t.Fatal("người được mời phải join ngay")
	}
	if len(result.JoinedIDs) != 2 {
		t.Fatalf("mong đợi 2 người joined, nhận %v", result.JoinedIDs)
	}

	snapshot, err := hub.GetGroupCallSnapshot(session.CallID)
	if err != nil {
		t.Fatalf("GetGroupCallSnapshot: %v", err)
	}
	if session.ExpiresAt != nil {
		t.Fatal("ExpiresAt phải được xoá khi đã có 2 người join")
	}
	if len(snapshot.Joined) != 2 {
		t.Fatalf("snapshot joined sai: %v", snapshot.Joined)
	}
}

func TestRequestJoinCallFromStrangerNeedsApproval(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b"}, []string{"caller", "b", "stranger"})

	result, err := hub.RequestJoinCall("stranger", session.CallID)
	if err != nil {
		t.Fatalf("RequestJoinCall: %v", err)
	}
	if result.JoinedNow {
		t.Fatal("người không được mời không được tự join")
	}

	approved, err := hub.ApproveJoinCall("caller", "stranger", session.CallID)
	if err != nil {
		t.Fatalf("ApproveJoinCall: %v", err)
	}
	if !contains(approved.ParticipantIDs, "stranger") {
		t.Fatalf("người được duyệt phải nằm trong Participants: %v", approved.ParticipantIDs)
	}
	if approved.CallerID != "caller" {
		t.Fatalf("ApproveJoinResult phải trả caller hiện tại, nhận %q", approved.CallerID)
	}
}

func TestApproveJoinCallRejectsNonCreator(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b"}, []string{"caller", "b"})
	_, _ = hub.RequestJoinCall("stranger", session.CallID)

	if _, err := hub.ApproveJoinCall("b", "stranger", session.CallID); err == nil {
		t.Fatal("chỉ creator mới được duyệt")
	}
}

func TestGetParticipantsSnapshotRequiresParticipant(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", nil, []string{"caller"})

	if _, err := hub.GetParticipantsSnapshot(session.CallID, "stranger"); err == nil {
		t.Fatal("người ngoài không được xem participants")
	}
	if _, err := hub.GetParticipantsSnapshot(session.CallID, "caller"); err != nil {
		t.Fatalf("creator phải xem được: %v", err)
	}
}

// ─── Signal targeting ───────────────────────────────────────────

func TestRelaySignalTargetsScopesToOnePeer(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b", "c"}, []string{"caller", "b", "c"})
	mustJoin(t, hub, "b", session.CallID)
	mustJoin(t, hub, "c", session.CallID)

	chatID, recipients, err := hub.RelaySignalTargets(session.CallID, "caller", "c")
	if err != nil {
		t.Fatalf("RelaySignalTargets: %v", err)
	}
	if chatID != "chat-1" {
		t.Fatalf("chatID sai: %q", chatID)
	}
	if len(recipients) != 1 || recipients[0] != "c" {
		t.Fatalf("signal phải chỉ tới c, nhận %v", recipients)
	}
}

func TestRelaySignalTargetsRejectsInvalidTarget(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b"}, []string{"caller", "b"})

	if _, _, err := hub.RelaySignalTargets(session.CallID, "caller", "b"); err == nil {
		t.Fatal("không được gửi signal tới người chưa join")
	}
	if _, _, err := hub.RelaySignalTargets(session.CallID, "caller", "caller"); err == nil {
		t.Fatal("không được gửi signal cho chính mình")
	}
}

func TestRelaySignalTargetsWithoutTargetBroadcasts(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b", "c"}, []string{"caller", "b", "c"})
	mustJoin(t, hub, "b", session.CallID)
	mustJoin(t, hub, "c", session.CallID)

	_, recipients, err := hub.RelaySignalTargets(session.CallID, "caller", "")
	if err != nil {
		t.Fatalf("RelaySignalTargets: %v", err)
	}
	if len(recipients) != 2 || contains(recipients, "caller") {
		t.Fatalf("broadcast phải gửi cho mọi người trừ sender, nhận %v", recipients)
	}
}

func TestRelaySignalTargetsRejectsNonJoinedSender(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b"}, []string{"caller", "b"})

	if _, _, err := hub.RelaySignalTargets(session.CallID, "b", "caller"); err == nil {
		t.Fatal("người chưa join không được gửi signal")
	}
}

// ─── Mic / video ────────────────────────────────────────────────

func TestSetParticipantMutedRequiresCreator(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b"}, []string{"caller", "b"})
	mustJoin(t, hub, "b", session.CallID)

	if _, _, err := hub.SetParticipantMuted(session.CallID, "b", "caller", true, true); err == nil {
		t.Fatal("member không được tắt mic người khác")
	}

	changed, notifyIDs, err := hub.SetParticipantMuted(session.CallID, "caller", "b", true, true)
	if err != nil {
		t.Fatalf("creator mute: %v", err)
	}
	if !changed || len(notifyIDs) != 2 {
		t.Fatalf("mong đợi thay đổi và notify 2 người, nhận changed=%v notify=%v", changed, notifyIDs)
	}

	// Lặp lại cùng trạng thái → không thay đổi, không notify.
	if changed, _, _ := hub.SetParticipantMuted(session.CallID, "caller", "b", true, true); changed {
		t.Fatal("set lại cùng trạng thái không được coi là thay đổi")
	}

	// Creator tự mute phải dùng toggle-mic (requireCreator=false).
	if _, _, err := hub.SetParticipantMuted(session.CallID, "caller", "", true, false); err != nil {
		t.Fatalf("toggle mic của creator: %v", err)
	}
}

func TestSetParticipantVideoNotifiesJoined(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b"}, []string{"caller", "b"})
	mustJoin(t, hub, "b", session.CallID)

	changed, notifyIDs, err := hub.SetParticipantVideo(session.CallID, "b", false)
	if err != nil {
		t.Fatalf("SetParticipantVideo: %v", err)
	}
	if !changed || len(notifyIDs) != 2 {
		t.Fatalf("mong đợi changed + 2 notify, nhận changed=%v notify=%v", changed, notifyIDs)
	}
	if session.VideoEnabled["b"] {
		t.Fatal("video phải đang tắt")
	}
}

// ─── Kết thúc ───────────────────────────────────────────────────

func TestEndCallByCreatorEndsWholeCall(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b"}, []string{"caller", "b"})
	mustJoin(t, hub, "b", session.CallID)

	result, err := hub.EndCallByUser("caller", session.CallID)
	if err != nil {
		t.Fatalf("EndCallByUser: %v", err)
	}
	if !result.Ended {
		t.Fatal("creator kết thúc phải kết thúc cả cuộc gọi")
	}
	if _, err := hub.GetGroupCallSnapshot(session.CallID); err == nil {
		t.Fatal("session phải bị xoá")
	}
}

func TestEndCallByMemberOnlyLeaves(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b"}, []string{"caller", "b"})
	mustJoin(t, hub, "b", session.CallID)

	result, err := hub.EndCallByUser("b", session.CallID)
	if err != nil {
		t.Fatalf("EndCallByUser: %v", err)
	}
	if result.Ended {
		t.Fatal("member rời không được kết thúc cuộc gọi")
	}
	snapshot, err := hub.GetGroupCallSnapshot(session.CallID)
	if err != nil {
		t.Fatalf("session phải còn: %v", err)
	}
	if contains(snapshot.Joined, "b") {
		t.Fatal("member đã rời không còn trong Joined")
	}
}

func TestEndCallRejectsNonParticipant(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b"}, []string{"caller", "b"})

	if _, err := hub.EndCallByUser("b", session.CallID); err == nil {
		t.Fatal("người chưa join không được kết thúc cuộc gọi")
	}
}

// ─── Grace khi rớt socket ──────────────────────────────────────

func TestDisconnectGraceRemovesParticipant(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b"}, []string{"caller", "b"})
	mustJoin(t, hub, "b", session.CallID)

	if !hub.MarkDisconnected(session.CallID, "b") {
		t.Fatal("MarkDisconnected phải trả true cho người đang joined")
	}

	// Chưa hết ân hạn: vẫn còn trong cuộc gọi.
	hub.CleanupStaleCalls()
	snapshot, err := hub.GetGroupCallSnapshot(session.CallID)
	if err != nil {
		t.Fatalf("session phải còn trong ân hạn: %v", err)
	}
	if !contains(snapshot.Joined, "b") {
		t.Fatal("trong ân hạn người dùng phải còn trong Joined để rejoin")
	}

	// Ép mốc ngắt về quá khứ (session là con trỏ sống do CreateGroupCall trả về).
	session.DisconnectedAt["b"] = time.Now().UTC().Add(-time.Minute)
	hub.CleanupStaleCalls()

	snapshot, err = hub.GetGroupCallSnapshot(session.CallID)
	if err != nil {
		t.Fatalf("session phải còn vì creator vẫn ở trong: %v", err)
	}
	if contains(snapshot.Joined, "b") {
		t.Fatal("hết ân hạn phải rút khỏi Joined")
	}
}

func TestDisconnectTransfersCreatorRole(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b"}, []string{"caller", "b"})
	mustJoin(t, hub, "b", session.CallID)

	hub.MarkDisconnected(session.CallID, "caller")
	session.DisconnectedAt["caller"] = time.Now().UTC().Add(-time.Minute)
	hub.CleanupStaleCalls()

	snapshot, err := hub.GetGroupCallSnapshot(session.CallID)
	if err != nil {
		t.Fatalf("session phải còn: %v", err)
	}
	if snapshot.CallerID != "b" {
		t.Fatalf("quyền creator phải chuyển cho b, nhận %q", snapshot.CallerID)
	}
}

func TestDisconnectEndsCallWhenLastParticipantLeaves(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", nil, []string{"caller"})

	hub.MarkDisconnected(session.CallID, "caller")
	session.DisconnectedAt["caller"] = time.Now().UTC().Add(-time.Minute)
	hub.CleanupStaleCalls()

	if _, err := hub.GetGroupCallSnapshot(session.CallID); err == nil {
		t.Fatal("session phải bị xoá khi người cuối cùng rớt")
	}
	if ids := hub.ListCallIDsByUser("caller"); len(ids) != 0 {
		t.Fatalf("không còn call nào của caller, nhận %v", ids)
	}
}

func TestRejoinAfterGraceRestoresMembership(t *testing.T) {
	hub := groupws.NewHub()
	session := newCall(t, hub, "caller", []string{"b"}, []string{"caller", "b"})
	mustJoin(t, hub, "b", session.CallID)

	hub.MarkDisconnected(session.CallID, "b")
	session.DisconnectedAt["b"] = time.Now().UTC().Add(-time.Minute)
	hub.CleanupStaleCalls()

	// Người được mời vẫn nằm trong Participants nên rejoin không cần duyệt lại.
	result, err := hub.RequestJoinCall("b", session.CallID)
	if err != nil {
		t.Fatalf("RequestJoinCall sau ân hạn: %v", err)
	}
	if !result.JoinedNow {
		t.Fatal("người được mời phải rejoin được ngay")
	}
}

// ─── DTO ────────────────────────────────────────────────────────

func TestGroupCallSignalPayloadValidate(t *testing.T) {
	small := &dto.GroupCallSignalPayload{Signal: []byte(`{"type":"ice"}`)}
	if err := small.Validate(); err != nil {
		t.Fatalf("signal nhỏ phải hợp lệ: %v", err)
	}

	huge := &dto.GroupCallSignalPayload{Signal: make([]byte, 40*1024)}
	if err := huge.Validate(); err == nil {
		t.Fatal("signal > 32KB phải bị từ chối")
	}
}

// ─── helpers ────────────────────────────────────────────────────

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}
