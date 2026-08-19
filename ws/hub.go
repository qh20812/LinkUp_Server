package ws

import (
	"encoding/json"
	"sync"
)

// PresenceService defines the interface for presence operations.
type PresenceService interface {
	MarkOnline(userID string)
	MarkOffline(userID string)
}

type BroadcastMessage struct {
	ChatID string
	Data   []byte
}

type OutgoingMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Hub struct {
	rooms              map[string]map[*Client]bool
	clients            map[string]map[*Client]bool
	register           chan *Client
	unregister         chan *Client
	broadcast          chan *BroadcastMessage
	onClientDisconnect func(userID string)
	mu                 sync.RWMutex
	presence           PresenceService
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
	}
}

// SetPresenceService sets the presence service for the hub.
func (h *Hub) SetPresenceService(presence PresenceService) {
	h.presence = presence
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms["__global__"]; !ok {
				h.rooms["__global__"] = make(map[*Client]bool)
			}
			h.rooms["__global__"][client] = true
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mu.Unlock()

			// Mark user online in presence service
			if h.presence != nil {
				h.presence.MarkOnline(client.userID)
				h.broadcastPresenceUpdate(client.userID, "online")
			}

		case client := <-h.unregister:
			h.mu.Lock()
			h.removeFromClientsMap(client)
			h.mu.Unlock()
			if h.onClientDisconnect != nil {
				h.onClientDisconnect(client.userID)
			}
			close(client.send)

			// Check if user has any remaining connections
			if h.presence != nil {
				h.mu.RLock()
				hasConnections := len(h.clients[client.userID]) > 0
				h.mu.RUnlock()
				if !hasConnections {
					h.presence.MarkOffline(client.userID)
					h.broadcastPresenceUpdate(client.userID, "offline")
				}
			}

		case message := <-h.broadcast:
			// Phase 1 fix: Acquire RLock, snapshot the client set into a
			// slice, then release RLock before doing any I/O or mutation.
			// This avoids the RLock→Lock downgrade that caused concurrent
			// map panics when another goroutine mutated the map between
			// RUnlock and Lock.
			h.mu.RLock()
			roomClients := h.rooms[message.ChatID]
			snapshot := make([]*Client, 0, len(roomClients))
			for c := range roomClients {
				snapshot = append(snapshot, c)
			}
			h.mu.RUnlock()

			// Iterate the snapshot (owned by this goroutine, safe without lock).
			// Collect clients whose send channel is full — they are stuck
			// and must be removed to prevent goroutine leaks.
			var toRemove []*Client
			for _, c := range snapshot {
				select {
				case c.send <- message.Data:
				default:
					toRemove = append(toRemove, c)
				}
			}

			// Batch-remove stuck clients under a single write lock.
			var disconnected []string
			if len(toRemove) > 0 {
				h.mu.Lock()
				for _, c := range toRemove {
					// Double-check the client is still in the room map
					// (it may have been removed by unregister already).
					if roomClients[c] {
						delete(roomClients, c)
					}
					h.removeFromClientsMap(c)
					disconnected = append(disconnected, c.userID)
					close(c.send)
				}
				h.mu.Unlock()
			}

			// Notify after releasing the lock so hooks can safely re-enter the hub.
			if h.onClientDisconnect != nil {
				for _, uid := range disconnected {
					h.onClientDisconnect(uid)
				}
			}
		}
	}
}

func (h *Hub) JoinChat(chatID string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.rooms[chatID]; !ok {
		h.rooms[chatID] = make(map[*Client]bool)
	}
	h.rooms[chatID][client] = true
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// SetOnClientDisconnect registers a callback invoked after a client's
// connection is removed from the hub. It is only called once the user's
// sockets have actually been removed, so callers may re-check presence
// (e.g. IsUserOnline) to decide whether the user went fully offline.
func (h *Hub) SetOnClientDisconnect(fn func(userID string)) {
	h.onClientDisconnect = fn
}

// SendToUser sends a message to all WebSocket connections of the given user.
// Thread-safe: copies the client set under RLock, then mutates under Lock.
func (h *Hub) SendToUser(userID string, msg OutgoingMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	clients, ok := h.clients[userID]
	if !ok {
		h.mu.RUnlock()
		return
	}
	snapshot := make([]*Client, 0, len(clients))
	for c := range clients {
		snapshot = append(snapshot, c)
	}
	h.mu.RUnlock()

	// Iterate the snapshot (safe without lock — owned by this goroutine).
	// Collect clients whose send channel is full (stuck/disconnected).
	var toRemove []*Client
	for _, c := range snapshot {
		select {
		case c.send <- data:
		default:
			toRemove = append(toRemove, c)
		}
	}

	// Batch-remove stuck clients under a single write lock.
	if len(toRemove) > 0 {
		h.mu.Lock()
		for _, c := range toRemove {
			// Re-check: client may have already been removed by unregister.
			if clients[c] {
				delete(clients, c)
			}
			h.removeFromClientsMap(c)
			close(c.send)
		}
		// Clean up empty user entry.
		if len(clients) == 0 {
			delete(h.clients, userID)
		}
		h.mu.Unlock()
	}
}

func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID]) > 0
}

// removeFromClientsMap removes a client from all rooms and the per-userID
// clients map. Caller MUST hold h.mu (write lock).
func (h *Hub) removeFromClientsMap(c *Client) {
	for roomID, clients := range h.rooms {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.rooms, roomID)
		}
	}
	if clients, ok := h.clients[c.userID]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.clients, c.userID)
		}
	}
}

func (h *Hub) SendToUsers(userIDs []string, msg OutgoingMessage) {
	for _, uid := range userIDs {
		h.SendToUser(uid, msg)
	}
}

// broadcastPresenceUpdate broadcasts a presence update to all connected clients.
func (h *Hub) broadcastPresenceUpdate(userID string, status string) {
	msg := OutgoingMessage{
		Type: "presence:update",
		Data: map[string]interface{}{
			"user_id": userID,
			"status":  status,
		},
	}
	h.SendToAll(msg)
}

// SendToAll sends a message to all connected clients.
func (h *Hub) SendToAll(msg OutgoingMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	// Get all unique clients
	allClients := make(map[*Client]bool)
	for _, clients := range h.clients {
		for c := range clients {
			allClients[c] = true
		}
	}
	h.mu.RUnlock()

	// Send to all clients
	var toRemove []*Client
	for c := range allClients {
		select {
		case c.send <- data:
		default:
			toRemove = append(toRemove, c)
		}
	}

	// Remove stuck clients
	if len(toRemove) > 0 {
		h.mu.Lock()
		for _, c := range toRemove {
			h.removeFromClientsMap(c)
			close(c.send)
		}
		h.mu.Unlock()
	}
}

// GetUserClientCount returns the number of connections for a user.
func (h *Hub) GetUserClientCount(userID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID])
}
