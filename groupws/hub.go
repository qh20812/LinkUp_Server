package groupws

import (
	"encoding/json"
	"errors"
	"fmt"
	"linkup/utils"
	"sort"
	"sync"
	"time"
)

type BroadcastMessage struct {
	ChatID string
	Data   []byte
}

type Hub struct {
	rooms      map[string]map[*Client]bool
	clients    map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	groupCalls map[string]*GroupCallSession
	mu         sync.RWMutex
}

type GroupCallSession struct {
	CallID       string
	ChatID       string
	CallerID     string
	CallType     string
	Participants map[string]struct{}
	Joined       map[string]struct{}
	CreatedAt    time.Time
	Status       string
}

func (s *GroupCallSession) ParticipantIDs() []string {
	ids := make([]string, 0, len(s.Participants))
	for id := range s.Participants {
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

func (h *Hub) Run() {
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
		CallID:       utils.GenerateUUID(),
		ChatID:       chatID,
		CallerID:     callerID,
		CallType:     callType,
		Participants: selected,
		Joined:       map[string]struct{}{callerID: {}},
		CreatedAt:    time.Now().UTC(),
		Status:       "calling",
	}

	h.mu.Lock()
	h.groupCalls[session.CallID] = session
	h.mu.Unlock()

	return session, nil
}

func (h *Hub) JoinGroupCall(userID, callID string) (*GroupCallSession, error) {
	h.mu.RLock()
	session, ok := h.groupCalls[callID]
	h.mu.RUnlock()
	if !ok {
		return nil, errors.New("cuộc gọi không tồn tại")
	}

	if _, ok := session.Participants[userID]; !ok {
		return nil, errors.New("bạn không phải thành viên của cuộc gọi")
	}

	h.mu.Lock()
	if session.Joined == nil {
		session.Joined = make(map[string]struct{})
	}
	session.Joined[userID] = struct{}{}
	h.mu.Unlock()

	return session, nil
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
