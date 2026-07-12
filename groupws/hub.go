package groupws

import (
	"encoding/json"
	"errors"
	"fmt"
	"linkup/dto"
	"linkup/utils"
	"sort"
	"strings"
	"sync"
	"time"
)

type BroadcastMessage struct {
	ChatID string
	Data   []byte
}

type Hub struct {
	rooms        map[string]map[*Client]bool
	clients      map[string]map[*Client]bool
	register     chan *Client
	unregister   chan *Client
	broadcast    chan *BroadcastMessage
	groupCalls   map[string]*GroupCallSession
	groupChatHub *Hub
	mu           sync.RWMutex
}

type GroupCallSession struct {
	CallID             string
	ChatID             string
	CallerID           string
	CallType           string
	Participants       map[string]struct{}
	PendingRequests    map[string]time.Time // thời điểm yêu cầu được tạo để cleanup timeout
	Joined             map[string]struct{}
	ActiveParticipants map[string]struct{} // người đã connect media thành công
	Muted              map[string]bool
	VideoEnabled       map[string]bool
	CreatedAt          time.Time
	LastActivity       time.Time
	ExpiresAt          *time.Time // hết hạn nếu không có ai join sau một khoảng ngắn
	Status             string
}

func (s *GroupCallSession) ParticipantIDs() []string {
	ids := make([]string, 0, len(s.Participants))
	for id := range s.Participants {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (s *GroupCallSession) ActiveParticipantIDs() []string {
	ids := make([]string, 0, len(s.ActiveParticipants))
	for id := range s.ActiveParticipants {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
		groupCalls: make(map[string]*GroupCallSession),
	}
}

func (h *Hub) SetGroupChatHub(groupChatHub *Hub) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.groupChatHub = groupChatHub
}

func (h *Hub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			for chatID, clients := range h.rooms {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.rooms, chatID)
				}
			}
			if clients, ok := h.clients[client.userID]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.clients, client.userID)
				}
			}
			h.mu.Unlock()
			close(client.send)

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.rooms[message.ChatID]
			for client := range clients {
				select {
				case client.send <- message.Data:
				default:
					h.mu.RUnlock()
					h.mu.Lock()
					delete(clients, client)
					close(client.send)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			h.CleanupStaleCalls()
		}
	}
}

func (h *Hub) JoinChat(chatID string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[chatID] == nil {
		h.rooms[chatID] = make(map[*Client]bool)
	}
	h.rooms[chatID][client] = true
}

func (h *Hub) LeaveChat(chatID string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[chatID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, chatID)
		}
	}
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(chatID string, msg any) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.broadcast <- &BroadcastMessage{
		ChatID: chatID,
		Data:   data,
	}
}

func (h *Hub) CreateGroupCall(chatID, callerID, callType string, participantIDs, memberIDs []string) (*GroupCallSession, error) {
	callType = strings.ToLower(strings.TrimSpace(callType))
	if callType == "" {
		// Mặc định vẫn là video để giữ tương thích với client cũ, nhưng tuyệt đối không cho audio-only.
		callType = "video"
	}
	if callType != "video" {
		return nil, errors.New("group call hiện chỉ hỗ trợ video")
	}

	selected := make(map[string]struct{})

	memberSet := make(map[string]struct{}, len(memberIDs))
	for _, id := range memberIDs {
		if id == "" {
			continue
		}
		memberSet[id] = struct{}{}
	}

	if len(participantIDs) == 0 {
		for id := range memberSet {
			selected[id] = struct{}{}
		}
	} else {
		for _, id := range participantIDs {
			if id == "" {
				continue
			}
			if _, ok := memberSet[id]; !ok {
				return nil, fmt.Errorf("thành viên %s không thuộc nhóm", id)
			}
			selected[id] = struct{}{}
		}
	}

	selected[callerID] = struct{}{}

	session := &GroupCallSession{
		CallID:             utils.GenerateUUID(),
		ChatID:             chatID,
		CallerID:           callerID,
		CallType:           callType,
		Participants:       selected,
		PendingRequests:    make(map[string]time.Time),
		Joined:             map[string]struct{}{callerID: {}},
		ActiveParticipants: make(map[string]struct{}),
		Muted:              make(map[string]bool),
		VideoEnabled:       map[string]bool{},
		CreatedAt:          time.Now().UTC(),
		LastActivity:       time.Now().UTC(),
		Status:             "calling",
	}

	for id := range selected {
		session.Muted[id] = false
		session.VideoEnabled[id] = true
	}
	// Cuộc gọi mới chỉ tồn tại tối đa 60s nếu không ai join, giúp dọn session rác nhanh và an toàn.
	expiry := time.Now().UTC().Add(60 * time.Second)
	session.ExpiresAt = &expiry

	h.mu.Lock()
	h.groupCalls[session.CallID] = session
	h.mu.Unlock()

	return session, nil
}

func (h *Hub) RequestJoinCall(userID, callID string) (*GroupCallSession, error) {
	h.mu.RLock()
	session, ok := h.groupCalls[callID]
	h.mu.RUnlock()
	if !ok {
		return nil, errors.New("cuộc gọi không tồn tại")
	}

	if _, invited := session.Participants[userID]; invited {
		h.mu.Lock()
		if session.Joined == nil {
			session.Joined = make(map[string]struct{})
		}
		session.Joined[userID] = struct{}{}
		h.touchCallSessionLocked(session)
		// Khi đã có ít nhất 2 người join, call không còn được xem là stale theo mốc 60 giây.
		if len(session.Joined) > 1 {
			session.ExpiresAt = nil
		}
		session.UpdateActivity()
		h.mu.Unlock()
		return session, nil
	}

	h.mu.Lock()
	if session.PendingRequests == nil {
		session.PendingRequests = make(map[string]time.Time)
	}
	session.PendingRequests[userID] = time.Now().UTC()
	session.UpdateActivity()
	h.mu.Unlock()
	return session, nil
}

func (h *Hub) ApproveJoinCall(creatorID, userID, callID string) (*GroupCallSession, error) {
	h.mu.RLock()
	session, ok := h.groupCalls[callID]
	h.mu.RUnlock()
	if !ok {
		return nil, errors.New("cuộc gọi không tồn tại")
	}
	if session.CallerID != creatorID {
		return nil, errors.New("chỉ người tạo cuộc gọi mới được duyệt")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := session.PendingRequests[userID]; !ok {
		return nil, errors.New("người này không có yêu cầu tham gia")
	}

	delete(session.PendingRequests, userID)
	session.Participants[userID] = struct{}{}
	session.Joined[userID] = struct{}{}
	h.touchCallSessionLocked(session)
	if len(session.Joined) > 1 {
		session.ExpiresAt = nil
	}
	session.UpdateActivity()

	return session, nil
}

func (h *Hub) RejectJoinCall(creatorID, userID, callID string) (*GroupCallSession, error) {
	h.mu.RLock()
	session, ok := h.groupCalls[callID]
	h.mu.RUnlock()
	if !ok {
		return nil, errors.New("cuộc gọi không tồn tại")
	}
	if session.CallerID != creatorID {
		return nil, errors.New("chỉ người tạo cuộc gọi mới được từ chối")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	delete(session.PendingRequests, userID)
	return session, nil
}

// touchCallSessionLocked cập nhật hoạt động của session khi đã giữ mutex của hub.
func (h *Hub) touchCallSessionLocked(session *GroupCallSession) {
	if session == nil {
		return
	}
	session.LastActivity = time.Now().UTC()
}

func (h *Hub) EndCallByUser(enderID, callID string) (bool, *GroupCallSession, error) {
	h.mu.RLock()
	session, ok := h.groupCalls[callID]
	h.mu.RUnlock()
	if !ok {
		return false, nil, errors.New("cuộc gọi không tồn tại")
	}

	// Nếu creator kết thúc thì đóng toàn bộ session ngay lập tức.
	if session.CallerID == enderID {
		h.mu.Lock()
		delete(h.groupCalls, callID)
		h.mu.Unlock()
		return true, session, nil
	}

	// Người dùng thường chỉ rời khỏi session của họ.
	h.mu.Lock()
	if session.Joined != nil {
		delete(session.Joined, enderID)
	}
	if session.ActiveParticipants != nil {
		delete(session.ActiveParticipants, enderID)
	}
	ended := len(session.Joined) == 0
	if ended {
		delete(h.groupCalls, callID)
	}
	h.mu.Unlock()

	return ended, session, nil
}

// MarkParticipantActive đánh dấu user đã kết nối media thành công trong call.
func (h *Hub) MarkParticipantActive(callID, userID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	session, ok := h.groupCalls[callID]
	if !ok {
		return errors.New("cuộc gọi không tồn tại")
	}
	if !session.IsParticipant(userID) {
		return errors.New("người dùng không thuộc cuộc gọi này")
	}
	if session.ActiveParticipants == nil {
		session.ActiveParticipants = make(map[string]struct{})
	}
	session.ActiveParticipants[userID] = struct{}{}
	session.UpdateActivity()
	return nil
}

// ListCallsByUser trả về snapshot các session mà user đang còn tham gia để cleanup an toàn khi mất kết nối.
func (h *Hub) ListCallsByUser(userID string) []*GroupCallSession {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]*GroupCallSession, 0)
	for _, session := range h.groupCalls {
		if session == nil {
			continue
		}
		if _, ok := session.Joined[userID]; ok {
			result = append(result, session)
		}
	}
	return result
}

// CleanupStaleCalls dọn session rác định kỳ: call chưa có ai join sau 60s và request chờ quá 30s.
func (h *Hub) CleanupStaleCalls() {
	now := time.Now().UTC()

	type pendingTimeout struct {
		callID string
		userID string
	}

	type expiredCall struct {
		session *GroupCallSession
	}

	h.mu.Lock()
	pending := make([]pendingTimeout, 0)
	expired := make([]expiredCall, 0)
	for callID, session := range h.groupCalls {
		if session == nil {
			continue
		}

		if session.ExpiresAt != nil && now.After(*session.ExpiresAt) {
			delete(h.groupCalls, callID)
			expired = append(expired, expiredCall{session: session})
			continue
		}

		for userID, requestedAt := range session.PendingRequests {
			if now.Sub(requestedAt) <= 30*time.Second {
				continue
			}
			delete(session.PendingRequests, userID)
			pending = append(pending, pendingTimeout{callID: callID, userID: userID})
		}
	}
	h.mu.Unlock()

	for _, item := range pending {
		h.SendToUser(item.userID, dto.WsEvent{
			Type: "group:call:join-rejected",
			Payload: mustMarshal(map[string]any{
				"call_id": item.callID,
				"user_id": item.userID,
				"reason":  "timeout",
			}),
		})
	}

	for _, item := range expired {
		session := item.session
		if session == nil {
			continue
		}

		event := dto.WsEvent{
			Type: "group:call:ended",
			Payload: mustMarshal(map[string]any{
				"call_id": session.CallID,
				"reason":  "timeout",
			}),
		}
		h.SendToUsers(session.JoinedIDs(), event)
		if h.groupChatHub != nil {
			h.groupChatHub.Broadcast(session.ChatID, event)
		}
	}
}

func (h *Hub) GetGroupCall(callID string) (*GroupCallSession, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	session, ok := h.groupCalls[callID]
	if !ok {
		return nil, errors.New("cuộc gọi không tồn tại")
	}
	return session, nil
}

func (h *Hub) SendToUser(userID string, msg any) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	clients := h.clients[userID]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.send <- data:
		default:
			h.mu.Lock()
			delete(clients, client)
			close(client.send)
			if len(clients) == 0 {
				delete(h.clients, userID)
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) SendToUsers(userIDs []string, msg any) {
	for _, uid := range userIDs {
		h.SendToUser(uid, msg)
	}
}

func (s *GroupCallSession) IsParticipant(userID string) bool {
	_, ok := s.Participants[userID]
	return ok
}

func (s *GroupCallSession) IsJoined(userID string) bool {
	_, ok := s.Joined[userID]
	return ok
}

func (s *GroupCallSession) UpdateActivity() {
	now := time.Now().UTC()
	s.LastActivity = now
}

func (s *GroupCallSession) JoinedIDs() []string {
	ids := make([]string, 0, len(s.Joined))
	for id := range s.Joined {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
